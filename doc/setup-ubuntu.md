# Minecraft server / Ubuntu 24.04（単一 Paper）

## 概要

この手順は Ubuntu 上で Docker を使い、単一の Paper サーバー（world）を起動するためのもの。

## 前提

- Ubuntu 24.04
- Docker Engine / Docker Compose が利用可能
- `go` が利用可能

## セットアップ

```console
sudo apt update
sudo apt install -y make golang-go
```

リポジトリルートで実行:

```console
wslctl setup init
wslctl server up
```

## 確認

```console
wslctl server ps
wslctl server logs world
```

## world セットアップ

```console
wslctl world ensure
wslctl world setup
wslctl world spawn profile
wslctl world spawn stage
wslctl world spawn apply
```

資源ワールド再生成:

```console
wslctl world regenerate resource
wslctl world setup --world resource
wslctl world spawn profile
wslctl world spawn stage
wslctl world spawn apply
```

## 停止

```console
wslctl server down
```
