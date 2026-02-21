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

## 初回セットアップ

最初のサーバー開設時は以下の手順を実行すれば良い。
すべての設定がデフォルトなので、そのまますべての機能を使用することはできない点に注意。

1. Bot token を secret に保存する

```console
cp secrets/mc_link_discord_bot_token.txt.example secrets/mc_link_discord_bot_token.txt
chmod 600 secrets/mc_link_discord_bot_token.txt
```

2. `infra/docker-compose.yml` の `mc_link.environment.MCLINK_DISCORD_GUILD_ID` に対象 Guild ID を設定する

3. Velocity / Limbo の forwarding secret を同一値で設定する

```console
SECRET="$(openssl rand -hex 24)"
printf '%s\n' "$SECRET" > infra/velocity/config/forwarding.secret
sed -i "s/^secret = \".*\"$/secret = \"$SECRET\"/" infra/limbo/config/server.toml
```

4. 以下コマンドの実行

```console
mc-ctl asset init
mc-ctl asset stage
mc-ctl server up
mc-ctl world ensure
mc-ctl world setup
mc-ctl world spawn profile
mc-ctl world spawn stage
mc-ctl world spawn apply
```

`mc-ctl asset stage` は runtime ディレクトリの存在と書込可能状態を確認する。
設定反映自体は `server up`/`server restart` 時に実行される。
`world` は起動時に `infra/world/config/bootstrap.sh` を実行し、
`infra/world/plugins/dist/*` と `infra/velocity/config/forwarding.secret` を `/data` 側へ反映する。

## 各種ワールド設定変更の反映

各種ワールドに対する設定の変更を反映する場合は次を実行する。

```console
mc-ctl world setup
mc-ctl world spawn stage
mc-ctl world spawn apply
```

特定ワールドだけセットアップを適用したい場合:

```console
mc-ctl world setup --world mainhall
```

## ワールド再生成

`deletable: true` のワールドだけ再生成できるてんに注意

```console
mc-ctl world regenerate resource
mc-ctl world setup --world resource
mc-ctl world spawn profile
mc-ctl world spawn stage
mc-ctl world spawn apply
```

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
```

## 停止

```console
mc-ctl server down
```
