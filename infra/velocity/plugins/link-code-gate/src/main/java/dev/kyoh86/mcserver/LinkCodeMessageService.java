package dev.kyoh86.mcserver;

import com.velocitypowered.api.proxy.Player;
import net.kyori.adventure.text.Component;
import net.kyori.adventure.text.event.ClickEvent;
import net.kyori.adventure.text.event.HoverEvent;
import net.kyori.adventure.text.format.NamedTextColor;

final class LinkCodeMessageService {
  private final String discordGuildName;

  LinkCodeMessageService(String discordGuildName) {
    this.discordGuildName = discordGuildName;
  }

  void sendLinkMessage(Player player, String code) {
    String cmd = "/mc link code:" + code;
    String guide = discordGuildName.isEmpty()
      ? "コピーしたコマンドをDiscordで送信してください"
      : "コピーしたコマンドをDiscord「" + discordGuildName + "」で送信してください";
    player.sendMessage(
      Component.text("クリックしてコマンドをコピーしてください: ", NamedTextColor.GRAY)
        .append(Component.newline())
        .append(
          Component.text(" [" + cmd + "]", NamedTextColor.WHITE)
            .clickEvent(ClickEvent.copyToClipboard(cmd))
            .hoverEvent(HoverEvent.showText(Component.text("クリックでコマンドをコピー")))
        )
        .append(Component.newline())
        .append(Component.text(guide, NamedTextColor.GRAY))
    );
  }
}
