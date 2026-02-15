# Minecraft server / WSL 検証構成（Velocity + Paper）

## 概要

このドキュメントは、`WSL2 + Ubuntu` 上で `Velocity + Paper(lobby/survival)` を検証するための手順をまとめたもの。  
ここでの構成は検証用であり、常時運用の本番環境は Linux 実機に移行する前提。

## 構成

- `velocity` : エントリポイント（外部公開ポート `25565`）
- `lobby` : ロビー用 Paper サーバー
- `survival` : サバイバル用 Paper サーバー
- `lobby` には `LuckPerms` と `VeloSend` を自動導入する
- `survival` には `LuckPerms` を自動導入する
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
Datapack `setup/wsl/datapacks/lobby-base` の  
`data/mcserver/function/lobby_settings.mcfunction` に記述し、`/function` で再適用する。

```console
make lobby-apply
```

このコマンドは内部で以下を行う。

- `make lobby-datapack-sync`（datapack を world へ同期）
- `reload`
- `function mcserver:lobby_settings`

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

## Lobby ゲートの再適用

`-8, 63, -2` 付近のゲート演出は  
Datapack `setup/wsl/datapacks/lobby-base` の  
`data/mcserver/function/lobby_gate.mcfunction` に記述し、次で再適用する。

```console
make lobby-gate-apply
```

この適用では以下を実施する。

- 黒曜石フレームを配置
- ゲート内部を紫ガラスで作成
- `area_effect_cloud` でモヤ（`minecraft:portal`）を常駐
- ゲート前に感圧板とコマンドブロックを配置（踏むと `survival` へ転送）

## 感圧板でサーバー移動（最小プラグイン構成）

`VeloSend` を使って、感圧板直下のコマンドブロックから  
プレイヤーを `survival` へ転送する（単独検証向け）。

コマンドブロック例:

```mcfunction
execute in minecraft:overworld run vsend @r survival
```

補足:

- `VeloSend` は `lobby` の `SPIGET_RESOURCES` で自動導入される
- コマンドブロックは `Impulse` + `Needs Redstone`（感圧板トリガー）で運用する
- `vsend` はコンソール直実行だと Paper 1.21.11 で例外化しやすいため、`execute in ... run` でワールド文脈を付与して実行する
- `@r` は「オンライン中の誰か1人」を対象にする。単独検証では踏んだ本人と一致するが、複数同時接続で厳密に踏んだ本人を保証したい場合は専用プラグイン化が必要

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
- `make lobby-datapack-sync` : `setup/wsl/datapacks/lobby-base` を `runtime/lobby/world/datapacks/` へ同期
- `make lobby-apply` : `function mcserver:lobby_settings` を実行
- `make lobby-gate-apply` : `function mcserver:lobby_gate` を実行
