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
import net.kyori.adventure.text.format.TextDecoration;
import org.slf4j.Logger;

import java.io.BufferedInputStream;
import java.io.BufferedOutputStream;
import java.io.BufferedReader;
import java.io.IOException;
import java.net.Socket;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.security.SecureRandom;
import java.time.Instant;
import java.util.HashSet;
import java.util.Optional;
import java.util.Set;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;

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
    String addr = envOr("MCLINK_REDIS_ADDR", "redis:6379");
    String[] hp = addr.split(":", 2);
    this.redisHost = hp[0];
    this.redisPort = hp.length == 2 ? parseIntOr(hp[1], 6379) : 6379;
    this.redisDB = parseIntOr(envOr("MCLINK_REDIS_DB", "0"), 0);
    this.allowlistPath = envOr("MCLINK_ALLOWLIST_PATH", DEFAULT_ALLOWLIST_PATH);
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

    String code = generateCode();
    long expires = Instant.now().getEpochSecond() + TTL_SECONDS;
    String type = "uuid";
    String value = player.getUniqueId().toString();
    try {
      appendEntry(code, type, value, expires);
    } catch (IOException e) {
      logger.error("failed to write link code for {}", player.getUsername(), e);
      player.sendMessage(Component.text("リンクコード発行に失敗しました。しばらくして再接続してください。", NamedTextColor.RED));
      return;
    }

    sendLinkMessage(player, code);
    logger.info("issued link code {} for {} {} {}", code, player.getUsername(), type, value);
  }

  private void sendLinkMessage(Player player, String code) {
    String cmd = "/mc link code:" + code;
    player.sendMessage(Component.text("手順: Tキーでチャットを開く -> LINK CODEをコピー -> Discordで /mc link を実行", NamedTextColor.YELLOW));
    player.sendMessage(
      Component.text("LINK CODE: ", NamedTextColor.GOLD)
        .append(
          Component.text(code, NamedTextColor.AQUA, TextDecoration.BOLD)
            .clickEvent(ClickEvent.copyToClipboard(code))
            .hoverEvent(HoverEvent.showText(Component.text("クリックでコピー")))
        )
    );
    player.sendMessage(
      Component.text("実行コマンドをコピー: ", NamedTextColor.GRAY)
        .append(
          Component.text(cmd, NamedTextColor.WHITE, TextDecoration.BOLD)
            .clickEvent(ClickEvent.copyToClipboard(cmd))
            .hoverEvent(HoverEvent.showText(Component.text("クリックでコマンドをコピー")))
        )
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
    try (Socket socket = new Socket(redisHost, redisPort);
         BufferedOutputStream out = new BufferedOutputStream(socket.getOutputStream());
         BufferedInputStream in = new BufferedInputStream(socket.getInputStream())) {
      if (redisDB != 0) {
        sendAndExpectOK(out, in, "SELECT", Integer.toString(redisDB));
      }
      sendAndExpectInt(out, in,
        "HSET", key,
        "code", code,
        "type", type,
        "value", value,
        "expires_unix", Long.toString(expiresAt),
        "claimed", "false",
        "claimed_by", "",
        "claimed_at_unix", "0");
      sendAndExpectInt(out, in, "EXPIREAT", key, Long.toString(expiresAt));
    }
  }

  private WhitelistEntries loadWhitelistEntries() {
    Set<String> uuids = new HashSet<>();
    try (BufferedReader reader = Files.newBufferedReader(Paths.get(allowlistPath), StandardCharsets.UTF_8)) {
      String line;
      String section = "";
      while ((line = reader.readLine()) != null) {
        String trimmed = stripComments(line.trim());
        if (trimmed.isEmpty()) {
          continue;
        }
        if ("uuids:".equalsIgnoreCase(trimmed)) {
          section = "uuids";
          continue;
        }
        if (trimmed.startsWith("-")) {
          String value = parseYamlListValue(trimmed);
          if (value.isBlank()) {
            continue;
          }
          if ("uuids".equals(section)) {
            uuids.add(value.toLowerCase());
          }
        }
      }
    } catch (IOException e) {
      logger.warn("failed to read allowlist file: {}", allowlistPath, e);
    }
    return new WhitelistEntries(uuids);
  }

  private static String stripComments(String line) {
    int idx = line.indexOf('#');
    if (idx >= 0) {
      return line.substring(0, idx).trim();
    }
    return line;
  }

  private static String parseYamlListValue(String line) {
    int idx = line.indexOf('-');
    if (idx < 0 || idx + 1 >= line.length()) {
      return "";
    }
    String raw = line.substring(idx + 1).trim();
    if (raw.length() >= 2 && ((raw.startsWith("\"") && raw.endsWith("\"")) || (raw.startsWith("'") && raw.endsWith("'")))) {
      return raw.substring(1, raw.length() - 1).trim();
    }
    return raw.trim();
  }

  private record WhitelistEntries(Set<String> uuidSet) {}

  private static void sendAndExpectOK(BufferedOutputStream out, BufferedInputStream in, String... parts) throws IOException {
    sendCommand(out, parts);
    String line = readLine(in);
    if (!"+OK".equals(line)) {
      throw new IOException("redis response was not OK: " + line);
    }
  }

  private static void sendAndExpectInt(BufferedOutputStream out, BufferedInputStream in, String... parts) throws IOException {
    sendCommand(out, parts);
    String line = readLine(in);
    if (line == null || line.isEmpty() || line.charAt(0) != ':') {
      throw new IOException("redis response was not integer: " + line);
    }
  }

  private static void sendCommand(BufferedOutputStream out, String... parts) throws IOException {
    out.write(("*" + parts.length + "\r\n").getBytes(StandardCharsets.UTF_8));
    for (String part : parts) {
      byte[] bytes = part.getBytes(StandardCharsets.UTF_8);
      out.write(("$" + bytes.length + "\r\n").getBytes(StandardCharsets.UTF_8));
      out.write(bytes);
      out.write("\r\n".getBytes(StandardCharsets.UTF_8));
    }
    out.flush();
  }

  private static String readLine(BufferedInputStream in) throws IOException {
    StringBuilder sb = new StringBuilder();
    while (true) {
      int b = in.read();
      if (b < 0) {
        throw new IOException("redis connection closed");
      }
      if (b == '\r') {
        int next = in.read();
        if (next != '\n') {
          throw new IOException("invalid redis line ending");
        }
        return sb.toString();
      }
      sb.append((char) b);
    }
  }

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
