# Minecraft server / WSL 検証構成（Velocity + Paper）

## 概要

このドキュメントは、`WSL2 + Ubuntu` 上で `Velocity + Paper(lobby/survival)` を検証するための手順をまとめたもの。  
ここでの構成は検証用であり、常時運用の本番環境は Linux 実機に移行する前提。

## 構成

- `velocity` : エントリポイント（外部公開ポート `25565`）
- `lobby` : ロビー用 Paper サーバー
- `survival` : サバイバル用 Paper サーバー
- `lobby/survival` には `LuckPerms` を自動導入する
- バックエンドサーバーは外部公開しない

## 前提

- WSL2 で Ubuntu が利用可能
- Docker Desktop + WSL integration 有効、または WSL 側 Docker Engine が利用可能

## 初期化

リポジトリルートで実行する。

```console
./setup/wsl/init.sh
# または
make init
```

これで以下が生成される。

- `setup/wsl/runtime/velocity/velocity.toml`
- `setup/wsl/runtime/velocity/forwarding.secret`（初回のみ安全なランダム値を自動生成）
- `setup/wsl/runtime/lobby/`
- `setup/wsl/runtime/survival/`

`forwarding.secret` は既存ファイルがあれば上書きされない。

## 起動

```console
docker compose -f setup/wsl/docker-compose.yml up -d
# または
make up
```

状態確認:

```console
docker compose -f setup/wsl/docker-compose.yml ps
docker compose -f setup/wsl/docker-compose.yml logs -f velocity
# または
make ps
make logs-velocity
```

## Velocity と Paper の連携設定

`paper-global.yml` は手編集せず、テンプレートとコマンドで反映する。

対象ファイル:

- `setup/wsl/runtime/lobby/config/paper-global.yml`
- `setup/wsl/runtime/survival/config/paper-global.yml`

管理テンプレート:

- `setup/wsl/templates/paper-global.velocity.yml`

反映コマンド:

```console
make configure-paper
```

`make configure-paper` は以下を実施する。

- `proxies.velocity.enabled` をテンプレート値へ反映
- `proxies.velocity.online-mode` をテンプレート値へ反映
- `proxies.velocity.secret` を `setup/wsl/runtime/velocity/forwarding.secret` と同期

反映後、再起動:

```console
docker compose -f setup/wsl/docker-compose.yml restart lobby survival velocity
# または
make restart
```

## 接続確認

- Minecraft クライアントから `localhost:25565` へ接続
- Velocity 経由で lobby へ入れることを確認
- サーバー移動コマンド（例: `/server survival`）で移動確認

## Lobby 内部設定の再適用

ロビーの内部設定（gamerule, time, difficulty, worldspawn）は  
`setup/wsl/lobby-settings.mcfunction` に記述し、コマンドで再適用する。

```console
make lobby-apply
```

初期値は `1.21.11+` の gamerule 名に合わせている。

- `advance_time false`
- `advance_weather false`
- `spawn_mobs false`
- `respawn_radius 0`
- `pvp false`
- `time set noon`
- `difficulty peaceful`
- `weather clear`
- `setworldspawn 0 64 0`

## 検証終了

```console
docker compose -f setup/wsl/docker-compose.yml down
# または
make down
```

データを消して作り直す場合のみ、`setup/wsl/runtime/` を削除して再初期化する。

`forwarding.secret` をローテーションしたい場合は、以下を実行して再生成できる。

```console
rm setup/wsl/runtime/velocity/forwarding.secret
make init
```

再生成後は `lobby/survival` の `paper-global.yml` に新しい値を反映して `make restart` する。

```console
make sync-secret
make restart
```

## Make ターゲット一覧

- `make init` : 検証用ディレクトリとテンプレートを初期化
- `make up` : 検証構成をバックグラウンド起動
- `make down` : 構成を停止
- `make restart` : `velocity/lobby/survival` を再起動
- `make ps` : コンテナ状態の確認
- `make logs` : 全サービスのログ追跡
- `make logs-velocity` : Velocity ログ追跡
- `make logs-lobby` : lobby ログ追跡
- `make logs-survival` : survival ログ追跡
- `make sync-secret` : `forwarding.secret` の値だけを `paper-global.yml` に同期
- `make configure-paper` : テンプレートに基づいて `paper-global.yml` を構成
- `make bootstrap` : `init -> up -> configure-paper -> restart` を一括実行
- `make op-lobby PLAYER=<id>` : lobby で一時的に `op` を付与
- `make deop-lobby PLAYER=<id>` : lobby で `op` を剥奪
- `make lp-admin PLAYER=<id>` : `lobby/survival` で `admin` グループ作成とユーザー割り当て
- `make lobby-apply` : `setup/wsl/lobby-settings.mcfunction` の内容を lobby へ適用
