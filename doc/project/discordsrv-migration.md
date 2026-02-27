# DiscordSRV 置換方針（目下の変更）

## 目的

`mc-link` と `LinkCodeGate` に分散している Discord 連携と認可導線をやめ、
DiscordSRV に要件を集約する。

## スコープ

- 対象:
  - DiscordSRV を Discord/Minecraft 連携の主系統にする
  - 既存 `mc-link` / `LinkCodeGate` を即時停止する
  - 通知チャンネル構成を確定する
- 非対象:
  - 旧運用との後方互換
  - 既存 allowlist/link code の移行救済
  - 参加許可ロール判定をサーバー側で実施すること

## 制約

- 後方互換は考慮しない
- `mc-link` / `LinkCodeGate` は即時停止前提
- アプリ利用者の制限は Discord 側ロール設定で行う
- 参加許可ロールをサーバー側で強制しない

## 通知・連携チャンネル

- `#mc-chat`
  - Minecraft <-> Discord チャット連携
- `#mc-member`
  - join/leave 通知
- `#mc-adv`
  - advancement 通知

補足:

- `#playit` は DiscordSRV 対象外のため、別経路（Webhook 等）で運用する。

## 完了条件

- DiscordSRV が上記 3 チャンネル構成で稼働している
- `mc-link` / `LinkCodeGate` が停止され、運用経路から外れている
- 恒久運用に必要な最終仕様を `doc/design/` へ反映できる状態になっている
