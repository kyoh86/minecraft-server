package dev.kyoh86.mcserver;

import com.sk89q.worldedit.bukkit.BukkitAdapter;
import com.sk89q.worldguard.WorldGuard;
import com.sk89q.worldguard.protection.managers.RegionManager;
import com.sk89q.worldguard.protection.regions.ProtectedRegion;
import java.util.Set;
import org.bukkit.entity.Player;

final class RegionAccessService {
  private final Set<String> allowedRegionIds;

  RegionAccessService(Set<String> allowedRegionIds) {
    this.allowedRegionIds = allowedRegionIds;
  }

  RegionStatus detectRegionStatus(Player player, String spawnProtectedRegionId) {
    if (isPlayerInRegion(player, spawnProtectedRegionId)) {
      return RegionStatus.SPAWN_PROTECTED;
    }
    if (isClickMobsAllowed(player)) {
      return RegionStatus.CLICKMOBS_ALLOWED;
    }
    return RegionStatus.NONE;
  }

  boolean isClickMobsAllowed(Player player) {
    if (allowedRegionIds.isEmpty()) {
      return false;
    }

    RegionManager manager = WorldGuard.getInstance()
      .getPlatform()
      .getRegionContainer()
      .get(BukkitAdapter.adapt(player.getWorld()));
    if (manager == null) {
      return false;
    }

    int x = player.getLocation().getBlockX();
    int y = player.getLocation().getBlockY();
    int z = player.getLocation().getBlockZ();
    for (String regionID : allowedRegionIds) {
      ProtectedRegion region = manager.getRegion(regionID);
      if (region == null) {
        continue;
      }
      if (region.contains(x, y, z)) {
        return true;
      }
    }
    return false;
  }

  private boolean isPlayerInRegion(Player player, String regionId) {
    if (regionId == null || regionId.isEmpty()) {
      return false;
    }
    RegionManager manager = WorldGuard.getInstance()
      .getPlatform()
      .getRegionContainer()
      .get(BukkitAdapter.adapt(player.getWorld()));
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
}
