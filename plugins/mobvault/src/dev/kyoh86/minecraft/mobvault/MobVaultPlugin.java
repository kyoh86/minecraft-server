package dev.kyoh86.minecraft.mobvault;

import java.io.IOException;
import java.io.StringReader;
import java.io.StringWriter;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.time.OffsetDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.Properties;
import java.util.UUID;
import org.bukkit.Bukkit;
import org.bukkit.DyeColor;
import org.bukkit.Material;
import org.bukkit.NamespacedKey;
import org.bukkit.World;
import org.bukkit.attribute.Attribute;
import org.bukkit.attribute.AttributeInstance;
import org.bukkit.command.Command;
import org.bukkit.command.CommandSender;
import org.bukkit.entity.Ageable;
import org.bukkit.entity.Cat;
import org.bukkit.entity.Entity;
import org.bukkit.entity.EntityType;
import org.bukkit.entity.LivingEntity;
import org.bukkit.entity.Player;
import org.bukkit.entity.Sittable;
import org.bukkit.entity.Tameable;
import org.bukkit.entity.Wolf;
import org.bukkit.event.EventHandler;
import org.bukkit.event.EventPriority;
import org.bukkit.event.Listener;
import org.bukkit.event.entity.EntityDamageByEntityEvent;
import org.bukkit.event.inventory.InventoryClickEvent;
import org.bukkit.event.player.PlayerInteractEntityEvent;
import org.bukkit.inventory.Inventory;
import org.bukkit.inventory.InventoryHolder;
import org.bukkit.inventory.ItemStack;
import org.bukkit.inventory.meta.ItemMeta;
import org.bukkit.persistence.PersistentDataType;
import org.bukkit.plugin.java.JavaPlugin;
import org.bukkit.potion.PotionEffect;
import io.papermc.paper.registry.RegistryAccess;
import io.papermc.paper.registry.RegistryKey;
import net.kyori.adventure.text.Component;
import net.kyori.adventure.text.format.NamedTextColor;
import net.kyori.adventure.text.serializer.plain.PlainTextComponentSerializer;

public final class MobVaultPlugin extends JavaPlugin implements Listener {
  private String jdbcUrl;
  private String dbUser;
  private String dbPassword;
  private int connectTimeoutSeconds;
  private String sourceServer;
  private double searchRadius;
  private int listLimit;
  private Material toolMaterial;
  private String toolDisplayName;
  private NamespacedKey toolKey;
  private NamespacedKey vaultEntryIdKey;
  private final PlainTextComponentSerializer plainText = PlainTextComponentSerializer.plainText();

  @Override
  public void onEnable() {
    saveDefaultConfig();
    toolKey = new NamespacedKey(this, "mobvault_tool");
    vaultEntryIdKey = new NamespacedKey(this, "mobvault_entry_id");
    loadSettings();
    ensureDriverLoaded();
    createTableIfNeeded();
    getServer().getPluginManager().registerEvents(this, this);
    getLogger().info("MobVault enabled");
  }

  @Override
  public boolean onCommand(CommandSender sender, Command command, String label, String[] args) {
    if (!(sender instanceof Player player)) {
      sender.sendMessage("player only");
      return true;
    }
    if (args.length < 1) {
      sendUsage(player);
      return true;
    }
    return switch (args[0].toLowerCase()) {
      case "tool" -> onTool(player);
      case "deposit" -> onDeposit(player);
      case "list" -> onList(player);
      case "withdraw" -> onWithdraw(player, args);
      default -> {
        sendUsage(player);
        yield true;
      }
    };
  }

  private void loadSettings() {
    reloadConfig();
    jdbcUrl = getConfig().getString("database.jdbc_url", "jdbc:postgresql://postgres:5432/minecraft");
    dbUser = getConfig().getString("database.user", "minecraft");
    dbPassword = getConfig().getString("database.password", "minecraft");
    connectTimeoutSeconds = getConfig().getInt("database.connect_timeout_seconds", 5);
    sourceServer = getConfig().getString("vault.source_server", "lobby");
    searchRadius = getConfig().getDouble("vault.search_radius", 4.0d);
    listLimit = getConfig().getInt("vault.list_limit", 20);
    toolDisplayName = getConfig().getString("tool.display_name", "MobVault Wand");
    String materialName = getConfig().getString("tool.material", "CARROT_ON_A_STICK");
    Material material = Material.matchMaterial(materialName);
    if (material == null || material.isAir()) {
      throw new IllegalArgumentException("invalid tool.material: " + materialName);
    }
    toolMaterial = material;
  }

