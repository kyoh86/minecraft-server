# Minecraft サーバー構成（Paper + Velocity）

## 概要

このプロジェクトは以下の 2 層で動作する。

- `velocity`（公開入口）
  - 公開ポート: `25565`
  - Mojang/Microsoft 認証: `online-mode=true`
  - `player-info-forwarding-mode=modern`
- `world`（バックエンド Paper）
  - 外部非公開
  - `online-mode=false`
  - `enforce-secure-profile=false`
  - `proxies.velocity.enabled=true`
  - `proxies.velocity.secret` は `infra/velocity/forwarding.secret` と一致させる
  - `velocity` からのみ到達

Bot 検証時は、Bot を `world` 側へ直接接続できる。

## 導入プラグイン

`infra/docker-compose.yml` の `MODRINTH_PROJECTS` で以下を導入する。

- `multiverse-core`
- `multiverse-portals`
- `worldedit`
- `worldguard`
- `clickmobs`

あわせて、ローカル配布プラグインを以下で導入する。

- `infra/plugins/ClickMobs/config.yml`
  - `ClickMobs` 本体設定
  - `whitelisted_mobs: [?all]` により全モブを捕獲可能にする
- `infra/plugins/ClickMobsRegionGuard.jar`
  - ワールドガードのリージョンIDに基づき `ClickMobs` を制御する
- `infra/plugins/ClickMobsRegionGuard/config.yml`
  - `allowed_regions.<world>` に許可リージョンIDを列挙する

## コマンド体系

コマンドは以下の構成になっている。

- `wslctl setup init`
    - ディレクトリ構成初期化
    - runtimeディレクトリ作成と書き込み可能状態の保証を行う。
- `wslctl server up|down|restart|ps|logs velocity|logs world|reload`
    - サーバーの起動、停止、リスタート、状態やログの確認
- `wslctl world ensure|regenerate|setup|spawn profile|spawn stage|spawn apply|function run`
- `wslctl world drop|delete`
- `wslctl player op ...|admin ...`

`server/world/player` 系でコンソール送信を伴うコマンドは、コンテナが
`running + healthy` になり、`/tmp/minecraft-console-in` パイプが生成されるまで
待機してから実行される。

## 初回セットアップ

```console
wslctl setup init
wslctl server up
wslctl world ensure
wslctl world setup
wslctl world spawn profile
wslctl world spawn stage
wslctl world spawn apply
```

## 設定変更の反映

設定変更を反映する場合は次を実行する。

```console
wslctl world setup
wslctl world spawn stage
wslctl world spawn apply
```

特定ワールドだけセットアップを適用したい場合:

```console
wslctl world setup --world mainhall
```

`mainhall` は `LEVEL` 基底ワールドのため `world.env.yml` は持たず、
`worlds/mainhall/setup.commands` を読み込んで適用する。
`mainhall` の MV 管理項目は `worlds/mainhall/world.policy.yml` で管理する。
`wslctl world setup` は固定値適用（`setup.commands` と `world.policy.yml`）のみを扱う。
座標依存の反映は `wslctl world spawn profile/stage/apply` で行う。
テンプレートは `worlds/mainhall/portals.yml.tmpl` と
`worlds/<world>/worldguard.regions.yml.tmpl` を使用する。

## ワールド再生成

`deletable: true` のワールドだけ再生成できる。

```console
wslctl world regenerate resource
wslctl world setup --world resource
wslctl world spawn profile
wslctl world spawn stage
wslctl world spawn apply
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
- `make world-ensure`
- `make world-regenerate WORLD=<name>`
- `make world-drop WORLD=<name>`
- `make world-delete WORLD=<name>`
- `make world-setup [WORLD=<name>]`
- `make world-spawn-profile`
- `make world-spawn-stage`
- `make world-spawn-apply`
- `make world-function FUNCTION=<id>`
- `make player-op-grant PLAYER=<id>`
- `make player-op-revoke PLAYER=<id>`
- `make player-admin-grant PLAYER=<id>`
- `make player-admin-revoke PLAYER=<id>`

## ファイル構成

- `runtime/world`: Paper 本体データ
- `runtime/velocity`: Velocity 本体データとプラグインデータ
- `infra/docker-compose.yml`
  - `world` / `velocity` サービス定義
- `infra/velocity/velocity.toml`
  - Velocity のルーティング設定
  - `mainhall = "world:25565"` へ転送
- `infra/velocity/forwarding.secret`
  - Velocity modern forwarding の共有シークレット
- `infra/world-patches/paper-velocity-forwarding.json`
  - `paper-global.yml` へ Velocity forwarding 設定を起動時に適用
