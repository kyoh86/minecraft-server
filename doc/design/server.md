# Minecraft サーバー構成

## 概要

このプロジェクトにおける各サーバーは以下の7コンテナで動作する。

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
    - `${LOCAL_UID}:${LOCAL_GID}` で実行し、`secrets/limbo/server.toml` を読み込む
    - `bind=0.0.0.0:25565`
    - MODERN forwarding (`secrets/limbo/server.toml`)
    - `velocity` からのみ到達
- `mc-link`（Discord認証受付サーバー）
    - 外部非公開
    - Discord API への外向き接続のみで動作
    - `${LOCAL_UID}:${LOCAL_GID}` で実行し、`/run/secrets/mc_link_discord` を読取可能にする
- `redis`（link-code 一時コード保存）
    - 外部非公開
    - `velocity` / `mc-link` から内部ネットワーク接続のみ許可
- `ngrok`（ngrok トンネルエージェント）
    - 外部非公開
    - `velocity` と同一の内部ネットワーク上で動作
    - ngrok への外向き接続のみを使って `velocity:25565` を公開
- `ngrok-log-notifier`（Vector ログ監視）
    - 外部非公開
    - Docker socket から `mc-ngrok` のログを購読
    - `ngrok` URL を抽出して Discord Webhook へ通知
    - 起動順は `ngrok-log-notifier` を先行し、`ngrok` は `velocity` と `ngrok-log-notifier` の `service_healthy` を待って起動する
- `member-log-notifier`（Vector ログ監視）
    - 外部非公開
    - Docker socket から `mc-world` のログを購読
    - `joined the game` / `left the game` を抽出して Discord Webhook へ通知
    - `world` への `depends_on` は持たず、単独で常時待機する
- `health-heartbeat`（Healthchecks heartbeat）
    - 外部非公開
    - Docker socket から `mc-world` / `mc-velocity` / `mc-ngrok` の状態を確認する
    - `running` および（healthcheck 定義時）`healthy` を確認し、正常時に Healthchecks.io ping を送信する
    - 起動直後は対象コンテナが `running/healthy` になるまで待機し、初期化中の `fail` は送信しない
    - 異常時は `.../fail` を送信し、復旧後は通常 ping に戻す

## 導入プラグイン

`infra/docker-compose.yml` の `MODRINTH_PROJECTS` で以下を導入している。

- `MultiVerse-core`
- `MultiVerse-portals`
- `FastAsyncWorldEdit`
- `WorldGuard`
- `ClickMobs`
    - `infra/world/plugins/clickmobs/config/config.yml`
        - `ClickMobs` 本体設定
        - `whitelisted_mobs: [?all]` により全モブを捕獲可能にする

あわせて、以下のローカルプラグインを導入している。

- `ClickMobsRegionGuard`
    - `WorldGuard` のリージョンIDに基づき `ClickMobs` を制御する
    - `allowed_region_ids` に列挙したリージョン内のみ `ClickMobs` 操作を許可する
    - 実装責務は以下に分割する
        - `ClickMobsRegionGuardPlugin`: イベント配線と依存初期化
        - `ClickMobsGuardConfig`: 設定読込と既定値補完
        - `RegionAccessService`: WorldGuardリージョン判定
        - `ClickMobsPermissionService`: `clickmobs.pickup` / `clickmobs.place` の付与管理
        - `ClickMobsActionDetector`: ClickMobsアイテム/操作判定
    - 本体は `infra/world/plugins/clickmobs-region-guard/src` を `infra/world/Dockerfile` の build 時に生成
    - 設定: `infra/world/plugins/clickmobs-region-guard/config/config.yml`
        - `allowed_region_ids` に許可リージョンIDを列挙する
- `RegionStatusUI`
    - `WorldGuard` のリージョン状態を bossbar 表示する
    - `spawn_protected_region_id` と `allowed_region_ids` の一致状態で表示を切り替える
    - 本体は `infra/world/plugins/region-status-ui/src` を `infra/world/Dockerfile` の build 時に生成
    - 設定: `infra/world/plugins/region-status-ui/config/config.yml`
        - `allowed_region_ids` に `ClickMobs` 許可リージョンIDを列挙する
        - `status_bossbar` で表示文言・色・リージョンIDを設定する
