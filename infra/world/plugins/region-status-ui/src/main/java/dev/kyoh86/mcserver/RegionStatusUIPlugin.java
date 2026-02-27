package dev.kyoh86.mcserver;

import java.nio.file.Files;
import java.nio.file.Path;
import org.bukkit.entity.Player;
import org.bukkit.event.EventHandler;
import org.bukkit.event.Listener;
import org.bukkit.event.player.PlayerChangedWorldEvent;
import org.bukkit.event.player.PlayerJoinEvent;
import org.bukkit.event.player.PlayerMoveEvent;
import org.bukkit.event.player.PlayerQuitEvent;
import org.bukkit.plugin.Plugin;
import org.bukkit.plugin.java.JavaPlugin;

public class RegionStatusUIPlugin extends JavaPlugin implements Listener {
  private RegionStatusUIConfig config;
  private RegionAccessService regionAccessService;
  private RegionStatusBarService statusBarService;

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
    config = RegionStatusUIConfig.load(this);
    regionAccessService = new RegionAccessService(config.allowedRegionIds);
    statusBarService = new RegionStatusBarService(config);

    getServer().getPluginManager().registerEvents(this, this);
    for (Player player : getServer().getOnlinePlayers()) {
      statusBarService.updateStatusDisplay(player, true, regionAccessService);
    }
  }

  @Override
  public void onDisable() {
    statusBarService.clear(getServer().getOnlinePlayers());
  }

  @EventHandler
  public void onJoin(PlayerJoinEvent event) {
    statusBarService.updateStatusDisplay(event.getPlayer(), true, regionAccessService);
  }

  @EventHandler
  public void onQuit(PlayerQuitEvent event) {
    statusBarService.removeStatusBar(event.getPlayer());
  }

  @EventHandler
  public void onChangedWorld(PlayerChangedWorldEvent event) {
    statusBarService.updateStatusDisplay(event.getPlayer(), true, regionAccessService);
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
    statusBarService.updateStatusDisplay(event.getPlayer(), false, regionAccessService);
  }
}
