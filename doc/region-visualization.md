# 保護/許可エリアの可視化方針

## 目的

プレイ中に次の範囲を現地で把握しやすくする。

- 破壊不可能エリア（`spawn_protected`）
- Mob 捕獲/放逐許可エリア（`clickmobs_allowed`）

## 比較

### 案1: 物理ブロックで常設表示

- 内容: 境界にガラス等を常設する
- 長所: 常に見える
- 短所: 景観を壊しやすく、建築導線を邪魔しやすい

### 案2: WorldEdit など管理者ツールで都度表示

- 内容: 管理者が選択表示を使って確認する
- 長所: 導入コストが低い
- 短所: 一般プレイヤーが現地で把握しにくい

### 案3: Datapack 関数でパーティクル境界を一時表示（採用）

- 内容: 必要時のみパーティクルで境界線を表示する
- 長所: 景観を固定で汚さない、全員が同時に確認できる
- 短所: 描画密度を上げると負荷が上がる

## 採用方式

案3（Datapack 関数で一時表示）を常用方式とする。

- `spawn_protected`: `minecraft:flame` で境界表示
- `clickmobs_allowed`: `minecraft:happy_villager` で境界表示
- 表示はワールドごとの固定座標・固定Y基準で描画する

## 実装

以下の関数を追加済み。

- `mcserver:region/show_spawn_protected`
- `mcserver:region/show_clickmobs_allowed`
- `mcserver:region/show_all`
- `mcserver:region/show_all_init`
- `mcserver:region/show_all_loop`

`show_spawn_protected` / `show_clickmobs_allowed` は
基準Yから `-50 .. +50` を 10 刻み（計11層）で境界線を描画する。
`show_all_loop` は `schedule` で 1秒ごとに `show_all` を再実行する。
`show_all_loop` の基準Yはワールドごとに固定値を使う。

- `mainhall`: `-58`
- `residence`: `68`
- `resource`: `106`
- `factory`: `63`

## 実行手順

`world setup` 後に `spawn apply` まで実行済みであることを前提とする。

### mainhall

```mcfunction
/execute in minecraft:overworld run function mcserver:region/show_spawn_protected
```

### resource

```mcfunction
/execute in minecraft:resource run function mcserver:region/show_all
```

### residence

```mcfunction
/execute in minecraft:residence run function mcserver:region/show_all
```

### factory

```mcfunction
/execute in minecraft:factory run function mcserver:region/show_all
```

## 各ワールド確認手順

1. 対象ワールドで表示関数を実行する
2. `spawn_protected` 境界内外でブロック設置可否を確認する
3. `clickmobs_allowed` 境界内外で Mob 捕獲/放逐可否を確認する（resource/residence/factory）
4. 境界線と実際の可否が一致することを確認する

## hub 常設表示

`world/hub_layout.mcfunction` では、Hub 内のリピートコマンドブロックを
`function mcserver:region/show_all_init` 実行にしている。
`show_all_init` が `show_all_loop` を開始し、以降は 1秒ごとに表示する。
