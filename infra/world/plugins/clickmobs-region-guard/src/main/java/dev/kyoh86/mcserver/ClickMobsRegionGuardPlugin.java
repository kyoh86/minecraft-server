package dev.kyoh86.mcserver;

import java.nio.file.Files;
import java.nio.file.Path;
import org.bukkit.entity.Player;
import org.bukkit.event.EventHandler;
import org.bukkit.event.EventPriority;
import org.bukkit.event.Listener;
import org.bukkit.event.block.BlockPlaceEvent;
import org.bukkit.event.player.PlayerChangedWorldEvent;
import org.bukkit.event.player.PlayerInteractEntityEvent;
import org.bukkit.event.player.PlayerJoinEvent;
import org.bukkit.event.player.PlayerQuitEvent;
import org.bukkit.inventory.EquipmentSlot;
import org.bukkit.plugin.Plugin;
import org.bukkit.plugin.java.JavaPlugin;

public class ClickMobsRegionGuardPlugin extends JavaPlugin implements Listener {
  private static final String BYPASS_PERMISSION = "clickmobsregionguard.bypass";

  private ClickMobsGuardConfig config;
  private RegionAccessService regionAccessService;
  private ClickMobsPermissionService permissionService;

  @Override
  public void onEnable() {
    Plugin worldGuard = getServer().getPluginManager().getPlugin("WorldGuard");
    if (worldGuard == null || !worldGuard.isEnabled()) {
      getLogger().severe("WorldGuard is required; disabling plugin.");
      getServer().getPluginManager().disablePlugin(this);
      return;
    }

    Path configPath = getDataFolder().toPath().resolve("config.yml");
    if (!Files.isRegularFile(configPath)) {
      getLogger().severe("Missing config file: " + configPath);
      getServer().getPluginManager().disablePlugin(this);
      return;
    }
    reloadConfig();
    config = ClickMobsGuardConfig.load(this);
    regionAccessService = new RegionAccessService(config.allowedRegionIds);
    permissionService = new ClickMobsPermissionService(this);

    getServer().getPluginManager().registerEvents(this, this);
    for (Player player : getServer().getOnlinePlayers()) {
      permissionService.ensureClickMobsPermission(player);
    }
  }

  @Override
  public void onDisable() {
    permissionService.clear(getServer().getOnlinePlayers());
  }

  @EventHandler
  public void onJoin(PlayerJoinEvent event) {
    permissionService.ensureClickMobsPermission(event.getPlayer());
  }

  @EventHandler
  public void onQuit(PlayerQuitEvent event) {
    permissionService.removeAttachment(event.getPlayer());
  }

  @EventHandler
  public void onChangedWorld(PlayerChangedWorldEvent event) {
    permissionService.ensureClickMobsPermission(event.getPlayer());
  }

  @EventHandler(priority = EventPriority.LOWEST, ignoreCancelled = true)
  public void onPlayerInteractEntity(PlayerInteractEntityEvent event) {
    if (event.getHand() != EquipmentSlot.HAND) {
      return;
    }

    Player player = event.getPlayer();
    permissionService.ensureClickMobsPermission(player);
    if (player.hasPermission(BYPASS_PERMISSION)) {
      return;
    }
    if (regionAccessService.isClickMobsAllowed(player)) {
      return;
    }

    if (ClickMobsActionDetector.isPickupAttempt(event) || ClickMobsActionDetector.isVehiclePlaceAttempt(event)) {
      event.setCancelled(true);
      player.sendActionBar("§cこのエリアでは ClickMobs は使えません");
    }
  }

  @EventHandler(priority = EventPriority.LOWEST, ignoreCancelled = true)
  public void onBlockPlace(BlockPlaceEvent event) {
    Player player = event.getPlayer();
    permissionService.ensureClickMobsPermission(player);
    if (player.hasPermission(BYPASS_PERMISSION)) {
      return;
    }
    if (regionAccessService.isClickMobsAllowed(player)) {
      return;
    }

    if (ClickMobsActionDetector.isClickMobsItem(player.getInventory().getItemInMainHand())
      || ClickMobsActionDetector.isClickMobsItem(player.getInventory().getItemInOffHand())) {
      event.setCancelled(true);
      player.sendActionBar("§cこのエリアでは ClickMobs は使えません");
    }
  }
}
