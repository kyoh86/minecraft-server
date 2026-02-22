# Minecraft サーバー構成

## 概要

このプロジェクトにおける各サーバーは以下の5コンテナで動作する。

- `velocity`（公開入口）
    - 公開ポート: `25565`
    - Mojang/Microsoft 認証: `online-mode=true`
    - `player-info-forwarding-mode=modern`
    - `prevent-client-proxy-connections=true`
- `world`（バックエンドPaperサーバー）
    - 外部非公開
    - `online-mode=false`
    - `enforce-secure-profile=false`
    - `proxies.velocity.enabled=true`
    - `proxies.velocity.secret` は `secrets/mc_forwarding_secret.txt` と一致させる
    - `velocity` からのみ到達
- `limbo`（認証待機 PicoLimbo）
    - 外部非公開
    - `bind=0.0.0.0:25565`
    - MODERN forwarding (`runtime/limbo/server.toml`)
    - `velocity` からのみ到達
- `mc-link`（Discord認証受付サーバー）
    - 外部非公開
    - Discord API への外向き接続のみで動作
- `redis`（link-code 一時コード保存）
    - 外部非公開
    - `velocity` / `mc-link` から内部ネットワーク接続のみ許可

## 導入プラグイン

`infra/docker-compose.yml` の `MODRINTH_PROJECTS` で以下を導入している。

- `MultiVerse-core`
- `MultiVerse-portals`
- `WorldEdit`
- `WorldGuard`
- `ClickMobs`
    - `infra/world/plugins/clickmobs/config/config.yml`
        - `ClickMobs` 本体設定
        - `whitelisted_mobs: [?all]` により全モブを捕獲可能にする

あわせて、以下のローカルプラグインを導入している。

- `ClickMobsRegionGuard`
    - `WorldGuard` のリージョンIDに基づき `ClickMobs` を制御する
    - 本体は `infra/world/plugins/clickmobs-region-guard/src` を `infra/world/Dockerfile` の build 時に生成
    - 設定: `infra/world/plugins/clickmobs-region-guard/config/config.yml`
        - `allowed_regions.<world>` に許可リージョンIDを列挙する
- `LinkCodeGate`
    - 未認証プレイヤーを `limbo` に隔離し、ワンタイムコードをチャット表示するVelocityプラグイン
    - 本体は `infra/velocity/plugins/link-code-gate/src` を `infra/velocity/Dockerfile` の build 時に生成

`LuckPerms` は `infra/docker-compose.yml` の `SPIGET_RESOURCES` で導入している。

`world` コンテナは `runtime/world` を `/data` として bind mount し、
起動時に `/config`（composeで bind）から `/data/config` へ設定を同期する。

## 認可管理

認可の判定はローカルプラグイン `LinkCodeGate` が `allowlist.yml` を直接参照して行う。
許可エントリは 認可処理Discord bot `mc-link-bot` （後述）が更新する。
判定キーは UUID のみを使用し、nickname によるフォールバックは行わない。

- `runtime/velocity/allowlist.yml`
    - 実運用時の実体

未登録プレイヤーがログインしようとすると、Velocity の `LinkCodeGate` 一時コードを自動発行する。
当該ユーザーを `limbo`（認証待機用 PicoLimbo）へ接続させたうえで、`limbo` 内チャットに一時コードとDiscordでの操作案内を表示する。
NOTE: ワンタイムコードは Redis（`runtime/redis`）に保存される。

`mc-link` コンテナが Discord の `/mc link <code>` を受け取り、`runtime/velocity/allowlist.yml` にエントリを追加します。
`mc-link` が書き込む bind mount は `allowlist.yml` 単体のみとし、
`runtime/velocity` 全体にはアクセスさせない。

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
- `runtime/limbo/server.toml`
    - `mc-ctl init` が `infra/limbo/config/server.toml` から描画する PicoLimbo 設定
- `infra/docker-compose.yml`
    - 各種サービス定義
    - `world` コンテナ（`itzg/minecraft-server:java25`、内部向け）
    - `limbo` コンテナ（`ghcr.io/quozul/picolimbo@sha256:e031331bda1a3c4aebb7a5222458367ac097a248d6212beca48da047de24199e`、未認証プレイヤー待機用）
    - `velocity` コンテナ（`itzg/mc-proxy:java25`、公開入口 `25565`）
    - `redis` コンテナ（`/mc link` ワンタイムコード保存）
    - `mc-link` コンテナ（Discord `/mc link` 連携）
        - `../runtime/velocity/allowlist.yml` のみを `/allowlist.yml` として書き込みマウントする
    - 各種ローカル / リモートプラグイン の導入
        - `LinkCodeGate` / `LuckPerms` / `Multiverse-Core` / `Multiverse-Portals` / `WorldEdit` / `WorldGuard`
    - healthcheck
        - `redis`: `redis-cli ping`
        - `world`: `mc-health`
        - `velocity`: `pgrep -f velocity`
        - `mc-link`: `pgrep -f mc-link-bot`
        - `limbo`: `pico_limbo --help`
