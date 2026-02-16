# minecraft-server

このリポジトリは、`WSL2/Ubuntu` 上で Minecraft Java サーバー（Paper 1.21.11）を
ローカル検証するための構成を管理する。

現行構成は **単一サーバー（world）** のみ。

## 構成

- `setup/wsl/docker-compose.yml`
  - `world` コンテナ（`itzg/minecraft-server:java21`）
  - 公開ポート `25565`
  - `LuckPerms` / `Multiverse-Core` / `Multiverse-Portals` を自動導入
- `setup/wsl/runtime/world`
  - サーバーデータ永続化先（`make init` で作成）
- `setup/wsl/datapacks/world-base`
  - world初期化用 Datapack
- `setup/wsl/worlds/*/world.env.yml`
  - ワールド作成/import用の定義
  - 例: `lobby` は `world_type: flat`

## 使い方

```console
make init
make up
make ps
make logs-world
```

停止:

```console
make down
```

## ワールド作成と初期化

`world.env.yml` に従って world/lobby/resource/factory を作成/import し、
各ワールドの `init.mcfunction` を実行する。

```console
make worlds-bootstrap
```

1ワールドだけ再生成する:

```console
make world-reset WORLD=resource
```

## World 設定の再適用

Datapack を同期して `mcserver:world_settings` を実行する。

```console
make world-apply
```

## 運用補助

- OP 付与: `make op-world PLAYER=<player>`
- OP 剥奪: `make deop-world PLAYER=<player>`
- LuckPerms admin 付与: `make lp-admin PLAYER=<player>`
- LuckPerms権限リセット（admin 剥奪）: `make lp-reset PLAYER=<player>`

## ドキュメント

- WSL 手順: `doc/setup-wsl-velocity.md`
- Ubuntu 手順: `doc/setup-ubuntu.md`
- ワールド再現方針: `doc/world-layout.md`
