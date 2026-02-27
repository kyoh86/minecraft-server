package dev.kyoh86.mcserver;

import java.util.Collection;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;
import org.bukkit.entity.Player;
import org.bukkit.permissions.PermissionAttachment;
import org.bukkit.plugin.java.JavaPlugin;

final class ClickMobsPermissionService {
  private static final String CLICKMOBS_PICKUP = "clickmobs.pickup";
  private static final String CLICKMOBS_PLACE = "clickmobs.place";

  private final JavaPlugin plugin;
  private final Map<UUID, PermissionAttachment> attachments = new HashMap<>();

  ClickMobsPermissionService(JavaPlugin plugin) {
    this.plugin = plugin;
  }

  void ensureClickMobsPermission(Player player) {
    PermissionAttachment attachment = attachments.computeIfAbsent(player.getUniqueId(), uuid -> player.addAttachment(plugin));
    attachment.setPermission(CLICKMOBS_PICKUP, true);
    attachment.setPermission(CLICKMOBS_PLACE, true);
    player.recalculatePermissions();
  }

  void removeAttachment(Player player) {
    PermissionAttachment attachment = attachments.remove(player.getUniqueId());
    if (attachment == null) {
      return;
    }
    player.removeAttachment(attachment);
    player.recalculatePermissions();
  }

  void clear(Collection<? extends Player> players) {
    for (Player player : players) {
      removeAttachment(player);
    }
    attachments.clear();
  }
}