  private void createTableIfNeeded() {
    String sql =
        """
        CREATE TABLE IF NOT EXISTS mob_vault_entries (
          id UUID PRIMARY KEY,
          owner_uuid UUID NOT NULL,
          source_server TEXT NOT NULL,
          entity_type TEXT NOT NULL,
          payload TEXT NOT NULL,
          created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        )
        """;
    try (Connection connection = openConnection();
        PreparedStatement statement = connection.prepareStatement(sql)) {
      statement.execute();
    } catch (SQLException e) {
      throw new IllegalStateException("Failed to create mob_vault_entries table", e);
    }
  }

  private Connection openConnection() throws SQLException {
    Properties properties = new Properties();
    properties.setProperty("user", dbUser);
    properties.setProperty("password", dbPassword);
    properties.setProperty("connectTimeout", Integer.toString(connectTimeoutSeconds));
    return DriverManager.getConnection(jdbcUrl, properties);
  }

  private void ensureDriverLoaded() {
    try {
      Class.forName("org.postgresql.Driver");
    } catch (ClassNotFoundException e) {
      throw new IllegalStateException("PostgreSQL JDBC driver is not available", e);
    }
  }

  private boolean onDeposit(Player player) {
    Optional<LivingEntity> targetOpt = findNearestTarget(player);
    if (targetOpt.isEmpty()) {
      info(player, NamedTextColor.RED, "近くに預けられるモブが見つかりません");
      return true;
    }

    return depositEntity(player, targetOpt.get());
  }

  private boolean onList(Player player) {
    openVaultGui(player);
    return true;
  }

  private boolean onWithdraw(Player player, String[] args) {
    if (args.length < 2) {
      info(player, NamedTextColor.RED, "usage: /mobvault withdraw <uuid>");
      return true;
    }

    UUID id;
    try {
      id = UUID.fromString(args[1]);
    } catch (IllegalArgumentException e) {
      info(player, NamedTextColor.RED, "UUID が不正です");
      return true;
    }

    withdrawById(player, id);
    return true;
  }

  private Optional<LivingEntity> findNearestTarget(Player player) {
    double radiusSquared = searchRadius * searchRadius;
    LivingEntity nearest = null;
    double nearestDistanceSquared = Double.MAX_VALUE;

    for (Entity entity :
        player.getWorld().getNearbyEntities(player.getLocation(), searchRadius, searchRadius, searchRadius)) {
      if (!(entity instanceof LivingEntity livingEntity)) {
        continue;
      }
      if (livingEntity instanceof Player) {
        continue;
      }
      if (livingEntity.getType() == EntityType.ARMOR_STAND) {
        continue;
      }
      double distanceSquared = livingEntity.getLocation().distanceSquared(player.getLocation());
      if (distanceSquared > radiusSquared) {
        continue;
      }
      if (distanceSquared < nearestDistanceSquared) {
        nearest = livingEntity;
        nearestDistanceSquared = distanceSquared;
      }
    }
    return Optional.ofNullable(nearest);
  }

  private boolean onTool(Player player) {
    ItemStack tool = new ItemStack(toolMaterial);
    ItemMeta meta = tool.getItemMeta();
    meta.displayName(Component.text(toolDisplayName, NamedTextColor.AQUA));
    meta.getPersistentDataContainer().set(toolKey, PersistentDataType.BYTE, (byte) 1);
    tool.setItemMeta(meta);
    player.getInventory().addItem(tool);
    info(player, NamedTextColor.GREEN, "MobVaultツールを配布しました");
    return true;
  }

