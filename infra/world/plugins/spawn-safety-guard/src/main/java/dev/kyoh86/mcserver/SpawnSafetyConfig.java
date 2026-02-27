package dev.kyoh86.mcserver;

import org.bukkit.plugin.java.JavaPlugin;

final class SpawnSafetyConfig {
  final boolean loginSafetyEnabled;
  final String loginSafetyMainhallWorld;
  final String loginSafetyMessage;
  final int spawnSafeMinX;
  final int spawnSafeMinY;
  final int spawnSafeMinZ;
  final int spawnSafeMaxX;
  final int spawnSafeMaxY;
  final int spawnSafeMaxZ;

  private SpawnSafetyConfig(
    boolean loginSafetyEnabled,
    String loginSafetyMainhallWorld,
    String loginSafetyMessage,
    int spawnSafeMinX,
    int spawnSafeMinY,
    int spawnSafeMinZ,
    int spawnSafeMaxX,
    int spawnSafeMaxY,
    int spawnSafeMaxZ
  ) {
    this.loginSafetyEnabled = loginSafetyEnabled;
    this.loginSafetyMainhallWorld = loginSafetyMainhallWorld;
    this.loginSafetyMessage = loginSafetyMessage;
    this.spawnSafeMinX = spawnSafeMinX;
    this.spawnSafeMinY = spawnSafeMinY;
    this.spawnSafeMinZ = spawnSafeMinZ;
    this.spawnSafeMaxX = spawnSafeMaxX;
    this.spawnSafeMaxY = spawnSafeMaxY;
    this.spawnSafeMaxZ = spawnSafeMaxZ;
  }

  static SpawnSafetyConfig load(JavaPlugin plugin) {
    return new SpawnSafetyConfig(
      plugin.getConfig().getBoolean("login_safety.enabled", true),
      plugin.getConfig().getString("login_safety.mainhall_world", "mainhall"),
      plugin.getConfig().getString("login_safety.teleport_message", "§e安全な地点へ移動しました"),
      plugin.getConfig().getInt("spawn_safe.min.x", -64),
      plugin.getConfig().getInt("spawn_safe.min.y", -80),
      plugin.getConfig().getInt("spawn_safe.min.z", -64),
      plugin.getConfig().getInt("spawn_safe.max.x", 64),
      plugin.getConfig().getInt("spawn_safe.max.y", 0),
      plugin.getConfig().getInt("spawn_safe.max.z", 64)
    );
  }
}
