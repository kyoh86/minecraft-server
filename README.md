# minecraft-server

このリポジトリは、`WSL2/Ubuntu` 上で Minecraft Java サーバー（Paper 1.21.11）を
ローカル検証するための構成を管理する。

## 構成

- `setup/wsl/docker-compose.yml`
  - `world` コンテナ（`itzg/minecraft-server:java21`）
  - 公開ポート `25565`
  - `LuckPerms` / `Multiverse-Core` / `Multiverse-Portals` / `WorldEdit` / `WorldGuard` を自動導入
- `setup/wsl/runtime/world`
  - サーバーデータ永続化先
- `setup/wsl/datapacks/world-base`
  - ワールド初期化用 Datapack
- `setup/wsl/worlds/*/world.env.yml`
  - Multiverse 管理ワールド（`residence/resource/factory`）の作成/import用定義
- `setup/wsl/worlds/*/world.policy.yml`
  - ワールド運用ポリシー（`mv modify` で適用）
- `setup/wsl/worlds/*/setup.commands`
  - ワールド初期化コマンド（1行1コマンド）
- `setup/wsl/worlds/*/worldguard.regions.yml`
  - スポーン周辺保護リージョン定義（WorldGuard）

## CLI 方針

`wslctl` は以下のプリミティブコマンドで構成する。

- `setup` : 初期ディレクトリ作成
- `server` : 起動・停止・ログ・reload
- `assets` : Datapack などの配置
- `world` : world create/import・再生成・セットアップ適用
  - `setup.commands` を対象次元で実行
  - `world.policy.yml` にある MV 管理項目も適用
- `player` : op/admin 権限操作

## 典型フロー（初回）

```console
wslctl setup init
wslctl server up
wslctl assets stage
wslctl server reload
wslctl world ensure
wslctl world setup
```

## 典型フロー（設定変更反映）

```console
wslctl assets stage
wslctl server reload
wslctl world setup --world mainhall
```

## 典型フロー（1ワールド再生成）

```console
wslctl world regenerate resource
wslctl world setup --world resource
```

## Makefile

`Makefile` は `wslctl` の薄いショートカット。

- `make setup-init`
- `make server-up`
- `make assets-stage`
- `make world-ensure`
- `make world-regenerate WORLD=resource`
- `make world-drop WORLD=resource`
- `make world-delete WORLD=resource`
- `make world-setup WORLD=mainhall`

## ドキュメント

- WSL 手順: `doc/setup-wsl-velocity.md`
- Ubuntu 手順: `doc/setup-ubuntu.md`
- ワールド再現方針: `doc/world-layout.md`
