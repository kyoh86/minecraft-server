# Minecraft server / WSL 検証構成（単一 Paper: world）

## 概要

このドキュメントは、`WSL2 + Ubuntu` 上で単一 `Paper` サーバーを動かし、
複数ワールド（`mainhall` / `residence` / `resource` / `factory`）を運用する手順を示す。

ここでの方針は、合成コマンドではなくプリミティブな操作を明示的に実行すること。

## 前提

- WSL2 で Ubuntu が利用可能
- Docker Desktop + WSL integration 有効、または WSL 側 Docker Engine が利用可能
- `go` コマンドが利用可能（`cmd/wslctl` 実行に使用）

## コマンド体系

- `wslctl setup init`
- `wslctl server up|down|restart|ps|logs|reload`
- `wslctl assets stage`
- `wslctl world ensure|regenerate|setup|function run`
- `wslctl world drop|delete`
- `wslctl player op ...|admin ...`

`server/world/player` 系でコンソール送信を伴うコマンドは、`world` コンテナが
`running + healthy` になり、`/tmp/minecraft-console-in` パイプが生成されるまで
待機してから実行される。

## 初回セットアップ

```console
wslctl setup init
wslctl server up
wslctl assets stage
wslctl server reload
wslctl world ensure
wslctl world setup
```

## 設定変更の反映

Datapack / mcfunction を更新した場合は次を実行する。

```console
wslctl assets stage
wslctl server reload
wslctl world setup
```

特定ワールドだけセットアップを適用したい場合:

```console
wslctl world setup --world mainhall
```

`mainhall` は `LEVEL` 基底ワールドのため `world.env.yml` は持たず、
`setup/wsl/worlds/mainhall/setup.commands` を読み込んで適用する。
`mainhall` の MV 管理項目は `setup/wsl/worlds/mainhall/world.policy.yml` で管理する。
`setup/wsl/worlds/mainhall/portals.yml` がある場合は runtime の
`plugins/Multiverse-Portals/portals.yml` へ同期する。
`setup/wsl/worlds/<world>/worldguard.regions.yml` がある場合は runtime の
`plugins/WorldGuard/worlds/<world>/regions.yml` へ同期し、`wg reload` を実行する。
ポータル定義反映には `wslctl server restart` が必要。

## ワールド再生成

`deletable: true` のワールドだけ再生成できる。

```console
wslctl world regenerate resource
wslctl world setup --world resource
```

## ワールド drop / delete

```console
wslctl world drop resource
wslctl world delete --yes resource
```

- `drop` は unload + remove だけ実行し、ワールドディスクは残す。
- `delete` は `drop` に加えてワールドディスクを削除する。
- `mainhall` は保護され、`drop`/`delete` できない。
- `delete` は `world.env.yml` の `deletable: true` が必要。
- `world ensure` / `world setup --world mainhall` 実行時は、`mainhall_nether` と `mainhall_the_end` を自動で drop する。
- `world setup --world mainhall` は `world.policy.yml` の `mv_set` を適用し、Hub の運用ポリシーを固定する。

## 任意 function 実行

```console
wslctl world function run mcserver:hello
```

## プレイヤー権限管理

```console
wslctl player op grant kyoh86
wslctl player op revoke kyoh86
wslctl player admin grant kyoh86
wslctl player admin revoke kyoh86
```

## 停止

```console
wslctl server down
```

## Make ターゲット（ショートカット）

- `make setup-init`
- `make server-up`
- `make server-down`
- `make server-restart`
- `make server-ps`
- `make server-logs`
- `make server-reload`
- `make assets-stage`
- `make world-ensure`
- `make world-regenerate WORLD=<name>`
- `make world-drop WORLD=<name>`
- `make world-delete WORLD=<name>`
- `make world-setup [WORLD=<name>]`
- `make world-function FUNCTION=<id>`
- `make player-op-grant PLAYER=<id>`
- `make player-op-revoke PLAYER=<id>`
- `make player-admin-grant PLAYER=<id>`
- `make player-admin-revoke PLAYER=<id>`
