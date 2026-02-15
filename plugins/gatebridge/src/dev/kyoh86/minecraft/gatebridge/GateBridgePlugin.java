package dev.kyoh86.minecraft.gatebridge;

import java.io.ByteArrayOutputStream;
import java.io.DataOutputStream;
import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import java.util.Objects;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import org.bukkit.Location;
import org.bukkit.Material;
import org.bukkit.World;
import org.bukkit.block.Block;
import org.bukkit.configuration.ConfigurationSection;
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
  private final Map<GateKey, GateRoute> gates = new HashMap<>();

  private long cooldownMs;
  private long joinGraceMs;

  @Override
  public void onEnable() {
    saveDefaultConfig();
    loadSettings();
    getServer().getMessenger().registerOutgoingPluginChannel(this, "BungeeCord");
    getServer().getPluginManager().registerEvents(this, this);
    getLogger().info("GateBridge enabled: gates=" + gates.size());
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

  private void trySwitch(Player player, Block block) {
    GateRoute route = findRoute(block);
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

  private GateRoute findRoute(Block block) {
    GateKey key = new GateKey(block.getWorld().getName(), block.getX(), block.getY(), block.getZ());
    GateRoute route = gates.get(key);
    if (route == null) {
      return null;
    }
    return block.getType() == route.triggerMaterial() ? route : null;
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

    gates.clear();
    ConfigurationSection gatesSection = getConfig().getConfigurationSection("gates");
    if (gatesSection == null) {
      throw new IllegalArgumentException("missing section: gates");
    }
    for (String gateId : gatesSection.getKeys(false)) {
      ConfigurationSection routeSection = gatesSection.getConfigurationSection(gateId);
      if (routeSection == null) {
        continue;
      }
      String world = requireString(routeSection, "world");
      int x = routeSection.getInt("x");
      int y = routeSection.getInt("y");
      int z = routeSection.getInt("z");

      String triggerBlockName = requireString(routeSection, "trigger_block");
      Material triggerMaterial = Material.matchMaterial(triggerBlockName);
      if (triggerMaterial == null) {
        throw new IllegalArgumentException("invalid trigger_block: " + triggerBlockName);
      }

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

      GateKey key = new GateKey(world, x, y, z);
      GateRoute route =
          new GateRoute(
              triggerMaterial,
              destinationServer,
              returnWorld,
              returnX,
              returnY,
              returnZ,
              returnYaw,
              returnPitch);
      gates.put(key, route);
    }
  }

  private static String requireString(ConfigurationSection section, String fieldName) {
    String value = section.getString(fieldName);
    if (value != null && !value.isBlank()) {
      return value;
    }
    throw new IllegalArgumentException("missing or invalid field: " + section.getCurrentPath() + "." + fieldName);
  }

  private record GateKey(String world, int x, int y, int z) {
    private GateKey {
      Objects.requireNonNull(world, "world");
    }
  }

  private record GateRoute(
      Material triggerMaterial,
      String destinationServer,
      String returnWorldName,
      double returnX,
      double returnY,
      double returnZ,
      float returnYaw,
      float returnPitch) {}
}
