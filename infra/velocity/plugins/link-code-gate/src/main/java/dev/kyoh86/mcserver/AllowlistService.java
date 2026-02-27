package dev.kyoh86.mcserver;

import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.UUID;
import org.slf4j.Logger;
import org.yaml.snakeyaml.Yaml;

final class AllowlistService {
  private final Path allowlistPath;
  private final Logger logger;

  AllowlistService(Path allowlistPath, Logger logger) {
    this.allowlistPath = allowlistPath;
    this.logger = logger;
  }

  boolean isAllowed(UUID playerUUID) {
    WhitelistEntries entries = loadWhitelistEntries();
    return entries.uuidSet().contains(playerUUID.toString().toLowerCase());
  }

  private WhitelistEntries loadWhitelistEntries() {
    Set<String> uuids = new HashSet<>();
    try (InputStream in = Files.newInputStream(allowlistPath)) {
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
}
