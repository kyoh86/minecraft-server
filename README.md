# minecraft-server

このリポジトリは、`WSL2/Ubuntu` 上で Minecraft Java サーバー（Paper 1.21.11）を
ローカル検証するための構成を管理する。

## プレイヤー向け概要

### ワールド構成

- `mainhall`: 移動ハブ（Adventure / Peaceful / PvP無効 / モブ湧き無効）
- `residence`: 生活・拠点建築（Survival / Normal）
- `resource`: 資源採取（Survival / Hard）
- `factory`: 装置・自動化（Survival / Normal）

### 移動

- ワールド移動は `mainhall` のゲート経由
- `residence` / `resource` / `factory` には `mainhall` 戻りゲートあり

### ハブ周辺の保護

- 建築保護エリア（`spawn_protected`）
  - 範囲 `x,z=-32..32`, `y=-64..319`
  - ブロック設置・破壊・爆発破壊を禁止
  - 回路操作の右クリックは許可

### ClickMobs

- 全モブ捕獲を有効化
- 利用可能エリア（`clickmobs_allowed`）
  - `residence` / `resource` / `factory` の `x,z=-64..64`, `y=-64..319`
- `mainhall` では無効

### ワールド境界

- `residence` / `factory`: 中心 `0,0`、直径 `10000`
- `resource`: 境界設定なし

## 構成

- `infra/docker-compose.yml`
  - `world` コンテナ（`itzg/minecraft-server:java21`）
  - 公開ポート `25565`
  - `LuckPerms` / `Multiverse-Core` / `Multiverse-Portals` / `WorldEdit` / `WorldGuard` を自動導入
- `runtime/world`
  - サーバーデータ永続化先
- `datapacks/world-base`
  - ワールド初期化用 Datapack（runtime へそのままコピー）
- `worlds/*/world.env.yml`
  - Multiverse 管理ワールド（`residence/resource/factory`）の作成/import用定義
- `worlds/*/world.policy.yml`
  - ワールド運用ポリシー（`mv modify` で適用）
- `worlds/*/setup.commands`
  - ワールド初期化コマンド（1行1コマンド）
- `worlds/*/worldguard.regions.yml.tmpl`
  - スポーン周辺保護リージョン定義テンプレート（WorldGuard）
- `worlds/mainhall/portals.yml.tmpl`
  - 帰還ポータル定義テンプレート（Multiverse-Portals）

## CLI 方針

`wslctl` は以下のプリミティブコマンドで構成する。

- `setup` : 初期ディレクトリ作成
- `server` : 起動・停止・ログ・reload
- `world` : world create/import・再生成・セットアップ適用
  - `setup.commands` を対象次元で実行
  - `world.policy.yml` にある MV 管理項目を適用
  - `spawn profile/stage/apply` でYプロファイル・テンプレ反映・レイアウト適用
- `player` : op/admin 権限操作

## 典型フロー（初回）

```console
wslctl setup init
wslctl server up
wslctl world ensure
wslctl world setup
wslctl world spawn profile
wslctl world spawn stage
wslctl world spawn apply
```

## 典型フロー（設定変更反映）

```console
wslctl world setup --world mainhall
wslctl world spawn stage
wslctl world spawn apply
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
- `make world-ensure`
- `make world-regenerate WORLD=resource`
- `make world-drop WORLD=resource`
- `make world-delete WORLD=resource`
- `make world-setup WORLD=mainhall`
- `make world-spawn-profile`
- `make world-spawn-stage`
- `make world-spawn-apply`

## ドキュメント

- WSL 手順: `doc/setup-wsl-velocity.md`
- Ubuntu 手順: `doc/setup-ubuntu.md`
- ワールド再現方針: `doc/world-layout.md`
- Codex 運用: `doc/codex-operations.md`