- `infra/limbo/config/server.toml`
    - PicoLimbo 設定テンプレート
    - `mc-ctl init` が `secrets/mc_forwarding_secret.txt` を埋め込んで `runtime/limbo/server.toml` を生成する
- `infra/velocity/Dockerfile`
    - Velocity用カスタムイメージ定義
    - `infra/velocity/plugins/link-code-gate/src` を Maven でビルドし、生成JARを `/plugins/LinkCodeGate.jar` へ同梱する
- `infra/velocity/config/velocity.toml`
    - Velocity のルーティング設定
    - `mainhall = "world:25565"` へ転送
    - `secrets/mc_forwarding_secret.txt`
        - Velocity modern forwarding の共有シークレット
        - `mc-ctl init` がユーザーの入力としてここに保存する
- `infra/velocity/plugins/link-code-gate`
    - Velocity用ローカルプラグイン `LinkCodeGate` の管理ディレクトリ
    - `src`: プラグイン実装（Mavenプロジェクト）
    - `dist`: 配布設定ファイル置き場（現状未使用）
- `infra/mc-link/Dockerfile`
    - Discord連携Bot `mc-link-bot` のマルチステージビルド定義
    - Goバイナリをビルドし、最小ランタイムイメージへ配置する
- `infra/world/Dockerfile`
    - world用カスタムイメージ定義
    - `infra/world/plugins/clickmobs-region-guard/src` を Maven でビルドし、生成JARを `/plugins/ClickMobsRegionGuard.jar` へ同梱する
    - `infra/world/plugins/clickmobs/config/config.yml` と
      `infra/world/plugins/clickmobs-region-guard/config/config.yml` を同梱する
- `infra/world/config/paper-global.yml.tmpl`
    - Paper 用 `paper-global.yml` のテンプレート
    - `mc-ctl init` が forwarding secret を埋め込んで `secrets/world/paper-global.yml` を生成する
- `secrets/world/paper-global.yml`
    - world 用の生成済み設定（secret 含む）
    - compose で `/config/paper-global.yml` に bind し、起動時に `/data/config` へ同期される
- `infra/world/plugins/clickmobs-region-guard`
    - world用ローカルプラグイン `ClickMobsRegionGuard` のビルド環境
    - `src`: プラグイン実装（Mavenプロジェクト）
    - `config`: プラグイン設定ファイル
- `infra/world/plugins/clickmobs`
    - `ClickMobs` 設定ファイルの管理ディレクトリ
    - `config/config.yml` を world イメージへ同梱する
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
    - `worlds/*/worldguard.regions.yml`
        - スポーン周辺保護リージョン定義（WorldGuard）
    - `worlds/mainhall/portals.yml.tmpl`
        - 帰還ポータル定義テンプレート（Multiverse-Portals）

## `mc-ctl`

ほとんどの管理作業を自動化するCLIとして `mc-ctl` というコマンドを用意した。
`mc-ctl` は以下のようなプリミティブなサブコマンド構成になっている。

- `mc-ctl init`
    - runtime ディレクトリ初期化
    - secrets 設定（対話入力。未入力時は既定値で補完）
    - `runtime/limbo/server.toml` 描画
- `mc-ctl server up|down|restart|ps|logs velocity|logs world|reload`
    - サーバーの起動、停止、リスタート、状態やログの確認
- `mc-ctl world ensure|regenerate|setup|spawn profile|spawn stage|spawn apply|function run`
    - `mc-ctl world setup` は固定値適用（`setup.commands` と `world.policy.yml`）のみを扱う。
    - 座標依存の反映は `mc-ctl world spawn profile/stage/apply` で行い、ポータル定義などを読み込む。
- `mc-ctl world drop|delete`
- `mc-ctl player op ...|admin ...`

`server/world/player` 系でコンソール送信を伴うコマンドは、コンテナが
`running + healthy` になり、`/tmp/minecraft-console-in` パイプが生成されるまで
待機してから実行される。
