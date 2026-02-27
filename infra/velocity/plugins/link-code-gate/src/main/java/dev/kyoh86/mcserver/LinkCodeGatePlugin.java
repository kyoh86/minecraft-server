package dev.kyoh86.mcserver;

import com.google.inject.Inject;
import com.velocitypowered.api.event.player.ServerConnectedEvent;
import com.velocitypowered.api.event.player.PlayerChooseInitialServerEvent;
import com.velocitypowered.api.event.player.ServerPreConnectEvent;
import com.velocitypowered.api.event.PostOrder;
import com.velocitypowered.api.event.Subscribe;
import com.velocitypowered.api.plugin.Plugin;
import com.velocitypowered.api.proxy.Player;
import com.velocitypowered.api.proxy.ProxyServer;
import com.velocitypowered.api.proxy.server.RegisteredServer;
import net.kyori.adventure.text.Component;
import net.kyori.adventure.text.event.ClickEvent;
import net.kyori.adventure.text.event.HoverEvent;
import net.kyori.adventure.text.format.NamedTextColor;
import org.slf4j.Logger;
import redis.clients.jedis.Jedis;

import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.security.SecureRandom;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.HashSet;
import java.util.Optional;
import java.util.Set;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.Map;
import org.yaml.snakeyaml.Yaml;

@Plugin(
  id = "linkcodegate",
  name = "LinkCodeGate",
  version = "0.1.0",
  description = "Route unlinked players to limbo and issue one-time link codes."
)
public final class LinkCodeGatePlugin {
  private static final String ALPHABET = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ";
  private static final int CODE_LENGTH = 8;
  private static final long TTL_SECONDS = 10 * 60;
  private static final String MAINHALL_SERVER = "mainhall";
  private static final String LIMBO_SERVER = "limbo";
  private static final String DEFAULT_ALLOWLIST_PATH = "/server/allowlist.yml";
  private static final long NOTICE_INTERVAL_MILLIS = 3_000;
  private static final int REDIS_CONNECT_TIMEOUT_MILLIS = 1_500;
  private static final int REDIS_READ_TIMEOUT_MILLIS = 1_500;

  private final ProxyServer proxy;
  private final Logger logger;
  private final SecureRandom random = new SecureRandom();
  private final String redisHost;
  private final int redisPort;
  private final int redisDB;
  private final String allowlistPath;
  private final ConcurrentMap<UUID, Long> lastNoticeAt = new ConcurrentHashMap<>();

  @Inject
  public LinkCodeGatePlugin(ProxyServer proxy, Logger logger) {
    this.proxy = proxy;
    this.logger = logger;
    String addr = envOr("MC_LINK_REDIS_ADDR", "redis:6379");
    String[] hp = addr.split(":", 2);
    this.redisHost = hp[0];
    this.redisPort = hp.length == 2 ? parseIntOr(hp[1], 6379) : 6379;
    this.redisDB = parseIntOr(envOr("MC_LINK_REDIS_DB", "0"), 0);
    this.allowlistPath = envOr("MC_LINK_ALLOWLIST_PATH", DEFAULT_ALLOWLIST_PATH);
  }

  @Subscribe(order = PostOrder.LAST)
  public void onChooseInitialServer(PlayerChooseInitialServerEvent event) {
    Player player = event.getPlayer();
    Optional<RegisteredServer> initial = isAllowed(player) ? serverByName(MAINHALL_SERVER) : serverByName(LIMBO_SERVER);
    initial.ifPresent(event::setInitialServer);
  }

  @Subscribe(order = PostOrder.LAST)
  public void onServerPreConnect(ServerPreConnectEvent event) {
    Player player = event.getPlayer();
    if (isAllowed(player)) {
      return;
    }
    String target = event.getOriginalServer().getServerInfo().getName();
    if (LIMBO_SERVER.equalsIgnoreCase(target)) {
      return;
    }
    Optional<RegisteredServer> limbo = serverByName(LIMBO_SERVER);
    if (limbo.isPresent()) {
      event.setResult(ServerPreConnectEvent.ServerResult.allowed(limbo.get()));
    } else {
      event.setResult(ServerPreConnectEvent.ServerResult.denied());
    }
  }

