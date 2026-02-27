package dev.kyoh86.mcserver;

import org.bukkit.event.EventHandler;
import org.bukkit.event.Listener;
import org.bukkit.event.player.PlayerChangedWorldEvent;
import org.bukkit.event.player.PlayerJoinEvent;
import org.bukkit.plugin.java.JavaPlugin;

public class SpawnSafetyGuardPlugin extends JavaPlugin implements Listener {
  private SpawnSafetyConfig config;
  private LoginSafetyService loginSafetyService;

  @Override
  public void onEnable() {
    saveDefaultConfig();
    config = SpawnSafetyConfig.load(this);
    loginSafetyService = new LoginSafetyService(this, config);

    getServer().getPluginManager().registerEvents(this, this);
    getServer().getOnlinePlayers().forEach(loginSafetyService::scheduleChecks);
  }

  @EventHandler
  public void onJoin(PlayerJoinEvent event) {
    loginSafetyService.scheduleChecks(event.getPlayer());
  }

  @EventHandler
  public void onChangedWorld(PlayerChangedWorldEvent event) {
    loginSafetyService.scheduleChecks(event.getPlayer());
  }
}