  @EventHandler(priority = EventPriority.HIGHEST, ignoreCancelled = true)
  public void onRightClickEntity(PlayerInteractEntityEvent event) {
    if (!(event.getRightClicked() instanceof LivingEntity target)) {
      return;
    }
    Player player = event.getPlayer();
    if (!hasTool(player)) {
      return;
    }
    event.setCancelled(true);
    depositEntity(player, target);
  }

  @EventHandler(priority = EventPriority.HIGHEST, ignoreCancelled = true)
  public void onLeftClickEntity(EntityDamageByEntityEvent event) {
    if (!(event.getDamager() instanceof Player player)) {
      return;
    }
    if (!(event.getEntity() instanceof LivingEntity target)) {
      return;
    }
    if (!hasTool(player)) {
      return;
    }
    event.setCancelled(true);
    depositEntity(player, target);
  }

  @EventHandler(priority = EventPriority.HIGHEST, ignoreCancelled = true)
  public void onInventoryClick(InventoryClickEvent event) {
    if (!(event.getWhoClicked() instanceof Player player)) {
      return;
    }
    if (!(event.getInventory().getHolder() instanceof MobVaultInventoryHolder)) {
      return;
    }

    event.setCancelled(true);
    ItemStack item = event.getCurrentItem();
    if (item == null || !item.hasItemMeta()) {
      return;
    }

    String idValue =
        item.getItemMeta().getPersistentDataContainer().get(vaultEntryIdKey, PersistentDataType.STRING);
    if (idValue == null || idValue.isBlank()) {
      return;
    }

    UUID id;
    try {
      id = UUID.fromString(idValue);
    } catch (IllegalArgumentException e) {
      return;
    }

    boolean ok = withdrawById(player, id);
    if (ok) {
      Bukkit.getScheduler().runTask(this, () -> openVaultGui(player));
    }
  }

  private boolean hasTool(Player player) {
    ItemStack hand = player.getInventory().getItemInMainHand();
    if (hand.getType() != toolMaterial || !hand.hasItemMeta()) {
      return false;
    }
    Byte marker =
        hand.getItemMeta().getPersistentDataContainer().get(toolKey, PersistentDataType.BYTE);
    return marker != null && marker == (byte) 1;
  }

  private boolean depositEntity(Player player, LivingEntity target) {
    if (target instanceof Player || target.getType() == EntityType.ARMOR_STAND) {
      info(player, NamedTextColor.RED, "この対象は預けられません");
      return true;
    }

    if (target instanceof Tameable tameable
        && tameable.isTamed()
        && tameable.getOwnerUniqueId() != null
        && !player.getUniqueId().equals(tameable.getOwnerUniqueId())) {
      info(player, NamedTextColor.RED, "他人のテイム済みモブは預けられません");
      return true;
    }

    UUID entryId = UUID.randomUUID();
    String payload = serializeEntity(target);
    String sql =
        "INSERT INTO mob_vault_entries (id, owner_uuid, source_server, entity_type, payload) VALUES (?, ?, ?, ?, ?)";
    try (Connection connection = openConnection();
        PreparedStatement statement = connection.prepareStatement(sql)) {
      statement.setObject(1, entryId);
      statement.setObject(2, player.getUniqueId());
      statement.setString(3, sourceServer);
      statement.setString(4, target.getType().name());
      statement.setString(5, payload);
      statement.executeUpdate();
    } catch (SQLException e) {
      getLogger().warning("Failed to save mob: " + e.getMessage());
      info(player, NamedTextColor.RED, "保存に失敗しました");
      return true;
    }

    target.remove();
    info(
        player,
        NamedTextColor.GREEN,
        "預けました: " + target.getType().name() + " id=" + entryId.toString().substring(0, 8));
    return true;
  }

