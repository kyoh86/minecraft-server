package dev.kyoh86.mcserver;

import java.util.HashSet;
import java.util.Set;
import org.bukkit.boss.BarColor;
import org.bukkit.plugin.java.JavaPlugin;

final class RegionStatusUIConfig {
  final Set<String> allowedRegionIds;
  final boolean statusBossbarEnabled;
  final String spawnProtectedRegionId;
  final String statusSpawnText;
  final String statusClickMobsText;
  final BarColor statusSpawnColor;
  final BarColor statusClickMobsColor;

  private RegionStatusUIConfig(
    Set<String> allowedRegionIds,
    boolean statusBossbarEnabled,
    String spawnProtectedRegionId,
    String statusSpawnText,
    String statusClickMobsText,
    BarColor statusSpawnColor,
    BarColor statusClickMobsColor
  ) {
    this.allowedRegionIds = allowedRegionIds;
    this.statusBossbarEnabled = statusBossbarEnabled;
    this.spawnProtectedRegionId = spawnProtectedRegionId;
    this.statusSpawnText = statusSpawnText;
    this.statusClickMobsText = statusClickMobsText;
    this.statusSpawnColor = statusSpawnColor;
    this.statusClickMobsColor = statusClickMobsColor;
  }

  static RegionStatusUIConfig load(JavaPlugin plugin) {
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

    return new RegionStatusUIConfig(
      allowedRegionIds,
      plugin.getConfig().getBoolean("status_bossbar.enabled", true),
      normalizeRegionId(plugin.getConfig().getString("status_bossbar.spawn_protected_region_id", "spawn_protected")),
      plugin.getConfig().getString("status_bossbar.spawn_protected_text", "保護エリア（建築・破壊不可）"),
      plugin.getConfig().getString("status_bossbar.clickmobs_allowed_text", "ClickMobs許可エリア"),
      parseColor(plugin.getConfig().getString("status_bossbar.spawn_protected_color", "RED"), BarColor.RED),
      parseColor(plugin.getConfig().getString("status_bossbar.clickmobs_allowed_color", "GREEN"), BarColor.GREEN)
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
