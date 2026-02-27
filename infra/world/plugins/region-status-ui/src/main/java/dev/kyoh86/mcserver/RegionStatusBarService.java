package dev.kyoh86.mcserver;

import java.util.Collection;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;
import org.bukkit.Bukkit;
import org.bukkit.boss.BarColor;
import org.bukkit.boss.BarStyle;
import org.bukkit.boss.BossBar;
import org.bukkit.entity.Player;

final class RegionStatusBarService {
  private final RegionStatusUIConfig config;
  private final Map<UUID, BossBar> statusBars = new HashMap<>();
  private final Map<UUID, RegionStatus> lastRegionStatus = new HashMap<>();

  RegionStatusBarService(RegionStatusUIConfig config) {
    this.config = config;
  }

  void updateStatusDisplay(Player player, boolean force, RegionAccessService regionAccessService) {
    RegionStatus next = regionAccessService.detectRegionStatus(player, config.spawnProtectedRegionId);
    RegionStatus previous = lastRegionStatus.getOrDefault(player.getUniqueId(), RegionStatus.NONE);
    if (!force && previous == next) {
      return;
    }
    lastRegionStatus.put(player.getUniqueId(), next);

    if (!config.statusBossbarEnabled || next == RegionStatus.NONE) {
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
      bar.setTitle(config.statusSpawnText);
      bar.setColor(config.statusSpawnColor);
    } else {
      bar.setTitle(config.statusClickMobsText);
      bar.setColor(config.statusClickMobsColor);
    }
    bar.setVisible(true);
  }

  void removeStatusBar(Player player) {
    BossBar bar = statusBars.remove(player.getUniqueId());
    lastRegionStatus.remove(player.getUniqueId());
    if (bar == null) {
      return;
    }
    bar.removePlayer(player);
    bar.setVisible(false);
  }

  void clear(Collection<? extends Player> players) {
    for (Player player : players) {
      removeStatusBar(player);
    }
    statusBars.clear();
    lastRegionStatus.clear();
  }

  private void hideStatusBar(Player player) {
    BossBar bar = statusBars.get(player.getUniqueId());
    if (bar == null) {
      return;
    }
    bar.setVisible(false);
  }
}