  private void openVaultGui(Player player) {
    List<VaultEntry> entries;
    try {
      entries = loadEntries(player.getUniqueId(), Math.min(listLimit, 45));
    } catch (SQLException e) {
      getLogger().warning("Failed to load entries for GUI: " + e.getMessage());
      info(player, NamedTextColor.RED, "一覧取得に失敗しました");
      return;
    }

    if (entries.isEmpty()) {
      info(player, NamedTextColor.YELLOW, "預けられたモブはありません");
      return;
    }

    int size = Math.max(9, ((entries.size() - 1) / 9 + 1) * 9);
    Inventory inventory =
        Bukkit.createInventory(new MobVaultInventoryHolder(), size, Component.text("MobVault", NamedTextColor.AQUA));

    for (int i = 0; i < entries.size() && i < size; i++) {
      VaultEntry entry = entries.get(i);
      inventory.setItem(i, createEntryItem(entry));
    }
    player.openInventory(inventory);
  }

  private List<VaultEntry> loadEntries(UUID ownerUuid, int limit) throws SQLException {
    String sql =
        "SELECT id, entity_type, source_server, created_at FROM mob_vault_entries WHERE owner_uuid = ? ORDER BY created_at DESC LIMIT ?";
    List<VaultEntry> entries = new ArrayList<>();
    try (Connection connection = openConnection();
        PreparedStatement statement = connection.prepareStatement(sql)) {
      statement.setObject(1, ownerUuid);
      statement.setInt(2, limit);
      try (ResultSet resultSet = statement.executeQuery()) {
        while (resultSet.next()) {
          entries.add(
              new VaultEntry(
                  (UUID) resultSet.getObject("id"),
                  resultSet.getString("entity_type"),
                  resultSet.getString("source_server"),
                  resultSet.getObject("created_at", OffsetDateTime.class)));
        }
      }
    }
    return entries;
  }

  private ItemStack createEntryItem(VaultEntry entry) {
    Material material = guessIconMaterial(entry.entityType());
    ItemStack item = new ItemStack(material);
    ItemMeta meta = item.getItemMeta();
    meta.displayName(Component.text(entry.entityType(), NamedTextColor.GREEN));
    List<Component> lore = new ArrayList<>();
    lore.add(Component.text("id: " + entry.id(), NamedTextColor.GRAY));
    lore.add(Component.text("from: " + entry.sourceServer(), NamedTextColor.GRAY));
    lore.add(Component.text("at: " + entry.createdAt(), NamedTextColor.DARK_GRAY));
    lore.add(Component.text("クリックで引き出し", NamedTextColor.YELLOW));
    meta.lore(lore);
    meta.getPersistentDataContainer().set(vaultEntryIdKey, PersistentDataType.STRING, entry.id().toString());
    item.setItemMeta(meta);
    return item;
  }

  private Material guessIconMaterial(String entityType) {
    return switch (entityType) {
      case "WOLF" -> Material.BONE;
      case "CAT" -> Material.COD;
      case "HORSE" -> Material.SADDLE;
      case "PARROT" -> Material.FEATHER;
      default -> Material.NAME_TAG;
    };
  }

  private boolean withdrawById(Player player, UUID id) {
    String selectSql =
        "SELECT entity_type, payload FROM mob_vault_entries WHERE id = ? AND owner_uuid = ?";
    String deleteSql = "DELETE FROM mob_vault_entries WHERE id = ? AND owner_uuid = ?";

    try (Connection connection = openConnection()) {
      connection.setAutoCommit(false);

      String entityTypeName;
      String payload;
      try (PreparedStatement select = connection.prepareStatement(selectSql)) {
        select.setObject(1, id);
        select.setObject(2, player.getUniqueId());
        try (ResultSet resultSet = select.executeQuery()) {
          if (!resultSet.next()) {
            connection.rollback();
            info(player, NamedTextColor.RED, "指定IDの預かりモブが見つかりません");
            return false;
          }
          entityTypeName = resultSet.getString("entity_type");
          payload = resultSet.getString("payload");
        }
      }

      EntityType entityType = EntityType.valueOf(entityTypeName);
      Entity spawned = player.getWorld().spawnEntity(player.getLocation(), entityType);
      if (!(spawned instanceof LivingEntity livingEntity)) {
        spawned.remove();
        connection.rollback();
        info(player, NamedTextColor.RED, "このモブ種は現在引き出し非対応です");
        return false;
      }

      try {
        applyPayload(livingEntity, payload);
      } catch (RuntimeException e) {
        livingEntity.remove();
        connection.rollback();
        info(player, NamedTextColor.RED, "引き出し復元に失敗しました");
        getLogger().warning("Failed to apply payload: " + e.getMessage());
        return false;
      }

      try (PreparedStatement delete = connection.prepareStatement(deleteSql)) {
        delete.setObject(1, id);
        delete.setObject(2, player.getUniqueId());
        delete.executeUpdate();
      }
      connection.commit();
      info(player, NamedTextColor.GREEN, "引き出しました: " + entityTypeName);
      return true;
    } catch (SQLException e) {
      getLogger().warning("Failed to withdraw mob: " + e.getMessage());
      info(player, NamedTextColor.RED, "引き出しに失敗しました");
      return false;
    }
  }

