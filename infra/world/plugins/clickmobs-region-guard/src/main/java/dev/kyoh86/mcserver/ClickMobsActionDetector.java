package dev.kyoh86.mcserver;

import org.bukkit.NamespacedKey;
import org.bukkit.entity.Boat;
import org.bukkit.entity.HumanEntity;
import org.bukkit.entity.LivingEntity;
import org.bukkit.entity.Minecart;
import org.bukkit.event.player.PlayerInteractEntityEvent;
import org.bukkit.inventory.ItemStack;
import org.bukkit.inventory.meta.ItemMeta;
import org.bukkit.persistence.PersistentDataContainer;
import org.bukkit.persistence.PersistentDataType;

final class ClickMobsActionDetector {
  private static final NamespacedKey CLICKMOBS_ENTITY_KEY = new NamespacedKey("clickmobs", "entity");

  private ClickMobsActionDetector() {}

  static boolean isPickupAttempt(PlayerInteractEntityEvent event) {
    if (!event.getPlayer().isSneaking()) {
      return false;
    }
    if (!(event.getRightClicked() instanceof LivingEntity)) {
      return false;
    }
    return !(event.getRightClicked() instanceof HumanEntity);
  }

  static boolean isVehiclePlaceAttempt(PlayerInteractEntityEvent event) {
    if (!event.getPlayer().isSneaking()) {
      return false;
    }
    if (!(event.getRightClicked() instanceof Boat) && !(event.getRightClicked() instanceof Minecart)) {
      return false;
    }
    return isClickMobsItem(event.getPlayer().getInventory().getItemInMainHand());
  }

  static boolean isClickMobsItem(ItemStack item) {
    if (item == null || item.getType().isAir()) {
      return false;
    }
    if (!item.getType().toString().equals("PLAYER_HEAD")) {
      return false;
    }
    ItemMeta meta = item.getItemMeta();
    if (meta == null) {
      return false;
    }
    PersistentDataContainer pdc = meta.getPersistentDataContainer();
    return pdc.has(CLICKMOBS_ENTITY_KEY, PersistentDataType.BOOLEAN);
  }
}