- `SpawnSafetyGuard`
    - `mainhall` の `spawn_safe` 範囲外にいるプレイヤーをスポーン地点へ退避させる
    - join / world change 後に複数tickで再確認し、遅延テレポートの競合を吸収する
    - 本体は `infra/world/plugins/spawn-safety-guard/src` を `infra/world/Dockerfile` の build 時に生成
    - 設定: `infra/world/plugins/spawn-safety-guard/config/config.yml`
        - `login_safety` で有効化・対象ワールド・通知文言を設定する
        - `spawn_safe` で安全範囲座標を設定する
- `LinkCodeGate`
    - 未認証プレイヤーを `limbo` に隔離し、ワンタイムコードをチャット表示するVelocityプラグイン
    - チャット案内は1行のみ表示し、`LINK CODE` と `/mc link code:XXXX` の両方を
      クリックコピー可能にする
    - `secrets/mc_link_discord_guild_name.txt`（または `MC_LINK_DISCORD_GUILD_NAME`）を参照し、
      値がある場合は案内文へ Discord サーバー名を埋め込む
    - `runtime/allowlist/allowlist.yml` は `SnakeYAML` で読み取り、`uuids` 配列を正規パースする
    - `allowlist.yml` の読込/パースに失敗した場合は起動時にエラーで停止する
    - Redis へのワンタイムコード書き込みは `Jedis` クライアントで実行する
    - `MC_LINK_REDIS_ADDR` / `MC_LINK_REDIS_DB` の不正値は既定値へ丸めず、起動時にエラーで停止する
    - 実装責務は以下に分割する
        - `LinkCodeGatePlugin`: イベント配線と非同期発行フロー制御
        - `LinkCodeGateConfig`: 環境変数/secretの解決
        - `AllowlistService`: allowlist.yml の読込とUUID許可判定
        - `LinkCodeStore`: Redis へのコード保存
        - `LinkCodeGenerator`: 一時コード生成
        - `LinkCodeMessageService`: チャット案内メッセージ生成
    - 本体は `infra/velocity/plugins/link-code-gate/src` を `infra/velocity/Dockerfile` の build 時に生成
- `HubTerraform`
    - ワールドHub周辺の整地を自動化するプラグイン
    - `spawn apply` が `hubterraform apply <world> <surfaceY>` を実行して適用する
    - 本体は `infra/world/plugins/hub-terraform` を `infra/world/Dockerfile` の build 時に生成

`LuckPerms` は `infra/docker-compose.yml` の `SPIGET_RESOURCES` で導入している。

`world` コンテナは `runtime/world` を `/data` として bind mount し、
起動時に `/config`（composeで bind）から `/data/config` へ設定を同期する。

## 認可管理

認可の判定はローカルプラグイン `LinkCodeGate` が `allowlist.yml` を直接参照して行う。
許可エントリは 認可処理Discord bot `mc-link-bot` （後述）が更新する。
判定キーは UUID のみを使用し、nickname によるフォールバックは行わない。
`mc-link-bot` の `/mc link` 実行者は `secrets/mc_link_discord.toml` の
`allowed_role_ids` で制限できる（空配列なら制限なし）。

- `runtime/allowlist/allowlist.yml`
    - 実運用時の実体

未登録プレイヤーがログインしようとすると、Velocity の `LinkCodeGate` 一時コードを自動発行する。
当該ユーザーを `limbo`（認証待機用 PicoLimbo）へ接続させたうえで、`limbo` 内チャットに一時コードとDiscordでの操作案内を表示する。
NOTE: ワンタイムコードは Redis（`runtime/redis`）に保存される。
`LinkCodeGate` の Redis 書き込みは接続イベント本体から分離して非同期実行し、
接続/読取タイムアウトを設定する。

`mc-link` コンテナが Discord の `/mc link <code>` を受け取り、`runtime/allowlist/allowlist.yml` にエントリを追加します。
`mc-link` は `runtime/allowlist` を `/allowlist` へ bind mount し、
`/allowlist/allowlist.yml` を更新する。
allowlist 更新時は Redis ロックを取得し、`allowlist.yml.tmp` への出力後に `rename` で置換する。
コード消費は Redis 上で原子的に確定し、同一コードの多重利用を防ぐ。
allowlist 更新に失敗した場合は、同一ユーザーによる当該 claim を巻き戻して再試行可能にする。
`LinkCodeGate` の通常ログにはワンタイムコード値を出力しない。

## ファイル構成

- `runtime/world`
    - Paper 本体データ
- `runtime/velocity`
    - Velocity 本体データとプラグインデータ
- `runtime/allowlist`
    - 認可リスト保存ディレクトリ
    - `runtime/allowlist/allowlist.yml`
        - 認可リスト
    - 旧パス `runtime/velocity/allowlist.yml` との互換コピーは行わない
