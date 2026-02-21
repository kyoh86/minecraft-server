# Minecraft サーバー管理マニュアル

このプロジェクトで構築されるMinecraftサーバーの管理方法をここに示す。

## 事前準備

1. Discord サーバーを用意する（認証運用を行う場合）。
2. Discord Developer Portal でアプリを作成し、Bot token を発行する。
3. 作成したアプリを対象 Discord サーバーへインストールする
   （`applications.commands` と `bot` スコープを付与）。

## `mcctl`

ほとんどの管理作業を自動化するCLIとして `mcctl` というコマンドを用意している
実行はGo buildまたはGo runで `go run ./cmd/mcctl` のように使用する。

## 初回セットアップ

最初のサーバー開設時は以下の手順を実行すれば良い。
すべての設定がデフォルトなので、そのまますべての機能を使用することはできない点に注意。

1. Bot token を secret に保存する

```console
cp secrets/mclink_discord_bot_token.txt.example secrets/mclink_discord_bot_token.txt
chmod 600 secrets/mclink_discord_bot_token.txt
```

2. `infra/docker-compose.yml` の `mclink.environment.MCLINK_DISCORD_GUILD_ID` に対象 Guild ID を設定する

3. Velocity / Limbo の forwarding secret を同一値で設定する

```console
SECRET="$(openssl rand -hex 24)"
printf '%s\n' "$SECRET" > infra/velocity/config/forwarding.secret
sed -i "s/^secret = \".*\"$/secret = \"$SECRET\"/" infra/limbo/config/server.toml
```

4. 以下コマンドの実行

```console
mcctl asset init
mcctl asset stage
mcctl server up
mcctl world ensure
mcctl world setup
mcctl world spawn profile
mcctl world spawn stage
mcctl world spawn apply
```

`mcctl asset stage` は runtime ディレクトリの存在と書込可能状態を確認する。
設定反映自体は `server up`/`server restart` 時に実行される。
`world` は起動時に `infra/world/config/bootstrap.sh` を実行し、
`infra/world/plugins/dist/*` と `infra/velocity/config/forwarding.secret` を `/data` 側へ反映する。

## 各種ワールド設定変更の反映

各種ワールドに対する設定の変更を反映する場合は次を実行する。

```console
mcctl world setup
mcctl world spawn stage
mcctl world spawn apply
```

特定ワールドだけセットアップを適用したい場合:

```console
mcctl world setup --world mainhall
```

## ワールド再生成

`deletable: true` のワールドだけ再生成できるてんに注意

```console
mcctl world regenerate resource
mcctl world setup --world resource
mcctl world spawn profile
mcctl world spawn stage
mcctl world spawn apply
```

## ワールド drop / delete

```console
mcctl world drop resource
mcctl world delete --yes resource
```

- `drop` は unload + remove だけ実行し、ワールドディスクは残す。
- `delete` は `drop` に加えてワールドディスクを削除する。
- `mainhall` は保護され、`drop`/`delete` できない。
- `delete` は `world.env.yml` の `deletable: true` が必要。
- `world ensure` / `world setup --world mainhall` 実行時は、`mainhall_nether` と `mainhall_the_end` を自動で drop する。
- `world setup --world mainhall` は `world.policy.yml` の `mv_set` を適用し、Hub の運用ポリシーを固定する。

## 任意 function 実行

```console
mcctl world function run mcserver:hello
```

## プレイヤー権限管理の変更

```console
mcctl player op grant kyoh86
mcctl player op revoke kyoh86
mcctl player admin grant kyoh86
mcctl player admin revoke kyoh86
```

## 停止

```console
mcctl server down
```
