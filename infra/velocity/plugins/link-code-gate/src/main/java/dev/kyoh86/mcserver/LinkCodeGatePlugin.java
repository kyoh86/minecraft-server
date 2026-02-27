package dev.kyoh86.mcserver;

import com.google.inject.Inject;
import com.velocitypowered.api.event.PostOrder;
import com.velocitypowered.api.event.player.ServerConnectedEvent;
import com.velocitypowered.api.event.player.PlayerChooseInitialServerEvent;
import com.velocitypowered.api.event.player.ServerPreConnectEvent;
import com.velocitypowered.api.event.Subscribe;
import com.velocitypowered.api.plugin.Plugin;
import com.velocitypowered.api.proxy.Player;
import com.velocitypowered.api.proxy.ProxyServer;
import com.velocitypowered.api.proxy.server.RegisteredServer;
import net.kyori.adventure.text.Component;
import net.kyori.adventure.text.format.NamedTextColor;
import org.slf4j.Logger;

import java.io.IOException;
import java.security.SecureRandom;
import java.time.Instant;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.Optional;

@Plugin(
  id = "linkcodegate",
  name = "LinkCodeGate",
  version = "0.1.0",
  description = "Route unlinked players to limbo and issue one-time link codes."
)
public final class LinkCodeGatePlugin {
  private static final long TTL_SECONDS = 10 * 60;
  private static final String MAINHALL_SERVER = "mainhall";
  private static final String LIMBO_SERVER = "limbo";
  private static final long NOTICE_INTERVAL_MILLIS = 3_000;

  private final ProxyServer proxy;
  private final Logger logger;
  private final AllowlistService allowlistService;
  private final LinkCodeStore linkCodeStore;
  private final LinkCodeMessageService linkCodeMessageService;
  private final LinkCodeGenerator linkCodeGenerator;
  private final ConcurrentMap<UUID, Long> lastNoticeAt = new ConcurrentHashMap<>();

  @Inject
  public LinkCodeGatePlugin(ProxyServer proxy, Logger logger) {
    this.proxy = proxy;
    this.logger = logger;
    LinkCodeGateConfig config = LinkCodeGateConfig.load(logger);
    this.allowlistService = new AllowlistService(config.allowlistPath(), logger);
    this.linkCodeStore = new LinkCodeStore(config.redisHost(), config.redisPort(), config.redisDB());
    this.linkCodeMessageService = new LinkCodeMessageService(config.discordGuildName());
    this.linkCodeGenerator = new LinkCodeGenerator(new SecureRandom());
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
    if (now - last < NOTICE_INTERVAL_MILLIS) {
      return;
    }
    lastNoticeAt.put(player.getUniqueId(), now);

    UUID playerUUID = player.getUniqueId();
    String playerName = player.getUsername();
    String code = linkCodeGenerator.generate();
    long expires = Instant.now().getEpochSecond() + TTL_SECONDS;
    String type = "uuid";
    String value = playerUUID.toString();

    proxy.getScheduler().buildTask(this, () -> {
      try {
        linkCodeStore.appendEntry(code, type, value, expires);
      } catch (IOException e) {
        logger.error("failed to write link code for {}", playerName, e);
        proxy.getPlayer(playerUUID).ifPresent(p ->
          p.sendMessage(Component.text("リンクコード発行に失敗しました。しばらくして再接続してください。", NamedTextColor.RED))
        );
        return;
      }
      proxy.getPlayer(playerUUID).ifPresent(p -> linkCodeMessageService.sendLinkMessage(p, code));
      logger.info("issued link code for {} {} {}", playerName, type, value);
    }).schedule();
  }

  private Optional<RegisteredServer> serverByName(String name) {
    return proxy.getServer(name);
  }

  private boolean isAllowed(Player player) {
    return allowlistService.isAllowed(player.getUniqueId());
  }
}