- `runtime/redis`
    - Redis データ
    - `/mc link` ワンタイムコードの保存先として利用
- `runtime/ngrok`
    - ngrok クライアント設定の保存先
- `infra/.env`
    - `mc-ctl init` が補完する compose 変数ファイル
    - `LOCAL_UID` / `LOCAL_GID` を保持し、compose の標準 `.env` 読込で使用する
- `secrets/limbo/server.toml`
    - `mc-ctl init` が `infra/limbo/config/server.toml.tmpl` から描画する PicoLimbo 設定
- `secrets/mc_link_discord.toml`
    - `mc-link-bot` 用 secret
    - `bot_token` / `guild_id` / `allowed_role_ids` を保持する
- `secrets/mc_link_discord_guild_name.txt`
    - Discord サーバー表示名
    - `velocity` の LinkCodeGate 案内文に利用する
- `secrets/ngrok_auth_token.txt`
    - ngrok 接続用 Authtoken
- `secrets/ngrok_discord_webhook_url.txt`
    - ngrok URL 通知用 Discord Webhook URL
- `secrets/member_discord_webhook_url.txt`
    - join/leave 通知用 Discord Webhook URL
- `secrets/healthchecks_heartbeat_url.txt`
    - Healthchecks.io heartbeat 送信用 ping URL
- `infra/ngrok-log-notifier/vector.toml`
    - `mc-ngrok` ログの監視設定
    - `tcp://...` URL 抽出・重複抑止・Discord Webhook POST を定義
- `infra/member-log-notifier/vector.toml`
    - `mc-world` ログの監視設定
    - `joined the game` / `left the game` 抽出・Discord Webhook POST を定義
- `infra/docker-compose.yml`
    - 各種サービス定義
    - `world` コンテナ（`itzg/minecraft-server:java25`、内部向け）
    - `limbo` コンテナ（`ghcr.io/quozul/picolimbo@sha256:e031331bda1a3c4aebb7a5222458367ac097a248d6212beca48da047de24199e`、未認証プレイヤー待機用）
    - `velocity` コンテナ（`itzg/mc-proxy:java25`、公開入口 `25565`）
    - `redis` コンテナ（`/mc link` ワンタイムコード保存）
    - `mc-link` コンテナ（Discord `/mc link` 連携）
        - `mc_link_discord.toml` を Docker secrets 経由で `/run/secrets/mc_link_discord` に注入する
        - `../runtime/allowlist` を `/allowlist` として書き込みマウントし、`/allowlist/allowlist.yml` を更新する
    - `ngrok` コンテナ（ngrok トンネル）
        - `runtime/ngrok` を `/home/ngrok/.config/ngrok` へ bind し、設定を保持する
        - `secrets/ngrok_auth_token.txt` を `/run/ngrok_auth_token.txt` に read-only bind mount する
        - `ngrok tcp velocity:25565` で公開トンネルを起動する
    - `ngrok-log-notifier` コンテナ（Vector）
        - `secrets/ngrok_discord_webhook_url.txt` を `/run/ngrok_discord_webhook_url.txt` に read-only bind mount する
        - `/var/run/docker.sock` を read-write mount し、`docker_logs` source で `mc-ngrok` ログを購読する
        - `infra/ngrok-log-notifier/vector.toml` を `/etc/vector/vector.toml` として read-only mount する
        - `tcp://...` URL を抽出し、重複URLを抑止した上で Discord Webhook へ通知する
        - `ngrok` への `depends_on` は持たず、単独で常時待機する
    - `member-log-notifier` コンテナ（Vector）
        - `secrets/member_discord_webhook_url.txt` を `/run/member_discord_webhook_url.txt` に read-only bind mount する
        - `/var/run/docker.sock` を read-write mount し、`docker_logs` source で `mc-world` ログを購読する
        - `infra/member-log-notifier/vector.toml` を `/etc/vector/vector.toml` として read-only mount する
        - `joined the game` / `left the game` を抽出し、Discord Webhook へ通知する
    - `health-heartbeat` コンテナ
        - `secrets/healthchecks_heartbeat_url.txt` を `/run/healthchecks_heartbeat_url.txt` に read-only bind mount する
        - `/var/run/docker.sock` を read-write mount し、`docker inspect` で対象コンテナ状態を確認する
        - `infra/health-heartbeat/heartbeat.sh` を実行し、60秒間隔で Healthchecks.io に heartbeat を送信する
    - 各種ローカル / リモートプラグイン の導入
        - `LinkCodeGate` / `LuckPerms` / `Multiverse-Core` / `Multiverse-Portals` / `FastAsyncWorldEdit` / `WorldGuard` / `HubTerraform`
    - healthcheck
        - `redis`: `redis-cli ping`
        - `world`: `mc-health`
        - `velocity`: `pgrep -f velocity`
        - `mc-link`: `pgrep -f mc-link-bot`
        - `limbo`: `pico_limbo --help`
