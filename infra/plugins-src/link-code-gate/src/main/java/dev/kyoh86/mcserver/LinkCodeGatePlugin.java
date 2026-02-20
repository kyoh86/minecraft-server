package dev.kyoh86.mcserver;

import com.google.inject.Inject;
import com.velocitypowered.api.event.PostOrder;
import com.velocitypowered.api.event.ResultedEvent;
import com.velocitypowered.api.event.Subscribe;
import com.velocitypowered.api.event.connection.LoginEvent;
import com.velocitypowered.api.plugin.Plugin;
import com.velocitypowered.api.plugin.annotation.DataDirectory;
import com.velocitypowered.api.proxy.Player;
import net.kyori.adventure.text.Component;
import net.kyori.adventure.text.serializer.plain.PlainTextComponentSerializer;
import org.slf4j.Logger;

import java.io.BufferedInputStream;
import java.io.BufferedOutputStream;
import java.io.IOException;
import java.net.Socket;
import java.nio.charset.StandardCharsets;
import java.nio.file.Path;
import java.security.SecureRandom;
import java.time.Instant;
import java.util.Locale;

@Plugin(
  id = "linkcodegate",
  name = "LinkCodeGate",
  version = "0.1.0",
  description = "Issue one-time link codes when whitelist denies login."
)
public final class LinkCodeGatePlugin {
  private static final String ALPHABET = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ";
  private static final int CODE_LENGTH = 8;
  private static final long TTL_SECONDS = 10 * 60;

  private final Logger logger;
  private final SecureRandom random = new SecureRandom();
  private final String redisHost;
  private final int redisPort;
  private final int redisDB;

  @Inject
  public LinkCodeGatePlugin(Logger logger, @DataDirectory Path dataDirectory) {
    this.logger = logger;
    String addr = envOr("MCLINK_REDIS_ADDR", "redis:6379");
    String[] hp = addr.split(":", 2);
    this.redisHost = hp[0];
    this.redisPort = hp.length == 2 ? parseIntOr(hp[1], 6379) : 6379;
    this.redisDB = parseIntOr(envOr("MCLINK_REDIS_DB", "0"), 0);
  }

  @Subscribe(order = PostOrder.LAST)
  public void onLogin(LoginEvent event) {
    ResultedEvent.ComponentResult result = event.getResult();
    if (result.isAllowed()) {
      return;
    }
    if (!isWhitelistDeny(result)) {
      return;
    }

    Player p = event.getPlayer();
    String code = generateCode();
    long expires = Instant.now().getEpochSecond() + TTL_SECONDS;
    String type;
    String value;
    if (p.getUniqueId() != null) {
      type = "uuid";
      value = p.getUniqueId().toString();
    } else {
      type = "nick";
      value = p.getUsername();
    }

    try {
      appendEntry(code, type, value, expires);
    } catch (IOException e) {
      logger.error("failed to write link code for {}", p.getUsername(), e);
      return;
    }

    String msg = "ホワイトリスト未登録です。Discordで /mc link code:" + code + " を実行してください。";
    event.setResult(ResultedEvent.ComponentResult.denied(Component.text(msg)));
    logger.info("issued link code {} for {} {} {}", code, p.getUsername(), type, value);
  }

  private boolean isWhitelistDeny(ResultedEvent.ComponentResult result) {
    String text = PlainTextComponentSerializer.plainText().serialize(result.getReasonComponent().orElse(Component.empty()));
    String lower = text.toLowerCase(Locale.ROOT);
    return lower.contains("whitelist") || text.contains("ホワイトリスト") || lower.contains("invited to the party");
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
    String key = "mclink:code:" + code;
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
