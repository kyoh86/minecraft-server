package dev.kyoh86.minecraft.gatebridge;

import java.io.ByteArrayOutputStream;
import java.io.DataOutputStream;
import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import org.bukkit.Location;
import org.bukkit.Material;
import org.bukkit.World;
import org.bukkit.block.Block;
import org.bukkit.configuration.ConfigurationSection;
import org.bukkit.entity.Entity;
import org.bukkit.entity.EntityType;
import org.bukkit.entity.Player;
import org.bukkit.event.EventHandler;
import org.bukkit.event.EventPriority;
import org.bukkit.event.Listener;
import org.bukkit.event.block.Action;
import org.bukkit.event.player.PlayerInteractEvent;
import org.bukkit.event.player.PlayerJoinEvent;
import org.bukkit.event.player.PlayerMoveEvent;
import org.bukkit.plugin.java.JavaPlugin;

public final class GateBridgePlugin extends JavaPlugin implements Listener {
  private final Map<UUID, Long> lastTriggeredAt = new ConcurrentHashMap<>();
  private final Map<UUID, Long> joinGraceUntil = new ConcurrentHashMap<>();
  private final Map<String, GateRoute> gatesByMarkerTag = new HashMap<>();

  private long cooldownMs;
  private long joinGraceMs;
  private Material triggerMaterial;
  private double markerSearchRadius;

  @Override
  public void onEnable() {
    saveDefaultConfig();
    loadSettings();
    getServer().getMessenger().registerOutgoingPluginChannel(this, "BungeeCord");
    getServer().getPluginManager().registerEvents(this, this);
    getLogger().info("GateBridge enabled: routes=" + gatesByMarkerTag.size());
  }

  @EventHandler(priority = EventPriority.MONITOR, ignoreCancelled = true)
  public void onPlayerMove(PlayerMoveEvent event) {
    Location to = event.getTo();
    Location from = event.getFrom();
    if (to == null) {
      return;
    }
    if (to.getBlockX() == from.getBlockX()
        && to.getBlockY() == from.getBlockY()
        && to.getBlockZ() == from.getBlockZ()) {
      return;
    }
    trySwitch(event.getPlayer(), to.getBlock());
    trySwitch(event.getPlayer(), to.clone().subtract(0, 1, 0).getBlock());
  }

  @EventHandler(priority = EventPriority.MONITOR, ignoreCancelled = true)
  public void onPlayerInteract(PlayerInteractEvent event) {
    if (event.getAction() != Action.PHYSICAL) {
      return;
    }
    Block block = event.getClickedBlock();
    if (block == null) {
      return;
    }
    trySwitch(event.getPlayer(), block);
  }

  @EventHandler(priority = EventPriority.MONITOR)
  public void onPlayerJoin(PlayerJoinEvent event) {
    joinGraceUntil.put(event.getPlayer().getUniqueId(), System.currentTimeMillis() + joinGraceMs);
  }

  private void trySwitch(Player player, Block triggerBlock) {
    GateRoute route = findRoute(triggerBlock);
    if (route == null) {
      return;
    }
    if (isCoolingDown(player)) {
      return;
    }
    World returnWorld = resolveWorld(player, route.returnWorldName());
    player.teleport(
        new Location(
            returnWorld,
            route.returnX(),
            route.returnY(),
            route.returnZ(),
            route.returnYaw(),
            route.returnPitch()));
    sendToServer(player, route.destinationServer());
  }

  private GateRoute findRoute(Block triggerBlock) {
    if (triggerBlock.getType() != triggerMaterial) {
      return null;
    }

    Location center = triggerBlock.getLocation().add(0.5d, 0.5d, 0.5d);
    Entity nearestMarker = null;
    double nearestDistanceSquared = Double.MAX_VALUE;
    GateRoute nearestRoute = null;

    for (Entity entity :
        triggerBlock
            .getWorld()
            .getNearbyEntities(
                center,
                markerSearchRadius,
                markerSearchRadius,
                markerSearchRadius,
                e -> e.getType() == EntityType.MARKER)) {
      for (String tag : entity.getScoreboardTags()) {
        GateRoute route = gatesByMarkerTag.get(tag);
        if (route == null) {
          continue;
        }
        double distanceSquared = entity.getLocation().distanceSquared(center);
        if (distanceSquared < nearestDistanceSquared) {
          nearestDistanceSquared = distanceSquared;
          nearestMarker = entity;
          nearestRoute = route;
        }
      }
    }

    if (nearestMarker == null) {
      return null;
    }
    return nearestRoute;
  }

