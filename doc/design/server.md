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
    - `proxies.velocity.secret` は `infra/velocity/config/forwarding.secret` と一致させる
    - `velocity` からのみ到達
- `limbo`（認証待機 PicoLimbo）
    - 外部非公開
    - `bind=0.0.0.0:25565`
    - MODERN forwarding (`infra/limbo/config/server.toml`)
    - `velocity` からのみ到達

## 導入プラグイン

`infra/docker-compose.yml` の `MODRINTH_PROJECTS` で以下を導入している。

- `MultiVerse-core`
- `MultiVerse-portals`
- `WorldEdit`
- `WorldGuard`
- `ClickMobs`
    - `infra/world/plugins/dist/ClickMobs/config.yml`
        - `ClickMobs` 本体設定
        - `whitelisted_mobs: [?all]` により全モブを捕獲可能にする

あわせて、以下のローカルプラグインを導入している。

- `ClickMobsRegionGuard`
    - `WorldGuard` のリージョンIDに基づき `ClickMobs` を制御する
    - 本体ファイル: `infra/world/plugins/dist/ClickMobsRegionGuard.jar`
    - 設定: `infra/world/plugins/dist/ClickMobsRegionGuard/config.yml`
        - `allowed_regions.<world>` に許可リージョンIDを列挙する
- `LinkCodeGate`
    - 未認証プレイヤーを `limbo` に隔離し、ワンタイムコードをチャット表示するVelocityプラグイン
    - 本体ファイル: `infra/velocity/plugins/dist/LinkCodeGate.jar`

`world` コンテナは `runtime/world` を `/data` として bind mount し、
起動時に `infra/world/config/bootstrap.sh` で `/config` から設定を反映する。

## 認可管理

認可の判定はローカルプラグイン `LinkCodeGate` が `allowlist.yml` を直接参照して行う。
許可エントリは 認可処理Discord bot `mc-link-bot` （後述）が更新する。

- `runtime/velocity/allowlist.yml`
    - 実運用時の実体

未登録プレイヤーがログインしようとすると、Velocity の `LinkCodeGate` 一時コードを自動発行する。
当該ユーザーを `limbo`（認証待機用 PicoLimbo）へ接続させたうえで、`limbo` 内チャットに一時コードとDiscordでの操作案内を表示する。
NOTE: ワンタイムコードは Redis（`runtime/redis`）に保存される。

`mc-link` コンテナが Discord の `/mc link <code>` を受け取り、`runtime/velocity/allowlist.yml` にエントリを追加します。

## ファイル構成

- `runtime/world`
    - Paper 本体データ
- `runtime/velocity`
    - Velocity 本体データとプラグインデータ
    - `runtime/velocity/allowlist.yml`
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
    - `mc-link` コンテナ（Discord `/mc link` 連携）
    - 各種ローカル / リモートプラグイン の導入
        - `LinkCodeGate` / `LuckPerms` / `Multiverse-Core` / `Multiverse-Portals` / `WorldEdit` / `WorldGuard`
- `infra/velocity/config/velocity.toml`
    - Velocity のルーティング設定
    - `mainhall = "world:25565"` へ転送
- `infra/velocity/config/forwarding.secret`
    - Velocity modern forwarding の共有シークレット
- `infra/world/config/bootstrap.sh`
    - `world` 起動時に `infra/world/plugins/dist/*` と forwarding secret を `/data` へ反映
- `infra/limbo/config/server.toml`
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

## `mc-ctl`

ほとんどの管理作業を自動化するCLIとして `mc-ctl` というコマンドを用意した。
`mc-ctl` は以下のようなプリミティブなサブコマンド構成になっている。

- `mc-ctl asset init`
    - ディレクトリ構成初期化
    - runtimeディレクトリ作成と書き込み可能状態の保証を行う。
- `mc-ctl asset stage`
    - runtime ディレクトリの存在と書込可能状態を確認
- `mc-ctl server up|down|restart|ps|logs velocity|logs world|reload`
    - サーバーの起動、停止、リスタート、状態やログの確認
- `mc-ctl world ensure|regenerate|setup|spawn profile|spawn stage|spawn apply|function run`
    - `mc-ctl world setup` は固定値適用（`setup.commands` と `world.policy.yml`）のみを扱う。
    - 座標依存の反映は `mc-ctl world spawn profile/stage/apply` で行い、ポータル定義などを読み込む。
- `mc-ctl world drop|delete`
- `mc-ctl player op ...|admin ...`
- `mc-ctl link issue --nick <name>|--uuid <uuid> [--ttl 10m]`

`server/world/player` 系でコンソール送信を伴うコマンドは、コンテナが
`running + healthy` になり、`/tmp/minecraft-console-in` パイプが生成されるまで
待機してから実行される。
