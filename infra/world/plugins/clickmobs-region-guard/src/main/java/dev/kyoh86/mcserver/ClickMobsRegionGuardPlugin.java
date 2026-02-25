package dev.kyoh86.mcserver;

import com.sk89q.worldedit.bukkit.BukkitAdapter;
import com.sk89q.worldguard.WorldGuard;
import com.sk89q.worldguard.protection.managers.RegionManager;
import com.sk89q.worldguard.protection.regions.ProtectedRegion;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.UUID;
import org.bukkit.Bukkit;
import org.bukkit.NamespacedKey;
import org.bukkit.boss.BarColor;
import org.bukkit.boss.BarStyle;
import org.bukkit.boss.BossBar;
import org.bukkit.entity.Boat;
import org.bukkit.entity.HumanEntity;
import org.bukkit.entity.LivingEntity;
import org.bukkit.entity.Minecart;
import org.bukkit.entity.Player;
import org.bukkit.event.EventHandler;
import org.bukkit.event.EventPriority;
import org.bukkit.event.Listener;
import org.bukkit.event.block.BlockPlaceEvent;
import org.bukkit.event.player.PlayerChangedWorldEvent;
import org.bukkit.event.player.PlayerInteractEntityEvent;
import org.bukkit.event.player.PlayerJoinEvent;
import org.bukkit.event.player.PlayerMoveEvent;
import org.bukkit.event.player.PlayerQuitEvent;
import org.bukkit.inventory.EquipmentSlot;
import org.bukkit.inventory.ItemStack;
import org.bukkit.inventory.meta.ItemMeta;
import org.bukkit.permissions.PermissionAttachment;
import org.bukkit.persistence.PersistentDataContainer;
import org.bukkit.persistence.PersistentDataType;
import org.bukkit.plugin.Plugin;
import org.bukkit.plugin.java.JavaPlugin;

public class ClickMobsRegionGuardPlugin extends JavaPlugin implements Listener {
  private static final String BYPASS_PERMISSION = "clickmobsregionguard.bypass";
  private static final String CLICKMOBS_PICKUP = "clickmobs.pickup";
  private static final String CLICKMOBS_PLACE = "clickmobs.place";
  private static final NamespacedKey CLICKMOBS_ENTITY_KEY = new NamespacedKey("clickmobs", "entity");

  private final Map<String, Set<String>> allowedRegionsByWorld = new HashMap<>();
  private final Map<UUID, PermissionAttachment> attachments = new HashMap<>();
  private final Map<UUID, BossBar> statusBars = new HashMap<>();
  private final Map<UUID, RegionStatus> lastRegionStatus = new HashMap<>();

  private boolean statusBossbarEnabled;
  private String statusSpawnText;
  private String statusClickMobsText;
  private BarColor statusSpawnColor;
  private BarColor statusClickMobsColor;
  private String spawnProtectedRegionId;

  private enum RegionStatus {
    NONE,
    CLICKMOBS_ALLOWED,
    SPAWN_PROTECTED
  }

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
    loadStatusConfig();

