# Lobby 権限設計（2ロール / LuckPerms）

## 目的

- ロビーを安全に運用する
- 管理者だけ Creative で建築・保守できるようにする
- 一般プレイヤーは Adventure で導線だけ利用できるようにする

## 方針

- 権限管理は `LuckPerms` を使用する
- この構成では `admin` と `default` の2ロールを採用する
- 一般プレイヤーは `default` グループで最小権限にする
- 管理者は `admin` グループで運用し、`op` 常用を避ける

## ロール定義

### default

- 対象: 全プレイヤー
- 想定ゲームモード: Adventure
- 権限:
  - `velocity.command.server`（サーバー移動を許可する場合）
  - チャット、移動など通常プレイ最低限

### admin

- 対象: 運営管理者
- 想定ゲームモード: Creative/Spectator（作業時のみ）
- 権限:
  - `*`（検証段階。公開前に絞り込み推奨）
  - 本番前に絞り込み推奨

## 初期セットアップ（最小）

1. `setup/wsl/docker-compose.yml` の `SPIGET_RESOURCES: "28140" # LuckPerms` で `lobby/survival` に自動導入する
2. `CREATE_CONSOLE_IN_PIPE: "true"` を `lobby/survival` に設定する
3. `docker compose -f setup/wsl/docker-compose.yml up -d --force-recreate lobby survival` で再作成する
4. 必要なら一時的に `op` を付与する
5. `make lp-admin PLAYER=<プレイヤーID>` で `admin` を作成し、自分を割り当てる
6. `make deop-lobby PLAYER=<プレイヤーID>` で `op` を外す

## 実行コマンド

```console
# 例: kyoh86 を管理者にする
make op-lobby PLAYER=kyoh86
make lp-admin PLAYER=kyoh86
make deop-lobby PLAYER=kyoh86
```

## 運用ルール

- `op` は緊急時のみ使用し、通常運用は LuckPerms で行う
- 建築作業は `admin` が実施し、作業後は Survival/Adventure に戻す
- 設定変更は本ドキュメントに追記する

## 拡張方針

- 建築担当を分離したくなった場合は `builder` ロールを追加する
- `builder` には `gamemode/tp` など必要最小限の権限のみ付与する

## 次の実装候補

- ロビーに WorldGuard 導入
- `spawn` 領域を保護し、`admin` のみ編集許可
- Join 時に Adventure へ固定するプラグインを追加