  private String serializeEntity(LivingEntity entity) {
    Properties payload = new Properties();
    payload.setProperty("type", entity.getType().name());
    Component customName = entity.customName();
    payload.setProperty("customName", customName == null ? "" : plainText.serialize(customName));
    payload.setProperty("customNameVisible", Boolean.toString(entity.isCustomNameVisible()));
    payload.setProperty("ai", Boolean.toString(entity.hasAI()));
    payload.setProperty("silent", Boolean.toString(entity.isSilent()));
    payload.setProperty("glowing", Boolean.toString(entity.isGlowing()));
    payload.setProperty("gravity", Boolean.toString(entity.hasGravity()));
    payload.setProperty("invulnerable", Boolean.toString(entity.isInvulnerable()));
    payload.setProperty("health", Double.toString(entity.getHealth()));

    AttributeInstance maxHealthAttribute = entity.getAttribute(Attribute.MAX_HEALTH);
    if (maxHealthAttribute != null) {
      payload.setProperty("maxHealth", Double.toString(maxHealthAttribute.getBaseValue()));
    }

    if (entity instanceof Ageable ageable) {
      payload.setProperty("age", Integer.toString(ageable.getAge()));
    }

    if (entity instanceof Tameable tameable) {
      payload.setProperty("tamed", Boolean.toString(tameable.isTamed()));
      if (tameable.getOwnerUniqueId() != null) {
        payload.setProperty("ownerUuid", tameable.getOwnerUniqueId().toString());
      }
      if (entity instanceof Sittable sittable) {
        payload.setProperty("sitting", Boolean.toString(sittable.isSitting()));
      }
    }

    if (entity instanceof Wolf wolf) {
      payload.setProperty("wolfAngry", Boolean.toString(wolf.isAngry()));
      payload.setProperty("wolfInterested", Boolean.toString(wolf.isInterested()));
      payload.setProperty("wolfTailAngle", Float.toString(wolf.getTailAngle()));
      payload.setProperty("wolfCollarColor", wolf.getCollarColor().name());
    }

    if (entity instanceof Cat cat) {
      payload.setProperty(
          "catType",
          RegistryAccess.registryAccess()
              .getRegistry(RegistryKey.CAT_VARIANT)
              .getKeyOrThrow(cat.getCatType())
              .toString());
      payload.setProperty("catCollarColor", cat.getCollarColor().name());
    }

    int idx = 0;
    for (PotionEffect potionEffect : entity.getActivePotionEffects()) {
      payload.setProperty(
          "effect."
              + idx,
          potionEffect.getType().getKey().toString()
              + ","
              + potionEffect.getDuration()
              + ","
              + potionEffect.getAmplifier()
              + ","
              + potionEffect.isAmbient()
              + ","
              + potionEffect.hasParticles()
              + ","
              + potionEffect.hasIcon());
      idx++;
    }
    payload.setProperty("effect.count", Integer.toString(idx));

    try {
      StringWriter writer = new StringWriter();
      payload.store(writer, null);
      return writer.toString();
    } catch (IOException e) {
      throw new IllegalStateException("Failed to serialize payload", e);
    }
  }

