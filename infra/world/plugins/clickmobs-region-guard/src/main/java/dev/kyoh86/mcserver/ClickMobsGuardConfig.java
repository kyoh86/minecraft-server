package dev.kyoh86.mcserver;

import java.util.HashSet;
import java.util.Set;
import org.bukkit.plugin.java.JavaPlugin;

final class ClickMobsGuardConfig {
  final Set<String> allowedRegionIds;

  private ClickMobsGuardConfig(Set<String> allowedRegionIds) {
    this.allowedRegionIds = allowedRegionIds;
  }

  static ClickMobsGuardConfig load(JavaPlugin plugin) {
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
    return new ClickMobsGuardConfig(allowedRegionIds);
  }

  private static String normalizeRegionId(String regionID) {
    if (regionID == null) {
      return "";
    }
    return regionID.trim().toLowerCase();
  }
}
