# Minecraft server / WSL 検証構成（Velocity + Paper）

## 概要

このドキュメントは、`WSL2 + Ubuntu` 上で `Velocity + Paper(lobby/survival)` を検証するための手順をまとめたもの。  
ここでの構成は検証用であり、常時運用の本番環境は Linux 実機に移行する前提。

## 構成

- `velocity` : エントリポイント（外部公開ポート `25565`）
- `lobby` : ロビー用 Paper サーバー
- `survival` : サバイバル用 Paper サーバー
- バックエンドサーバーは外部公開しない

## 前提

- WSL2 で Ubuntu が利用可能
- Docker Desktop + WSL integration 有効、または WSL 側 Docker Engine が利用可能

## 初期化

リポジトリルートで実行する。

```console
./setup/wsl/init.sh
# または
make init
```

これで以下が生成される。

- `setup/wsl/runtime/velocity/velocity.toml`
- `setup/wsl/runtime/velocity/forwarding.secret`
- `setup/wsl/runtime/lobby/`
- `setup/wsl/runtime/survival/`

## 起動

```console
docker compose -f setup/wsl/docker-compose.yml up -d
# または
make up
```

状態確認:

```console
docker compose -f setup/wsl/docker-compose.yml ps
docker compose -f setup/wsl/docker-compose.yml logs -f velocity
# または
make ps
make logs-velocity
```

## Velocity と Paper の連携設定

初回起動後、Paper 側に `paper-global.yml` が生成される。  
以下を `lobby` と `survival` の両方で設定する。

対象ファイル:

- `setup/wsl/runtime/lobby/config/paper-global.yml`
- `setup/wsl/runtime/survival/config/paper-global.yml`

設定値:

- `proxies.velocity.enabled: true`
- `proxies.velocity.online-mode: true`
- `proxies.velocity.secret: "<forwarding.secret と同じ値>"`

`<forwarding.secret と同じ値>` は次のファイルの値を使う。

- `setup/wsl/runtime/velocity/forwarding.secret`

設定変更後、再起動:

```console
docker compose -f setup/wsl/docker-compose.yml restart lobby survival velocity
# または
make restart
```

## 接続確認

- Minecraft クライアントから `localhost:25565` へ接続
- Velocity 経由で lobby へ入れることを確認
- サーバー移動コマンド（例: `/server survival`）で移動確認

## 検証終了

```console
docker compose -f setup/wsl/docker-compose.yml down
# または
make down
```

データを消して作り直す場合のみ、`setup/wsl/runtime/` を削除して再初期化する。

## Make ターゲット一覧

- `make init` : 検証用ディレクトリとテンプレートを初期化
- `make up` : 検証構成をバックグラウンド起動
- `make down` : 構成を停止
- `make restart` : `velocity/lobby/survival` を再起動
- `make ps` : コンテナ状態の確認
- `make logs` : 全サービスのログ追跡
- `make logs-velocity` : Velocity ログ追跡
- `make logs-lobby` : lobby ログ追跡
- `make logs-survival` : survival ログ追跡