- `infra/limbo/config/server.toml.tmpl`
    - PicoLimbo 設定テンプレート
    - `mc-ctl init` が `secrets/mc_forwarding_secret.txt` を埋め込んで `secrets/limbo/server.toml` を生成する
    - `server_list` を定義し、サーバー一覧応答（MOTD/最大人数/アイコン/オンライン人数表示）を固定する
    - `world` を定義し、スポーン座標・回転・次元・時刻を固定する
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
    - `infra/world/plugins/region-status-ui/src` を Maven でビルドし、生成JARを `/plugins/RegionStatusUI.jar` へ同梱する
    - `infra/world/plugins/spawn-safety-guard/src` を Maven でビルドし、生成JARを `/plugins/SpawnSafetyGuard.jar` へ同梱する
    - `infra/world/plugins/clickmobs/config/config.yml` と
      `infra/world/plugins/clickmobs-region-guard/config/config.yml` を同梱する
    - `infra/world/plugins/region-status-ui/config/config.yml` と
      `infra/world/plugins/spawn-safety-guard/config/config.yml` を同梱する
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
- `infra/world/plugins/region-status-ui`
    - world用ローカルプラグイン `RegionStatusUI` のビルド環境
    - `src`: プラグイン実装（Mavenプロジェクト）
    - `config`: プラグイン設定ファイル
- `infra/world/plugins/spawn-safety-guard`
    - world用ローカルプラグイン `SpawnSafetyGuard` のビルド環境
    - `src`: プラグイン実装（Mavenプロジェクト）
    - `config`: プラグイン設定ファイル
- `infra/world/plugins/clickmobs`
    - `ClickMobs` 設定ファイルの管理ディレクトリ
    - `config/config.yml` を world イメージへ同梱する
- `infra/world/schematics/hub.schem`
    - `mainhall` 以外のワールドへ貼り付ける Hub 建築スキーマ
    - `mc-ctl spawn stage` / `mc-ctl spawn apply` が runtime へ同期する
- `datapacks/world-base`
    - ワールド初期化用 Datapack（runtime へそのままコピー）
- `worlds`
    - Multiverse 管理ワールドの各種定義
    - `worlds/*/config.toml`
        - ワールド作成/import用定義と運用ポリシー（`mv modify`）の統合定義
    - `worlds/*/setup.commands`
        - ワールド初期化コマンド（1行1コマンド）
    - `worlds/mainhall/worldguard.regions.yml.tmpl`
    - `worlds/_defaults/worldguard.regions.yml.tmpl`
        - スポーン周辺保護リージョン定義（WorldGuard）
    - `worlds/mainhall/portals.yml.tmpl`
        - 帰還ポータル定義テンプレート（Multiverse-Portals）

## `mc-ctl`

ほとんどの管理作業を自動化するCLIとして `mc-ctl` というコマンドを用意した。
`mc-ctl` は以下のようなプリミティブなサブコマンド構成になっている。

- `mc-ctl init`
    - runtime ディレクトリ初期化
    - secrets 設定（対話入力。未入力時は既定値で補完）
    - `secrets/limbo/server.toml` 描画
- `mc-ctl server up|down|restart|ps|logs velocity|logs world|reload`
    - サーバーの起動、停止、リスタート、状態やログの確認
    - `mc-ctl server restart <service> --build` で image 再ビルド + 再作成を実行できる
- `mc-ctl world ensure|regenerate|setup|function run`
    - `mc-ctl world setup` は固定値適用（`setup.commands` と `config.toml`）のみを扱う。
    - 座標依存の反映は `mc-ctl spawn profile/stage/apply` で行い、ポータル定義などを読み込む。
- `mc-ctl spawn profile|stage|apply`
    - スポーン基準のプロファイル取得、テンプレート描画、反映を行う。
- `mc-ctl world drop|delete`
- `mc-ctl player op ...|admin ...|delink`

`server/world/player` 系でコンソール送信を伴うコマンドは、コンテナが
`running + healthy` になり、`/tmp/minecraft-console-in` パイプが生成されるまで
待機してから実行される。
