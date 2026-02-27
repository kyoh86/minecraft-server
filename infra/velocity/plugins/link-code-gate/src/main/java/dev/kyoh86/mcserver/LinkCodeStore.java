package dev.kyoh86.mcserver;

import java.io.IOException;
import java.util.Map;
import redis.clients.jedis.Jedis;

final class LinkCodeStore {
  private static final int REDIS_CONNECT_TIMEOUT_MILLIS = 1_500;
  private static final int REDIS_READ_TIMEOUT_MILLIS = 1_500;

  private final String redisHost;
  private final int redisPort;
  private final int redisDB;

  LinkCodeStore(String redisHost, int redisPort, int redisDB) {
    this.redisHost = redisHost;
    this.redisPort = redisPort;
    this.redisDB = redisDB;
  }

  void appendEntry(String code, String type, String value, long expiresAt) throws IOException {
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
}
