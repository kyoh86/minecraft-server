# Minecraft サーバー構成

## 概要

このプロジェクトにおける各サーバーは以下の3層で動作する。

- `velocity`（公開入口）
    - 公開ポート: `25565`
    - Mojang/Microsoft 認証: `online-mode=true`
    - `player-info-forwarding-mode=modern`
- `world`（バックエンドPaperサーバー）
    - 外部非公開
    - `online-mode=false`
    - `enforce-secure-profile=false`
    - `proxies.velocity.enabled=true`
    - `proxies.velocity.secret` は `infra/velocity/forwarding.secret` と一致させる
    - `velocity` からのみ到達
- `limbo`（認証待機 PicoLimbo）
    - 外部非公開
    - `bind=0.0.0.0:25565`
    - MODERN forwarding (`infra/pico-limbo/server.toml`)
    - `velocity` からのみ到達

## 導入プラグイン

`infra/docker-compose.yml` の `MODRINTH_PROJECTS` で以下を導入している。

- `MultiVerse-core`
- `MultiVerse-portals`
- `WorldEdit`
- `WorldGuard`
- `ClickMobs`
    - `infra/plugins/ClickMobs/config.yml`
        - `ClickMobs` 本体設定
        - `whitelisted_mobs: [?all]` により全モブを捕獲可能にする

あわせて、以下のローカルプラグインを導入している。

- `ClickMobsRegionGuard`
    - `WorldGuard` のリージョンIDに基づき `ClickMobs` を制御する
    - 本体ファイル: `infra/plugins/ClickMobsRegionGuard.jar`
    - 設定: `infra/plugins/ClickMobsRegionGuard/config.yml`
        - `allowed_regions.<world>` に許可リージョンIDを列挙する
- `LinkCodeGate`
    - 未認証プレイヤーを `limbo` に隔離し、ワンタイムコードをチャット表示するVelocityプラグイン
    - 本体ファイル: `infra/plugins/LinkCodeGate.jar`

## 認可管理

認可の判定はローカルプラグイン `LinkCodeGate` が `allowlist.yml` を直接参照して行う。
許可エントリは 認可処理Discord bot `mclink` （後述）が更新する。

- `infra/velocity/allowlist.yml`
    - 初期テンプレート
- `runtime/velocity/.wslctl/allowlist.yml`
    - 実運用時の実体

未登録プレイヤーがログインしようとすると、Velocity の `LinkCodeGate` 一時コードを自動発行する。
当該ユーザーを `limbo`（認証待機用 PicoLimbo）へ接続させたうえで、`limbo` 内チャットに一時コードとDiscordでの操作案内を表示する。
NOTE: ワンタイムコードは Redis（`runtime/redis`）に保存される。

`mclink` コンテナが Discord の `/mc link <code>` を受け取り、`runtime/velocity/.wslctl/allowlist.yml` にエントリを追加します。

## ファイル構成

- `runtime/world`
    - Paper 本体データ
- `runtime/velocity`
    - Velocity 本体データとプラグインデータ
    - `runtime/velocity/.wslctl/allowlist.yml`
        - 認可リスト
- `runtime/redis`
    - Redis データ
    - `/mc link` ワンタイムコードの保存先として利用
- `infra/docker-compose.yml`
    - 各種サービス定義
    - `world` コンテナ（`itzg/minecraft-server:java21`、内部向け）
    - `limbo` コンテナ（`ghcr.io/quozul/picolimbo:latest`、未認証プレイヤー待機用）
    - `velocity` コンテナ（`itzg/mc-proxy:java21`、公開入口 `25565`）
    - `redis` コンテナ（`/mc link` ワンタイムコード保存）
    - `mclink` コンテナ（Discord `/mc link` 連携）
    - 各種ローカル / リモートプラグイン の導入
        - `LinkCodeGate` / `LuckPerms` / `Multiverse-Core` / `Multiverse-Portals` / `WorldEdit` / `WorldGuard`
- `infra/velocity/velocity.toml`
    - Velocity のルーティング設定
    - `mainhall = "world:25565"` へ転送
- `infra/velocity/forwarding.secret`
    - Velocity modern forwarding の共有シークレット
- `infra/world-patches/paper-velocity-forwarding.json`
    - `paper-global.yml` へ Velocity forwarding 設定を起動時に適用
- `infra/pico-limbo/server.toml`
    - PicoLimbo 本体の待機サーバー設定
- `datapacks/world-base`
    - ワールド初期化用 Datapack（runtime へそのままコピー）
- `worlds`
    - Multiverse 管理ワールドの各種定義
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

## `wslctl`

ほとんどの管理作業を自動化するCLIとして `wslctl` というコマンドを用意した。
`wslctl` は以下のようなプリミティブなサブコマンド構成になっている。

- `wslctl setup init`
    - ディレクトリ構成初期化
    - runtimeディレクトリ作成と書き込み可能状態の保証を行う。
- `wslctl server up|down|restart|ps|logs velocity|logs world|reload`
    - サーバーの起動、停止、リスタート、状態やログの確認
- `wslctl world ensure|regenerate|setup|spawn profile|spawn stage|spawn apply|function run`
    - `wslctl world setup` は固定値適用（`setup.commands` と `world.policy.yml`）のみを扱う。
    - 座標依存の反映は `wslctl world spawn profile/stage/apply` で行い、ポータル定義などを読み込む。
- `wslctl world drop|delete`
- `wslctl player op ...|admin ...`
- `wslctl link issue --nick <name>|--uuid <uuid> [--ttl 10m]`

`server/world/player` 系でコンソール送信を伴うコマンドは、コンテナが
`running + healthy` になり、`/tmp/minecraft-console-in` パイプが生成されるまで
待機してから実行される。
