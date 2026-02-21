package dev.kyoh86.mcserver;

import com.sk89q.worldedit.bukkit.BukkitAdapter;
import com.sk89q.worldguard.WorldGuard;
import com.sk89q.worldguard.protection.managers.RegionManager;
import com.sk89q.worldguard.protection.regions.ProtectedRegion;
import org.bukkit.NamespacedKey;
import org.bukkit.entity.Boat;
import org.bukkit.entity.HumanEntity;
import org.bukkit.entity.LivingEntity;
import org.bukkit.entity.Minecart;
import org.bukkit.entity.Player;
import org.bukkit.event.EventHandler;
import org.bukkit.event.EventPriority;
import org.bukkit.event.Listener;
import org.bukkit.event.block.BlockPlaceEvent;
import org.bukkit.event.player.PlayerInteractEntityEvent;
import org.bukkit.inventory.EquipmentSlot;
import org.bukkit.inventory.ItemStack;
import org.bukkit.inventory.meta.ItemMeta;
import org.bukkit.permissions.PermissionAttachment;
import org.bukkit.persistence.PersistentDataContainer;
import org.bukkit.persistence.PersistentDataType;
import org.bukkit.plugin.Plugin;
import org.bukkit.plugin.java.JavaPlugin;

import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.UUID;

public class ClickMobsRegionGuardPlugin extends JavaPlugin implements Listener {
  private static final String BYPASS_PERMISSION = "clickmobsregionguard.bypass";
  private static final String CLICKMOBS_PICKUP = "clickmobs.pickup";
  private static final String CLICKMOBS_PLACE = "clickmobs.place";
  private static final NamespacedKey CLICKMOBS_ENTITY_KEY = new NamespacedKey("clickmobs", "entity");

  private final Map<String, Set<String>> allowedRegionsByWorld = new HashMap<>();
  private final Map<UUID, PermissionAttachment> attachments = new HashMap<>();

  @Override
  public void onEnable() {
    Plugin worldGuard = getServer().getPluginManager().getPlugin("WorldGuard");
    if (worldGuard == null || !worldGuard.isEnabled()) {
      getLogger().severe("WorldGuard is required; disabling plugin.");
      getServer().getPluginManager().disablePlugin(this);
      return;
    }

    saveDefaultConfig();
    loadAllowedRegions();

    getServer().getPluginManager().registerEvents(this, this);
    for (Player player : getServer().getOnlinePlayers()) {
      ensureClickMobsPermission(player);
    }
  }

  @Override
  public void onDisable() {
    for (Player player : getServer().getOnlinePlayers()) {
      removeAttachment(player);
    }
    attachments.clear();
  }

  @EventHandler
  public void onJoin(org.bukkit.event.player.PlayerJoinEvent event) {
    ensureClickMobsPermission(event.getPlayer());
  }

  @EventHandler(priority = EventPriority.LOWEST, ignoreCancelled = true)
  public void onPlayerInteractEntity(PlayerInteractEntityEvent event) {
    if (event.getHand() != EquipmentSlot.HAND) {
      return;
    }

    Player player = event.getPlayer();
    ensureClickMobsPermission(player);
    if (player.hasPermission(BYPASS_PERMISSION)) {
      return;
    }
    if (isClickMobsAllowed(player)) {
      return;
    }

    if (isPickupAttempt(event) || isVehiclePlaceAttempt(event)) {
      event.setCancelled(true);
      player.sendActionBar("§cこのエリアでは ClickMobs は使えません");
    }
  }

  @EventHandler(priority = EventPriority.LOWEST, ignoreCancelled = true)
  public void onBlockPlace(BlockPlaceEvent event) {
    Player player = event.getPlayer();
    ensureClickMobsPermission(player);
    if (player.hasPermission(BYPASS_PERMISSION)) {
      return;
    }
    if (isClickMobsAllowed(player)) {
      return;
    }

    if (isClickMobsItem(player.getInventory().getItemInMainHand()) || isClickMobsItem(player.getInventory().getItemInOffHand())) {
      event.setCancelled(true);
      player.sendActionBar("§cこのエリアでは ClickMobs は使えません");
    }
  }

  private void loadAllowedRegions() {
    allowedRegionsByWorld.clear();
    if (!getConfig().isConfigurationSection("allowed_regions")) {
      return;
    }
    for (String world : getConfig().getConfigurationSection("allowed_regions").getKeys(false)) {
      List<String> regionIDs = getConfig().getStringList("allowed_regions." + world);
      Set<String> ids = new HashSet<>();
      for (String regionID : regionIDs) {
        String normalized = regionID == null ? "" : regionID.trim().toLowerCase();
        if (!normalized.isEmpty()) {
          ids.add(normalized);
        }
      }
      allowedRegionsByWorld.put(world, ids);
    }
  }

  private boolean isClickMobsAllowed(Player player) {
    String worldName = player.getWorld().getName();
    Set<String> allowedRegions = allowedRegionsByWorld.getOrDefault(worldName, Collections.emptySet());
    if (allowedRegions.isEmpty()) {
      return false;
    }

    RegionManager manager = WorldGuard.getInstance().getPlatform().getRegionContainer().get(BukkitAdapter.adapt(player.getWorld()));
    if (manager == null) {
      return false;
    }

    int x = player.getLocation().getBlockX();
    int z = player.getLocation().getBlockZ();
    for (String regionID : allowedRegions) {
      ProtectedRegion region = manager.getRegion(regionID);
      if (region == null) {
        continue;
      }
      if (x < region.getMinimumPoint().x() || x > region.getMaximumPoint().x()) {
        continue;
      }
      if (z < region.getMinimumPoint().z() || z > region.getMaximumPoint().z()) {
        continue;
      }
      return true;
    }
    return false;
  }

  private void ensureClickMobsPermission(Player player) {
    PermissionAttachment attachment = attachments.computeIfAbsent(player.getUniqueId(), uuid -> player.addAttachment(this));
    attachment.setPermission(CLICKMOBS_PICKUP, true);
    attachment.setPermission(CLICKMOBS_PLACE, true);
    player.recalculatePermissions();
  }

  private void removeAttachment(Player player) {
    PermissionAttachment attachment = attachments.remove(player.getUniqueId());
    if (attachment == null) {
      return;
    }
    player.removeAttachment(attachment);
    player.recalculatePermissions();
  }

  private boolean isPickupAttempt(PlayerInteractEntityEvent event) {
    if (!event.getPlayer().isSneaking()) {
      return false;
    }
    if (!(event.getRightClicked() instanceof LivingEntity)) {
      return false;
    }
    return !(event.getRightClicked() instanceof HumanEntity);
  }

  private boolean isVehiclePlaceAttempt(PlayerInteractEntityEvent event) {
    if (!event.getPlayer().isSneaking()) {
      return false;
    }
    if (!(event.getRightClicked() instanceof Boat) && !(event.getRightClicked() instanceof Minecart)) {
      return false;
    }
    return isClickMobsItem(event.getPlayer().getInventory().getItemInMainHand());
  }

  private boolean isClickMobsItem(ItemStack item) {
    if (item == null || item.getType().isAir()) {
      return false;
    }
    if (!item.getType().toString().equals("PLAYER_HEAD")) {
      return false;
    }
    ItemMeta meta = item.getItemMeta();
    if (meta == null) {
      return false;
    }
    PersistentDataContainer pdc = meta.getPersistentDataContainer();
    return pdc.has(CLICKMOBS_ENTITY_KEY, PersistentDataType.BOOLEAN);
  }
}
