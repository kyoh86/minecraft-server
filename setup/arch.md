# Minecraft server / Arch Linux

## 概要

マインクラフトサーバーを冒険/資材用と建築用サーバーでマルチサーバー構成で作る

## 要件

- サーバー間を行き来するプロキシサーバーにはVelocityを使用する
- Server SoftwareにはPaper forkのFoliaを使用する

## 基盤系のインストール

```console
$ pacman -Syyu
$ pacman -S jre21-openjdk    # Java (JRE 21)
```

