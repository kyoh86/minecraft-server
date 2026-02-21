# ワールド設計

## ワールド責務（運用合意）

- `mainhall`
    - 役割は導線専用（Hub）
    - すべてのワールド移動は一度 `mainhall` を経由する
    - `mainhall` へ来たプレイヤーは、ゲート群を置いた閉鎖空間の固定座標へ着地させる
    - 破壊不能前提。基本ゲームモードは Adventure（Multiverse 設定で固定）
    - 地形は superflat、モブ湧きなし、難易度 peaceful、時間経過なし、天候変化なし
    - Overworld のみ（Nether/End なし）
- `residence`
    - 生活・拠点建築の主ワールド
    - World Border は `-5000 .. 5000` 想定
    - Nether/End あり
- `resource`
    - 採掘・伐採など資源回収の主ワールド
    - 定期再生成前提
    - Nether/End あり
- `factory`
    - 自動化・高負荷装置の集約ワールド
    - 拠点建築は自由
    - World Border は `-5000 .. 5000` 想定
    - Nether/End あり

上記 Border は `setup.commands` で `worldborder set 10000`（中心 `0 0`）として適用する。

## 流通導線

- プレイヤー導線は `mainhall` を中心に集約する
- アイテム流通も `mainhall` 経由を基本とする
- 座標整合が必要な独自ゲート導線は別途設計して対応する

## 定義場所

- `worlds/<name>/world.env.yml`
    - ワールド生成に必要な情報群
    - `mainhall` は `LEVEL` 基底ワールドのため `world.env.yml` は持たない（全て固定値）
    - 対象は Multiverse 管理ワールド（`residence` / `resource` / `factory`）
    - 先頭に `# yaml-language-server: $schema=../env.schema.json` を記述
    - `name` / `environment` / `world_type` / `seed` / `deletable`
    - `seed` が空文字の場合は `mv create -s` を付与せず、ランダムシードで生成する
- `worlds/<name>/world.policy.yml`
    - MultiVerse で設定するワールドごとの特性情報群
    - 対象は`mainhall` を含む全ワールド
    - 先頭に `# yaml-language-server: $schema=../policy.schema.json` を記述
    - `mv_set` に `mv modify <world> set ...` の項目を記述する
    - `mainhall` は `gamemode=adventure` などの Hub 制約を管理し、
      `residence/resource/factory` は `difficulty/pvp/gamemode` を管理する
- `worlds/<name>/setup.commands`
    - 対象は全ワールド（`mainhall` を含む）
    - ワールド内で発行する固定値の初期設定コマンド群。1行1コマンドで記述する
    - `mc-ctl world setup` 実行時に外側で `execute in <dimension> run <command>` を付与して実行する
    - `mv` 系コマンドはここに書かず、`world.policy.yml` に記述する
    - 座標依存の function 実行はここに書かない
- `worlds/<name>/worldguard.regions.yml`
    - `WorldGuard` のリージョン定義
    - `mc-ctl world spawn stage` が runtime の `plugins/WorldGuard/worlds/<name>/regions.yml` へコピーする
- `worlds/mainhall/portals.yml.tmpl`
    - `Multiverse-Portals` 用のポータル定義テンプレート
    - `mc-ctl world spawn stage` が runtime の `plugins/Multiverse-Portals/portals.yml` へ描画する
- `worlds/env.schema.json`
    - `world.env.yml` 用 JSON Schema
- `worlds/policy.schema.json`
    - `world.policy.yml` 用 JSON Schema

## Datapack とセットアップ

- Datapack 配置元: `datapacks/world-base`
- Datapack 出力先: `runtime/world/mainhall/datapacks/world-base`
- `mainhall` の地形生成は `infra/docker-compose.yml` の `LEVEL_TYPE=FLAT` で制御する
- `mainhall` のセットアップは `minecraft:overworld` を対象に実行する
- それ以外のワールドは `minecraft:<world>` を対象に実行する

## プリミティブ操作

- `mc-ctl world ensure`
    - world 定義に従って create/import する
    - `mainhall_nether` / `mainhall_the_end` は Overworld-only 方針のため自動で drop する
- `mc-ctl world setup [--world <name>]`
    - `setup.commands` を対象次元で実行する
    - `world.policy.yml` に定義された MV 管理項目を適用する
    - `world-base` datapack を runtime へそのままコピーする
- `mc-ctl world spawn profile`
    - `residence/resource/factory` の地表Y（`motion_blocking_no_leaves`）を検出する
    - `surface_y` と `anchor_y=surface_y-32` を runtime profile に保存する
    - 各ワールドに `mcserver_spawn_anchor_<world>` marker を配置する
    - `setworldspawn` と `mvsetspawn` を同期する
- `mc-ctl world spawn stage`
    - profile を必須入力として、`portals.yml.tmpl` を runtime に描画する
    - `worldguard.regions.yml` を runtime にコピーし、`portals.yml.tmpl` を反映する
    - `reload` / `wg reload` / `mvp config enforce-portal-access false` / `mv reload` を実行する
    - `spawn_protected` / `clickmobs_allowed` の Y 範囲は `-64 .. 319` とする
    - `*_to_mainhall` ポータルの Y 範囲は `surface_y .. surface_y+3` とする
