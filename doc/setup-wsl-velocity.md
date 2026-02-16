# Minecraft server / WSL 検証構成（単一 Paper: world）

## 概要

このドキュメントは、`WSL2 + Ubuntu` 上で単一 `Paper` サーバーを動かし、
ワールド分離（建設用/工業用/資源用など）で運用するための手順をまとめたもの。

ここでの構成は検証用であり、常時運用の本番環境は Linux 実機に移行する前提。

## 構成

- `world` : メイン Paper サーバー（外部公開ポート `25565`）
- `world` には以下を自動導入する
  - `LuckPerms`
  - `Multiverse-Core`
  - `Multiverse-Portals`
- ワールド定義は `setup/wsl/worlds/*/world.env.yml` で管理する

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

- `setup/wsl/runtime/world/`

## 起動

```console
docker compose -f setup/wsl/docker-compose.yml up -d --remove-orphans
# または
make up
```

状態確認:

```console
docker compose -f setup/wsl/docker-compose.yml ps
docker compose -f setup/wsl/docker-compose.yml logs -f world
# または
make ps
make logs-world
```

## ワールド作成と初期化

`make worlds-bootstrap` は以下をまとめて実行する。

- Datapack 同期（`setup/wsl/datapacks/world-base`）
- `reload`
- `world.env.yml` から world 作成/import（Multiverse）
- 各 world の `mcfunction` 初期化実行

現在の定義では `lobby` は `world_type: flat` で作成する。

```console
make worlds-bootstrap
```

1ワールドだけ再生成したい場合:

```console
make world-reset WORLD=resource
```

## World 設定の再適用

内部設定（gamerule, time, difficulty, worldspawn）は
Datapack `setup/wsl/datapacks/world-base` の
`data/mcserver/function/world_settings.mcfunction` に記述し、`/function` で再適用する。

```console
make world-apply
```

このコマンドは内部で以下を行う。

- `make world-datapack-sync`（datapack を world へ同期）
- `reload`
- `function mcserver:world_settings`

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

## Make ターゲット一覧

- `make init` : 検証用ディレクトリを初期化
- `make up` : 構成をバックグラウンド起動（不要サービスは orphan 削除）
- `make down` : 構成を停止
- `make restart` : `world` を再起動
- `make worlds-bootstrap` : `world.env.yml` 定義に従ってワールド作成/importと初期化を実行
- `make world-reset WORLD=<name>` : 指定ワールドを削除して再生成・再初期化（`resettable: true` のみ）
- `make resource-reset` : `make world-reset WORLD=resource` のエイリアス
- `make ps` : コンテナ状態の確認
- `make logs` : 全サービスのログ追跡
- `make logs-world` : world ログ追跡
- `make op-world PLAYER=<id>` : world で一時的に `op` を付与
- `make deop-world PLAYER=<id>` : world で `op` を剥奪
- `make lp-admin PLAYER=<id>` : world で `admin` グループ作成とユーザー割り当て
- `make lp-reset PLAYER=<id>` : world で `admin` の所属を剥奪
- `make world-datapack-sync` : `setup/wsl/datapacks/world-base` を `runtime/world/world/datapacks/` へ同期
- `make world-apply` : `function mcserver:world_settings` を実行
