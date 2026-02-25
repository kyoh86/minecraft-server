# Minecraft サーバー管理マニュアル

このプロジェクトで構築されるMinecraftサーバーの管理方法をここに示す。

## 事前準備

1. Discord サーバーを用意する（認証運用を行う場合）。
2. Discord Developer Portal でアプリを作成し、Bot token を発行する。
3. 作成したアプリを対象 Discord サーバーへインストールする
   （`applications.commands` と `bot` スコープを付与）。

## `mc-ctl`

ほとんどの管理作業を自動化するCLIとして `mc-ctl` というコマンドを用意している
実行はGo buildまたはGo runで `go run ./cmd/mc-ctl` のように使用する。

GitHub Releases から配布バイナリを取得して使うこともできる。

```console
curl -L -o mc-ctl.tar.gz https://github.com/kyoh86/minecraft-server/releases/download/vX.Y.Z/mc-ctl_vX.Y.Z_linux_amd64.tar.gz
tar -xzf mc-ctl.tar.gz mc-ctl
./mc-ctl version
```

## 初回セットアップ

最初のサーバー開設時は以下の手順を実行すれば良い。
すべての設定がデフォルトなので、そのまますべての機能を使用することはできない点に注意。

1. 以下コマンドを実行する

```console
mc-ctl init
mc-ctl server up
mc-ctl world ensure
mc-ctl world setup
mc-ctl world spawn profile
mc-ctl world spawn stage
mc-ctl world spawn apply
```

`mc-ctl init` は対話入力で secret 設定を促す。
未入力のまま進めた場合は `secrets/mc_link_discord.toml`（`bot_token` / `guild_id` / `allowed_role_ids`）に
プレースホルダを設定し、`secrets/mc_forwarding_secret.txt` に自動生成値を設定し、
`secrets/mc_link_discord_guild_name.txt` を含む設定ファイルを生成し、
`secrets/playit_secret_key.txt` を生成し、
`secrets/limbo/server.toml` と `secrets/world/paper-global.yml` を描画する。
あわせて `infra/.env` を補完し、`LOCAL_UID` / `LOCAL_GID` を保存する。
また、`runtime` 配下の所有者が実行ユーザーと一致しない場合はエラーで停止する。
Redis/allowlist の環境変数名は `MC_LINK_*` に統一されており、旧 `MCLINK_*` は使用しない。
`mc-link` の secret ファイルパスは `MC_LINK_DISCORD_SECRET_FILE` で上書きできる（既定: `/run/secrets/mc_link_discord`）。

設定反映は `server up`/`server restart` 時に実行される。
`docker compose` は標準の `.env` 自動読込により `infra/.env` の
`LOCAL_UID` / `LOCAL_GID` をコンテナ実行ユーザーへ反映する。
`mc-ctl server up` は `docker compose up` 後、全サービスが `running` かつ
（healthcheckがある場合は）`healthy` になるまで待機し、未到達時はエラーを返す。
`mc-ctl server restart <service>` は再起動後、対象サービスが `running` かつ
（healthcheckがある場合は）`healthy` になるまで待機し、未到達時はエラーを返す。
`mc-ctl server restart <service> --build` は image 再ビルド + コンテナ再作成を行ってから
同じ readiness 待機を行う。`infra/world/**` や `infra/velocity/**`（ローカルプラグイン含む）
を変更した場合はこちらを使う。
`world` は起動時に image 同梱プラグイン資産を `/data` 側へ反映し、
`/config/paper-global.yml`（`secrets/world/paper-global.yml` を bind）を `/data/config` へ同期する。

## 各種ワールド設定変更の反映

各種ワールドに対する設定の変更を反映する場合は次を実行する。

```console
mc-ctl world setup
mc-ctl world spawn stage
mc-ctl world spawn apply
```

## プラグイン更新の反映

`infra/docker-compose.yml` の `MODRINTH_PROJECTS` や `SPIGET_RESOURCES` を更新した場合は、
`world` コンテナの再作成が必要。

```console
mc-ctl server restart world --build
```

特定ワールドだけセットアップを適用したい場合:

```console
mc-ctl world setup --world mainhall
```

## ワールド再生成

`deletable: true` のワールドだけ再生成できるてんに注意

```console
mc-ctl world regenerate --world resource
mc-ctl world setup --world resource
mc-ctl world spawn profile --world resource
mc-ctl world spawn stage --world resource
mc-ctl world spawn apply --world resource
```

特定ワールドだけスポーン関連処理を実行したい場合は `--world <name>` を指定できる。
`world spawn stage` は対象ワールドの `WorldGuard` 設定だけを更新し、`portals.yml` は全ワールド定義で再描画する。

## ワールド drop / delete

```console
mc-ctl world drop resource
mc-ctl world delete --yes resource
```

- `drop` は unload + remove だけ実行し、ワールドディスクは残す。
- `delete` は `drop` に加えてワールドディスクを削除する。
- `mainhall` は保護され、`drop`/`delete` できない。
- `delete` は `world.env.yml` の `deletable: true` が必要。
- `world ensure` / `world setup --world mainhall` 実行時は、`mainhall_nether` と `mainhall_the_end` を自動で drop する。
- `world setup --world mainhall` は `world.policy.yml` の `mv_set` を適用し、Hub の運用ポリシーを固定する。

## 任意 function 実行

```console
mc-ctl world function run mcserver:hello
```

## プレイヤー権限管理の変更

```console
mc-ctl player op grant kyoh86
mc-ctl player op revoke kyoh86
mc-ctl player admin grant kyoh86
mc-ctl player admin revoke kyoh86
mc-ctl player delink
mc-ctl player delink <uuid>
```

`mc-ctl player delink` は `runtime/velocity/allowlist.yml` を読み込み、
対話選択した1件を削除する。
`mc-ctl player delink <uuid>` を指定した場合は、対話なしで UUID エントリを削除する。

## 停止

```console
mc-ctl server down
```

## playit.gg トンネルの初期設定

`mc-ctl server up` 後、`playit` コンテナのログに claim 用URLが表示される。

```console
mc-ctl server logs playit
```

claim 完了後、playit.gg 側で Minecraft(Java) 用トンネルを作成し、
ローカル宛先を `127.0.0.1:25565` に設定する。

`playit` の設定は `runtime/playit/playit.toml` に保存されるため、
再起動後も再claimは不要。
`secrets/playit_secret_key.txt` が未設定（プレースホルダ）の場合、`playit` は待機状態となり
ログに設定不足メッセージを出力する。
