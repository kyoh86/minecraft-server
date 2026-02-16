# Minecraft server / Ubuntu 24.04（単一 Paper）

## 概要

この手順は Ubuntu 上で Docker を使い、単一の Paper サーバー（lobby）を起動するためのもの。

## 前提

- Ubuntu 24.04
- Docker Engine / Docker Compose が利用可能
- `make` が利用可能

## セットアップ

```console
sudo apt update
sudo apt install -y make
```

リポジトリルートで実行:

```console
make init
make up
```

## 確認

```console
make ps
make logs-lobby
```

## 内部設定（gamerule など）再適用

```console
make lobby-apply
```

このコマンドは Datapack を同期し、`function mcserver:lobby_settings` を実行する。

## 停止

```console
make down
```