- `mc-ctl world spawn apply`
    - `mainhall` では `mcserver:mainhall/hub_layout` を適用する
    - `residence/resource/factory` では profile の `surface_y` を使い、
      `execute in <dimension> run execute positioned ...` で
      `mcserver:world/hub_layout` を適用する
- `mc-ctl world regenerate <name>`
    - world を削除して再生成する（`deletable: true` のみ）

## 補足

`world setup` は固定値適用のみを担当し、地表Y判定やポータル座標補正は行わない。
座標依存の反映は `world spawn profile/stage/apply` のみで行う。

## hub_layout

`mainhall` の初期スポーン付近に、導線確認用のデモ建築を配置できる。

```console
mc-ctl world function run mcserver:mainhall/hub_layout
```

この function は、御殿風の簡易ハブと `residence` / `resource` / `factory` 行きの
案内看板を設置する。
また、床下にゲート演出用の反復コマンドブロックを配置し、
`end_rod` と `enchant` のパーティクルを各ゲート面へ常時投影する。
各ゲートは背面を塞ぎ、フレーム中央に銅電球とレッドストーン入力を配置する。
西向き（`factory` 側）ゲートのガラス表示は、判定面への進入を妨げないよう
`x=-9.4` に配置する。
初回ログイン時の安全スポーン補正で屋根上に出ないよう、
中心座標（`0 -51 0`）の天井を開口している。

`mainhall` のハブは `mc-ctl world spawn apply` が
`mcserver:mainhall/hub_layout` を実行して構築する。
`residence` / `resource` / `factory` の小ハブは
`mc-ctl world spawn apply` が profile 座標を基準に構築する。
小ハブの東西出入口には、Mob に開けられないよう圧力板入力の鉄ドア回路を配置する。
同時に `worldguard.regions.yml` の反映結果により
スポーン周辺での建設・破壊・爆破を禁止する。
ただし回路操作のため、`spawn_protected` では `interact` / `use` を許可する。
`ClickMobs` の利用可否は `ClickMobsRegionGuard` の設定で制御する。
`ClickMobs` 本体は `whitelisted_mobs: [?all]` とし、全モブ捕獲を有効化する。

`Multiverse-Portals` のテンプレート反映と `WorldGuard` の設定反映:

```console
mc-ctl world spawn profile
mc-ctl world spawn stage
mc-ctl world spawn apply
```

`mainhall` の入口ポータルはゲート面に合わせて定義する
（`residence` は `z=-9` 面、`factory` は `x=-9` 面）。
`factory` 入口のみ `check-destination-safety: false` とし、
着地点安全判定による遷移拒否を防ぐ。

## ClickMobs のリージョン制御

`infra/world/plugins/clickmobs-region-guard` から build された
`ClickMobsRegionGuard.jar` を導入し、
`infra/world/plugins/clickmobs-region-guard/config/config.yml` の
`allowed_regions.<world>` に許可リージョンIDを列挙する。

```yaml
allowed_regions:
  mainhall: []
  residence:
    - clickmobs_allowed
  resource:
    - clickmobs_allowed
  factory:
    - clickmobs_allowed
```

列挙したリージョン内でのみ `ClickMobs` の捕獲・設置操作を許可する。
リージョン外では `ClickMobs` 操作イベントをキャンセルする。
`infra/world/plugins/clickmobs/config/config.yml` では `whitelisted_mobs: [?all]` を固定する。
標準では `residence/resource/factory` の `clickmobs_allowed`
（Hub 周辺 `x,z=-64..64`）を許可リージョンにする。

## 保護/許可エリアの可視化

プレイ中に次の範囲を現地で把握しやすくするため、

- 破壊不可能エリア（`spawn_protected`）
- Mob 捕獲/放逐許可エリア（`clickmobs_allowed`）

Datapack 関数でパーティクル境界を表示する。

- `spawn_protected`: `minecraft:cherry_leaves` で境界表示
- `clickmobs_allowed`: `minecraft:end_rod` で境界表示

表示はワールドごとの固定座標・固定Y基準で描画する

### 実装

以下の関数を利用する。

- `mcserver:region/show_spawn_protected`
- `mcserver:region/show_clickmobs_allowed`
- `mcserver:region/show_all`
- `mcserver:region/init`
- `mcserver:region/show_all_loop`

`show_spawn_protected` / `show_clickmobs_allowed` は
基準Yから `-50 .. +50` を 10 刻み（計11層）で境界線を描画する。
`show_all_loop` は `schedule` で 1秒ごとに `show_all` を再実行する。
`show_all_loop` の基準Yはワールドごとに固定値を使う。

- `mainhall`: `-58`
- `residence`: `68`
- `resource`: `106`
- `factory`: `63`

### 常設表示

`world/hub_layout.mcfunction` では、Hub 内のリピートコマンドブロックを
`function mcserver:region/init` 実行にしている。
`init` が `show_all_loop` を開始し、以降は 1秒ごとに表示する。