  private World resolveWorld(Player player, String worldName) {
    if (worldName == null || worldName.isBlank()) {
      return player.getWorld();
    }
    World world = getServer().getWorld(worldName);
    if (world != null) {
      return world;
    }
    getLogger().warning("Unknown world in config: " + worldName + ", fallback to current world");
    return player.getWorld();
  }

  private boolean isCoolingDown(Player player) {
    long now = System.currentTimeMillis();
    UUID id = player.getUniqueId();
    Long joinGrace = joinGraceUntil.get(id);
    if (joinGrace != null && now < joinGrace) {
      return true;
    }
    Long last = lastTriggeredAt.get(id);
    if (last != null && now - last < cooldownMs) {
      return true;
    }
    lastTriggeredAt.put(id, now);
    return false;
  }

  private void sendToServer(Player player, String serverName) {
    try {
      ByteArrayOutputStream bytes = new ByteArrayOutputStream();
      DataOutputStream out = new DataOutputStream(bytes);
      out.writeUTF("Connect");
      out.writeUTF(serverName);
      player.sendPluginMessage(this, "BungeeCord", bytes.toByteArray());
    } catch (IOException e) {
      getLogger().warning(
          "Failed to send player '" + player.getName() + "' to " + serverName + ": " + e.getMessage());
    }
  }

  private void loadSettings() {
    reloadConfig();

    cooldownMs = getConfig().getLong("cooldown_ms", 2000L);
    joinGraceMs = getConfig().getLong("join_grace_ms", 5000L);

    String triggerBlockName = getConfig().getString("trigger_block", "POLISHED_BLACKSTONE_PRESSURE_PLATE");
    Material loadedMaterial = Material.matchMaterial(triggerBlockName);
    if (loadedMaterial == null) {
      throw new IllegalArgumentException("invalid trigger_block: " + triggerBlockName);
    }
    triggerMaterial = loadedMaterial;

    markerSearchRadius = getConfig().getDouble("marker_search_radius", 1.5d);

    gatesByMarkerTag.clear();
    ConfigurationSection gatesSection = getConfig().getConfigurationSection("gates");
    if (gatesSection == null) {
      throw new IllegalArgumentException("missing section: gates");
    }

    for (String gateId : gatesSection.getKeys(false)) {
      ConfigurationSection routeSection = gatesSection.getConfigurationSection(gateId);
      if (routeSection == null) {
        continue;
      }

      String markerTag = requireString(routeSection, "marker_tag");

      String destinationServer = requireString(routeSection, "destination_server");
      ConfigurationSection ret = routeSection.getConfigurationSection("return");
      if (ret == null) {
        throw new IllegalArgumentException("missing section: gates." + gateId + ".return");
      }
      String returnWorld = ret.getString("world", "");
      double returnX = ret.getDouble("x");
      double returnY = ret.getDouble("y");
      double returnZ = ret.getDouble("z");
      float returnYaw = (float) ret.getDouble("yaw", 0.0d);
      float returnPitch = (float) ret.getDouble("pitch", 0.0d);

      gatesByMarkerTag.put(
          markerTag,
          new GateRoute(
              destinationServer,
              returnWorld,
              returnX,
              returnY,
              returnZ,
              returnYaw,
              returnPitch));
    }
  }

  private static String requireString(ConfigurationSection section, String fieldName) {
    String value = section.getString(fieldName);
    if (value != null && !value.isBlank()) {
      return value;
    }
    throw new IllegalArgumentException("missing or invalid field: " + section.getCurrentPath() + "." + fieldName);
  }

  private record GateRoute(
      String destinationServer,
      String returnWorldName,
      double returnX,
      double returnY,
      double returnZ,
      float returnYaw,
      float returnPitch) {}
}