  private void applyPayload(LivingEntity entity, String serialized) {
    Properties payload = new Properties();
    try {
      payload.load(new StringReader(serialized));
    } catch (IOException e) {
      throw new IllegalStateException("Invalid payload", e);
    }

    if (payload.containsKey("customName")) {
      String customName = payload.getProperty("customName");
      entity.customName(customName.isBlank() ? null : Component.text(customName));
    }
    entity.setCustomNameVisible(Boolean.parseBoolean(payload.getProperty("customNameVisible", "false")));
    entity.setAI(Boolean.parseBoolean(payload.getProperty("ai", "true")));
    entity.setSilent(Boolean.parseBoolean(payload.getProperty("silent", "false")));
    entity.setGlowing(Boolean.parseBoolean(payload.getProperty("glowing", "false")));
    entity.setGravity(Boolean.parseBoolean(payload.getProperty("gravity", "true")));
    entity.setInvulnerable(Boolean.parseBoolean(payload.getProperty("invulnerable", "false")));

    AttributeInstance maxHealthAttribute = entity.getAttribute(Attribute.MAX_HEALTH);
    if (maxHealthAttribute != null && payload.containsKey("maxHealth")) {
      maxHealthAttribute.setBaseValue(Double.parseDouble(payload.getProperty("maxHealth")));
    }

    if (payload.containsKey("health")) {
      double health = Double.parseDouble(payload.getProperty("health"));
      double maxHealth = entity.getAttribute(Attribute.MAX_HEALTH) != null
          ? entity.getAttribute(Attribute.MAX_HEALTH).getValue()
          : entity.getHealth();
      entity.setHealth(Math.max(0.1d, Math.min(health, maxHealth)));
    }

    if (entity instanceof Ageable ageable) {
      if (payload.containsKey("age")) {
        ageable.setAge(Integer.parseInt(payload.getProperty("age")));
      }
    }

    if (entity instanceof Tameable tameable) {
      boolean tamed = Boolean.parseBoolean(payload.getProperty("tamed", "false"));
      tameable.setTamed(tamed);
      if (tamed && payload.containsKey("ownerUuid")) {
        UUID ownerUuid = UUID.fromString(payload.getProperty("ownerUuid"));
        Player online = getServer().getPlayer(ownerUuid);
        if (online != null) {
          tameable.setOwner(online);
        }
      }
      if (entity instanceof Sittable sittable) {
        sittable.setSitting(Boolean.parseBoolean(payload.getProperty("sitting", "false")));
      }
    }

    if (entity instanceof Wolf wolf) {
      wolf.setAngry(Boolean.parseBoolean(payload.getProperty("wolfAngry", "false")));
      wolf.setInterested(Boolean.parseBoolean(payload.getProperty("wolfInterested", "false")));
      if (payload.containsKey("wolfCollarColor")) {
        wolf.setCollarColor(DyeColor.valueOf(payload.getProperty("wolfCollarColor")));
      }
    }

    if (entity instanceof Cat cat) {
      if (payload.containsKey("catType")) {
        NamespacedKey key = NamespacedKey.fromString(payload.getProperty("catType"));
        if (key != null) {
          Cat.Type variant =
              RegistryAccess.registryAccess().getRegistry(RegistryKey.CAT_VARIANT).get(key);
          if (variant != null) {
            cat.setCatType(variant);
          }
        }
      }
      if (payload.containsKey("catCollarColor")) {
        cat.setCollarColor(DyeColor.valueOf(payload.getProperty("catCollarColor")));
      }
    }
  }

  private void sendUsage(Player player) {
    info(player, NamedTextColor.YELLOW, "usage: /mobvault tool");
    info(player, NamedTextColor.YELLOW, "usage: /mobvault deposit");
    info(player, NamedTextColor.YELLOW, "usage: /mobvault list");
    info(player, NamedTextColor.YELLOW, "usage: /mobvault withdraw <uuid>");
  }

  private static void info(Player player, NamedTextColor color, String message) {
    player.sendMessage(Component.text(message, color));
  }

  private record VaultEntry(UUID id, String entityType, String sourceServer, OffsetDateTime createdAt) {}

  private static final class MobVaultInventoryHolder implements InventoryHolder {
    @Override
    public Inventory getInventory() {
      return Bukkit.createInventory(this, 9);
    }
  }
}