  @Subscribe(order = PostOrder.LAST)
  public void onServerConnected(ServerConnectedEvent event) {
    Player player = event.getPlayer();
    if (isAllowed(player)) {
      return;
    }
    String serverName = event.getServer().getServerInfo().getName();
    if (!LIMBO_SERVER.equalsIgnoreCase(serverName)) {
      return;
    }
    long now = System.currentTimeMillis();
    long last = lastNoticeAt.getOrDefault(player.getUniqueId(), 0L);
    if (now-last < NOTICE_INTERVAL_MILLIS) {
      return;
    }
    lastNoticeAt.put(player.getUniqueId(), now);

    UUID playerUUID = player.getUniqueId();
    String playerName = player.getUsername();
    String code = generateCode();
    long expires = Instant.now().getEpochSecond() + TTL_SECONDS;
    String type = "uuid";
    String value = playerUUID.toString();

    proxy.getScheduler().buildTask(this, () -> {
      try {
        appendEntry(code, type, value, expires);
      } catch (IOException e) {
        logger.error("failed to write link code for {}", playerName, e);
        proxy.getPlayer(playerUUID).ifPresent(p ->
          p.sendMessage(Component.text("リンクコード発行に失敗しました。しばらくして再接続してください。", NamedTextColor.RED))
        );
        return;
      }
      proxy.getPlayer(playerUUID).ifPresent(p -> sendLinkMessage(p, code));
      logger.info("issued link code for {} {} {}", playerName, type, value);
    }).schedule();
  }

  private void sendLinkMessage(Player player, String code) {
    String cmd = "/mc link code:" + code;
    player.sendMessage(
      Component.text("クリックしてコマンドをコピーしてください: ", NamedTextColor.GRAY)
        .append(
          Component.text(" [" + cmd + "]", NamedTextColor.WHITE)
            .clickEvent(ClickEvent.copyToClipboard(cmd))
            .hoverEvent(HoverEvent.showText(Component.text("クリックでコマンドをコピー")))
        )
        .append(Component.text("コピーしたコマンドをDiscordで送信してください", NamedTextColor.GRAY))
    );
  }

  private Optional<RegisteredServer> serverByName(String name) {
    return proxy.getServer(name);
  }

  private boolean isAllowed(Player player) {
    WhitelistEntries entries = loadWhitelistEntries();
    return entries.uuidSet.contains(player.getUniqueId().toString().toLowerCase());
  }

  private String generateCode() {
    StringBuilder sb = new StringBuilder(CODE_LENGTH);
    for (int i = 0; i < CODE_LENGTH; i++) {
      int idx = random.nextInt(ALPHABET.length());
      sb.append(ALPHABET.charAt(idx));
    }
    return sb.toString();
  }

  private void appendEntry(String code, String type, String value, long expiresAt) throws IOException {
    String key = "mc-link:code:" + code;
    try (Jedis jedis = new Jedis(redisHost, redisPort, REDIS_CONNECT_TIMEOUT_MILLIS, REDIS_READ_TIMEOUT_MILLIS)) {
      if (redisDB != 0) {
        jedis.select(redisDB);
      }
      jedis.hset(key, Map.of(
        "code", code,
        "type", type,
        "value", value,
        "expires_unix", Long.toString(expiresAt),
        "claimed", "false",
        "claimed_by", "",
        "claimed_at_unix", "0"
      ));
      jedis.expireAt(key, expiresAt);
    } catch (RuntimeException e) {
      throw new IOException("failed to write code entry to redis", e);
    }
  }

  private WhitelistEntries loadWhitelistEntries() {
    Set<String> uuids = new HashSet<>();
    try (InputStream in = Files.newInputStream(Paths.get(allowlistPath))) {
      Yaml yaml = new Yaml();
      AllowlistFile allowlist = yaml.loadAs(in, AllowlistFile.class);
      if (allowlist == null || allowlist.uuids == null) {
        return new WhitelistEntries(uuids);
      }
      for (String uuid : allowlist.uuids) {
        if (uuid == null || uuid.isBlank()) {
          continue;
        }
        uuids.add(uuid.trim().toLowerCase());
      }
    } catch (IOException e) {
      logger.warn("failed to read allowlist file: {}", allowlistPath, e);
    } catch (Exception e) {
      logger.warn("failed to parse allowlist file: {}", allowlistPath, e);
    }
    return new WhitelistEntries(uuids);
  }

  private static final class AllowlistFile {
    public List<String> uuids = new ArrayList<>();
    public List<String> nicks = new ArrayList<>();
  }

  private record WhitelistEntries(Set<String> uuidSet) {}

  private static String envOr(String key, String fallback) {
    String value = System.getenv(key);
    if (value == null || value.isBlank()) {
      return fallback;
    }
    return value.trim();
  }

  private static int parseIntOr(String value, int fallback) {
    try {
      return Integer.parseInt(value.trim());
    } catch (Exception ignored) {
      return fallback;
    }
  }
}
