package dev.kyoh86.mcserver;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import org.slf4j.Logger;

record LinkCodeGateConfig(
  String redisHost,
  int redisPort,
  int redisDB,
  Path allowlistPath,
  String discordGuildName
) {
  private static final String DEFAULT_ALLOWLIST_PATH = "/server/allowlist.yml";
  private static final String DEFAULT_DISCORD_GUILD_NAME_PATH = "/run/secrets/mc_link_discord_guild_name.txt";

  static LinkCodeGateConfig load(Logger logger) {
    String addr = envOr("MC_LINK_REDIS_ADDR", "redis:6379");
    String[] hp = addr.split(":", 2);
    if (hp.length != 2 || hp[0].isBlank() || hp[1].isBlank()) {
      throw new IllegalArgumentException("MC_LINK_REDIS_ADDR must be host:port");
    }
    String redisHost = hp[0].trim();
    int redisPort = parseIntStrict("MC_LINK_REDIS_ADDR port", hp[1]);
    int redisDB = parseIntStrict("MC_LINK_REDIS_DB", envOr("MC_LINK_REDIS_DB", "0"));
    Path allowlistPath = Path.of(envOr("MC_LINK_ALLOWLIST_PATH", DEFAULT_ALLOWLIST_PATH));
    String discordGuildName = resolveDiscordGuildName(logger);
    return new LinkCodeGateConfig(redisHost, redisPort, redisDB, allowlistPath, discordGuildName);
  }

  private static String resolveDiscordGuildName(Logger logger) {
    String fromEnv = envOr("MC_LINK_DISCORD_GUILD_NAME", "");
    if (!fromEnv.isEmpty() && !looksLikePlaceholder(fromEnv)) {
      return fromEnv;
    }
    String path = envOr("MC_LINK_DISCORD_GUILD_NAME_PATH", DEFAULT_DISCORD_GUILD_NAME_PATH);
    if (path.isEmpty()) {
      return "";
    }
    try {
      String value = Files.readString(Path.of(path)).trim();
      if (value.isEmpty() || looksLikePlaceholder(value)) {
        return "";
      }
      return value;
    } catch (IOException e) {
      logger.debug("discord guild name file is not available: {}", path);
      return "";
    }
  }

  private static String envOr(String key, String fallback) {
    String value = System.getenv(key);
    if (value == null || value.isBlank()) {
      return fallback;
    }
    return value.trim();
  }

  private static int parseIntStrict(String label, String value) {
    try {
      return Integer.parseInt(value.trim());
    } catch (Exception e) {
      throw new IllegalArgumentException(label + " must be an integer: " + value, e);
    }
  }

  private static boolean looksLikePlaceholder(String value) {
    return value.startsWith("REPLACE_WITH_");
  }
}
