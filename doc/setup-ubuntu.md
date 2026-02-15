# Minecraft server / Ubuntu 24

## 概要

マインクラフトサーバーを冒険/資材用と建築用サーバーでマルチサーバー構成で作る

## 要件

- サーバー間を行き来するプロキシサーバーにはVelocityを使用する
- Server SoftwareにはPaper forkのFoliaを使用する

## 基盤系のインストール

```console
$ sudo apt update --yes
$ sudo apt upgrade --yes
$ sudo apt install --yes \
    make
```

