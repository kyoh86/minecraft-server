# minecraft-server

このリポジトリは、`WSL2/Ubuntu` 上で Minecraft Java サーバー（Paper 1.21.11）を
ローカル検証するための構成を管理する。

現行構成は **単一サーバー（lobby）** のみ。

## 構成

- `setup/wsl/docker-compose.yml`
  - `lobby` コンテナ（`itzg/minecraft-server:java21`）
  - 公開ポート `25565`
  - `LuckPerms` 自動導入（`SPIGET_RESOURCES=28140`）
- `setup/wsl/runtime/lobby`
  - サーバーデータ永続化先（`make init` で作成）
- `setup/wsl/datapacks/lobby-base`
  - lobby 初期設定の Datapack

## 使い方

```console
make init
make up
make ps
make logs-lobby
```

停止:

```console
make down
```

## Lobby 内部設定の再適用

Datapack を同期して `mcserver:lobby_settings` を実行する。

```console
make lobby-apply
```

## 運用補助

- OP 付与: `make op-lobby PLAYER=<player>`
- OP 剥奪: `make deop-lobby PLAYER=<player>`
- LuckPerms admin 付与: `make lp-admin PLAYER=<player>`

## ドキュメント

- WSL 手順: `doc/setup-wsl-velocity.md`
- Ubuntu 手順: `doc/setup-ubuntu.md`
