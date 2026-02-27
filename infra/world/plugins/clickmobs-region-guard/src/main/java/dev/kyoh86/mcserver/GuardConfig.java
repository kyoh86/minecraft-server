package dev.kyoh86.mcserver;

import java.util.HashSet;
import java.util.Set;
import org.bukkit.boss.BarColor;
import org.bukkit.plugin.java.JavaPlugin;

final class GuardConfig {
  final Set<String> allowedRegionIds;
  final boolean statusBossbarEnabled;
  final String spawnProtectedRegionId;
  final String statusSpawnText;
  final String statusClickMobsText;
  final BarColor statusSpawnColor;
  final BarColor statusClickMobsColor;
  final boolean loginSafetyEnabled;
  final String loginSafetyMainhallWorld;
  final String loginSafetyMessage;
  final int spawnSafeMinX;
  final int spawnSafeMinY;
  final int spawnSafeMinZ;
  final int spawnSafeMaxX;
  final int spawnSafeMaxY;
  final int spawnSafeMaxZ;

  private GuardConfig(
    Set<String> allowedRegionIds,
    boolean statusBossbarEnabled,
    String spawnProtectedRegionId,
    String statusSpawnText,
    String statusClickMobsText,
    BarColor statusSpawnColor,
    BarColor statusClickMobsColor,
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
    this.allowedRegionIds = allowedRegionIds;
    this.statusBossbarEnabled = statusBossbarEnabled;
    this.spawnProtectedRegionId = spawnProtectedRegionId;
    this.statusSpawnText = statusSpawnText;
    this.statusClickMobsText = statusClickMobsText;
    this.statusSpawnColor = statusSpawnColor;
    this.statusClickMobsColor = statusClickMobsColor;
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

  static GuardConfig load(JavaPlugin plugin) {
    Set<String> allowedRegionIds = new HashSet<>();
    for (String regionID : plugin.getConfig().getStringList("allowed_region_ids")) {
      String normalized = normalizeRegionId(regionID);
      if (!normalized.isEmpty()) {
        allowedRegionIds.add(normalized);
      }
    }
    if (allowedRegionIds.isEmpty()) {
      allowedRegionIds.add("clickmobs_allowed");
    }

    return new GuardConfig(
      allowedRegionIds,
      plugin.getConfig().getBoolean("status_bossbar.enabled", true),
      normalizeRegionId(plugin.getConfig().getString("status_bossbar.spawn_protected_region_id", "spawn_protected")),
      plugin.getConfig().getString("status_bossbar.spawn_protected_text", "保護エリア（建築・破壊不可）"),
      plugin.getConfig().getString("status_bossbar.clickmobs_allowed_text", "ClickMobs許可エリア"),
      parseColor(plugin.getConfig().getString("status_bossbar.spawn_protected_color", "RED"), BarColor.RED),
      parseColor(plugin.getConfig().getString("status_bossbar.clickmobs_allowed_color", "GREEN"), BarColor.GREEN),
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

  private static String normalizeRegionId(String regionID) {
    if (regionID == null) {
      return "";
    }
    return regionID.trim().toLowerCase();
  }

  private static BarColor parseColor(String value, BarColor fallback) {
    if (value == null || value.isBlank()) {
      return fallback;
    }
    try {
      return BarColor.valueOf(value.trim().toUpperCase());
    } catch (IllegalArgumentException ignored) {
      return fallback;
    }
  }
}
