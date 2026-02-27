package dev.kyoh86.mcserver;

import org.bukkit.Bukkit;
import org.bukkit.Location;
import org.bukkit.World;
import org.bukkit.entity.Player;
import org.bukkit.plugin.java.JavaPlugin;

final class LoginSafetyService {
  private final JavaPlugin plugin;
  private final GuardConfig config;

  LoginSafetyService(JavaPlugin plugin, GuardConfig config) {
    this.plugin = plugin;
    this.config = config;
  }

  void scheduleChecks(Player player) {
    // Multiverse may apply destination after join/world-change handling,
    // so probe multiple ticks to catch post-teleport drift.
    Bukkit.getScheduler().runTask(plugin, () -> enforce(player));
    Bukkit.getScheduler().runTaskLater(plugin, () -> enforce(player), 20L);
    Bukkit.getScheduler().runTaskLater(plugin, () -> enforce(player), 60L);
  }

  private void enforce(Player player) {
    if (!config.loginSafetyEnabled || !player.isOnline()) {
      return;
    }
    World world = player.getWorld();
    if (!world.getName().equals(config.loginSafetyMainhallWorld)) {
      return;
    }
    if (isWithinSpawnSafeBounds(player.getLocation())) {
      return;
    }
    Location from = player.getLocation().clone();
    Location spawn = world.getSpawnLocation().clone().add(0.5, 0.0, 0.5);
    player.teleport(spawn);
    if (config.loginSafetyMessage != null && !config.loginSafetyMessage.isBlank()) {
      player.sendActionBar(config.loginSafetyMessage);
    }
    plugin.getLogger().info(
      "login safety: teleported " + player.getName()
        + " to spawn from " + formatBlockPosition(from)
    );
  }

  private boolean isWithinSpawnSafeBounds(Location location) {
    int x = location.getBlockX();
    int y = location.getBlockY();
    int z = location.getBlockZ();
    return x >= config.spawnSafeMinX && x <= config.spawnSafeMaxX
      && y >= config.spawnSafeMinY && y <= config.spawnSafeMaxY
      && z >= config.spawnSafeMinZ && z <= config.spawnSafeMaxZ;
  }

  private String formatBlockPosition(Location location) {
    return location.getWorld().getName()
      + ":" + location.getBlockX()
      + "," + location.getBlockY()
      + "," + location.getBlockZ();
  }
}
