package dev.kyoh86.mcserver;

import com.google.inject.Inject;
import com.velocitypowered.api.event.PostOrder;
import com.velocitypowered.api.event.Subscribe;
import com.velocitypowered.api.event.connection.LoginEvent;
import com.velocitypowered.api.event.ResultedEvent;
import com.velocitypowered.api.plugin.Plugin;
import com.velocitypowered.api.plugin.annotation.DataDirectory;
import com.velocitypowered.api.proxy.Player;
import net.kyori.adventure.text.Component;
import net.kyori.adventure.text.serializer.plain.PlainTextComponentSerializer;
import org.slf4j.Logger;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.SecureRandom;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
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
  private final Path storePath;
  private final SecureRandom random = new SecureRandom();
  private final Object lock = new Object();

  @Inject
  public LinkCodeGatePlugin(Logger logger, @DataDirectory Path dataDirectory) {
    this.logger = logger;
    this.storePath = Path.of("/server", ".wslctl", "mclink-codes.tsv");
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

    String msg = "ホワイトリスト未登録です。Discordで /mc link " + code + " を実行してください。";
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
    synchronized (lock) {
      Path dir = storePath.getParent();
      Files.createDirectories(dir);

      List<String> lines = new ArrayList<>();
      lines.add("# code\ttype\tvalue\texpires_unix\tclaimed\tclaimed_by\tclaimed_at_unix");
      if (Files.exists(storePath)) {
        for (String line : Files.readAllLines(storePath, StandardCharsets.UTF_8)) {
          if (line == null || line.isBlank() || line.startsWith("#")) {
            continue;
          }
          String[] cols = line.split("\t", -1);
          if (cols.length < 7) {
            continue;
          }
          long exp;
          try {
            exp = Long.parseLong(cols[3]);
          } catch (NumberFormatException e) {
            continue;
          }
          if (exp < Instant.now().getEpochSecond()) {
            continue;
          }
          lines.add(line);
        }
      }
      lines.add(code + "\t" + type + "\t" + value + "\t" + expiresAt + "\tfalse\t\t0");
      Files.write(storePath, lines, StandardCharsets.UTF_8);
    }
  }
}