    getServer().getPluginManager().registerEvents(this, this);
    for (Player player : getServer().getOnlinePlayers()) {
      ensureClickMobsPermission(player);
      updateStatusDisplay(player, true);
    }
  }

  @Override
  public void onDisable() {
    for (Player player : getServer().getOnlinePlayers()) {
      removeAttachment(player);
      removeStatusBar(player);
    }
    attachments.clear();
    statusBars.clear();
    lastRegionStatus.clear();
  }

  @EventHandler
  public void onJoin(PlayerJoinEvent event) {
    ensureClickMobsPermission(event.getPlayer());
    updateStatusDisplay(event.getPlayer(), true);
  }

  @EventHandler
  public void onQuit(PlayerQuitEvent event) {
    removeAttachment(event.getPlayer());
    removeStatusBar(event.getPlayer());
  }

  @EventHandler
  public void onChangedWorld(PlayerChangedWorldEvent event) {
    updateStatusDisplay(event.getPlayer(), true);
  }

  @EventHandler(ignoreCancelled = true)
  public void onMove(PlayerMoveEvent event) {
    if (event.getTo() == null) {
      return;
    }
    if (event.getFrom().getBlockX() == event.getTo().getBlockX()
      && event.getFrom().getBlockY() == event.getTo().getBlockY()
      && event.getFrom().getBlockZ() == event.getTo().getBlockZ()) {
      return;
    }
    updateStatusDisplay(event.getPlayer(), false);
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
        String normalized = normalizeRegionId(regionID);
        if (!normalized.isEmpty()) {
          ids.add(normalized);
        }
      }
      allowedRegionsByWorld.put(world, ids);
    }
  }

  private void loadStatusConfig() {
    statusBossbarEnabled = getConfig().getBoolean("status_bossbar.enabled", true);
    spawnProtectedRegionId = normalizeRegionId(getConfig().getString("status_bossbar.spawn_protected_region_id", "spawn_protected"));
    statusSpawnText = getConfig().getString("status_bossbar.spawn_protected_text", "保護エリア（建築・破壊不可）");
    statusClickMobsText = getConfig().getString("status_bossbar.clickmobs_allowed_text", "ClickMobs許可エリア");
    statusSpawnColor = parseColor(getConfig().getString("status_bossbar.spawn_protected_color", "RED"), BarColor.RED);
    statusClickMobsColor = parseColor(getConfig().getString("status_bossbar.clickmobs_allowed_color", "GREEN"), BarColor.GREEN);
  }

  private BarColor parseColor(String value, BarColor fallback) {
    if (value == null || value.isBlank()) {
      return fallback;
    }
    try {
      return BarColor.valueOf(value.trim().toUpperCase());
    } catch (IllegalArgumentException ignored) {
      return fallback;
    }
  }

  private void updateStatusDisplay(Player player, boolean force) {
    RegionStatus next = detectRegionStatus(player);
    RegionStatus previous = lastRegionStatus.getOrDefault(player.getUniqueId(), RegionStatus.NONE);
    if (!force && previous == next) {
      return;
    }
    lastRegionStatus.put(player.getUniqueId(), next);

    if (!statusBossbarEnabled || next == RegionStatus.NONE) {
      hideStatusBar(player);
      return;
    }

    BossBar bar = statusBars.computeIfAbsent(player.getUniqueId(), uuid -> {
      BossBar created = Bukkit.createBossBar("", BarColor.WHITE, BarStyle.SOLID);
      created.setProgress(1.0);
      created.addPlayer(player);
      return created;
    });

    if (!bar.getPlayers().contains(player)) {
      bar.addPlayer(player);
    }

    if (next == RegionStatus.SPAWN_PROTECTED) {
      bar.setTitle(statusSpawnText);
      bar.setColor(statusSpawnColor);
    } else {
      bar.setTitle(statusClickMobsText);
      bar.setColor(statusClickMobsColor);
    }
    bar.setVisible(true);
  }

  private void hideStatusBar(Player player) {
    BossBar bar = statusBars.get(player.getUniqueId());
    if (bar == null) {
      return;
    }
    bar.setVisible(false);
  }

  private void removeStatusBar(Player player) {
    BossBar bar = statusBars.remove(player.getUniqueId());
    lastRegionStatus.remove(player.getUniqueId());
    if (bar == null) {
      return;
    }
    bar.removePlayer(player);
    bar.setVisible(false);
  }

  private RegionStatus detectRegionStatus(Player player) {
    if (isPlayerInRegion(player, spawnProtectedRegionId)) {
      return RegionStatus.SPAWN_PROTECTED;
    }
    if (isClickMobsAllowed(player)) {
      return RegionStatus.CLICKMOBS_ALLOWED;
    }
    return RegionStatus.NONE;
  }

  private boolean isPlayerInRegion(Player player, String regionId) {
    if (regionId.isEmpty()) {
      return false;
    }
    RegionManager manager = WorldGuard.getInstance().getPlatform().getRegionContainer().get(BukkitAdapter.adapt(player.getWorld()));
    if (manager == null) {
      return false;
    }
    ProtectedRegion region = manager.getRegion(regionId);
    if (region == null) {
      return false;
    }
    int x = player.getLocation().getBlockX();
    int y = player.getLocation().getBlockY();
    int z = player.getLocation().getBlockZ();
    return region.contains(x, y, z);
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
    int y = player.getLocation().getBlockY();
    int z = player.getLocation().getBlockZ();
    for (String regionID : allowedRegions) {
      ProtectedRegion region = manager.getRegion(regionID);
      if (region == null) {
        continue;
      }
      if (!region.contains(x, y, z)) {
        continue;
      }
      return true;
    }
    return false;
  }

  private String normalizeRegionId(String regionID) {
    if (regionID == null) {
      return "";
    }
    return regionID.trim().toLowerCase();
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
