# ワールド構成と再現方針

## 目的

以下の4ワールド構成を、手順ではなく定義ファイルで再現できるようにする。

- `mainhall` : 導線・集約用（`LEVEL`）
- `residence` : 通常プレイ・居住用
- `resource` : 定期再生成前提の資源用
- `factory` : 自動化・装置用

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

- `setup/wsl/worlds/schema.json`
  - `world.env.yml` 用 JSON Schema
- `setup/wsl/worlds/policy.schema.json`
  - `world.policy.yml` 用 JSON Schema
- `setup/wsl/worlds/<name>/world.env.yml`
  - 対象は Multiverse 管理ワールド（`residence` / `resource` / `factory`）
  - 先頭に `# yaml-language-server: $schema=../schema.json` を記述
  - `name` / `environment` / `world_type` / `seed` / `deletable`
- `setup/wsl/worlds/<name>/world.policy.yml`
  - 対象は全ワールド（`mainhall` を含む）
  - 先頭に `# yaml-language-server: $schema=../policy.schema.json` を記述
  - `mv_set` に `mv modify <world> set ...` の項目を記述する
  - `mainhall` は `gamemode=adventure` などの Hub 制約を管理し、
    `residence/resource/factory` は `difficulty/pvp/gamemode` を管理する
- `setup/wsl/worlds/<name>/setup.commands`
  - 対象は全ワールド（`mainhall` を含む）
  - 1行1コマンドで記述する
  - `wslctl world setup` 実行時に外側で `execute in <dimension> run <command>` を付与して実行する
  - `mv` 系コマンドはここに書かず、`world.policy.yml` に記述する
  - 座標依存の function 実行はここに書かない
- `setup/wsl/worlds/<name>/worldguard.regions.yml.tmpl`
  - `WorldGuard` のリージョン定義テンプレート
  - `wslctl world spawn stage` が runtime の `plugins/WorldGuard/worlds/<name>/regions.yml` へ描画する
- `setup/wsl/worlds/mainhall/portals.yml.tmpl`
  - `Multiverse-Portals` 用のポータル定義テンプレート
  - `wslctl world spawn stage` が runtime の `plugins/Multiverse-Portals/portals.yml` へ描画する

## Datapack とセットアップ

- Datapack 配置元: `setup/wsl/datapacks/world-base`
- Datapack 出力先: `setup/wsl/runtime/world/mainhall/datapacks/world-base`
- `mainhall` の地形生成は `setup/wsl/docker-compose.yml` の `LEVEL_TYPE=FLAT` で制御する
- `mainhall` のセットアップは `minecraft:overworld` を対象に実行する
- それ以外のワールドは `minecraft:<world>` を対象に実行する

## プリミティブ操作

- `wslctl world ensure`
  - world 定義に従って create/import する
  - `mainhall_nether` / `mainhall_the_end` は Overworld-only 方針のため自動で drop する
- `wslctl world setup [--world <name>]`
  - `setup.commands` を対象次元で実行する
  - `world.policy.yml` に定義された MV 管理項目を適用する
  - `world-base` datapack を runtime へそのままコピーする
- `wslctl world spawn profile`
  - `residence/resource/factory` の地表Y（`motion_blocking_no_leaves`）を検出する
  - `surface_y` と `anchor_y=surface_y-32` を runtime profile に保存する
  - 各ワールドに `mcserver_spawn_anchor_<world>` marker を配置する
  - `setworldspawn` と `mvsetspawn` を同期する
- `wslctl world spawn stage`
  - profile を必須入力として、`.tmpl` を runtime に描画する
  - `worldguard.regions.yml.tmpl` / `portals.yml.tmpl` を反映する
  - `reload` / `wg reload` / `mvp reload` を実行する
- `wslctl world spawn apply`
  - `mainhall` では `mcserver:mainhall/hub_layout` を適用する
  - `residence/resource/factory` では profile の `surface_y` を使い、
    `execute in <dimension> run execute positioned ...` で
    `mcserver:world/hub_layout` を適用する
- `wslctl world regenerate <name>`
  - world を削除して再生成する（`deletable: true` のみ）

## 補足

`world setup` は固定値適用のみを担当し、地表Y判定やポータル座標補正は行わない。
座標依存の反映は `world spawn profile/stage/apply` のみで行う。

## hub_layout

`mainhall` の初期スポーン付近に、導線確認用のデモ建築を配置できる。

```console
wslctl world function run mcserver:mainhall/hub_layout
```

この function は、御殿風の簡易ハブと `residence` / `resource` / `factory` 行きの
案内看板を設置する。
また、床下にゲート演出用の反復コマンドブロックを配置し、
`end_rod` と `enchant` のパーティクルを各ゲート面へ常時投影する。
各ゲートは背面を塞ぎ、フレーム中央に銅電球とレッドストーン入力を配置する。
西向き（`factory` 側）ゲートのガラス表示は、判定面への進入を妨げないよう
`x=-9.4` に配置する。

`mainhall` のハブは `wslctl world spawn apply` が
`mcserver:mainhall/hub_layout` を実行して構築する。
`residence` / `resource` / `factory` の小ハブは
`wslctl world spawn apply` が profile 座標を基準に構築する。
小ハブの東西出入口には、Mob に開けられないよう圧力板入力の鉄ドア回路を配置する。
同時に `worldguard.regions.yml.tmpl` の描画結果により
スポーン周辺での建設・破壊・爆破を禁止する。

`Multiverse-Portals` と `WorldGuard` のテンプレート反映:

```console
wslctl world spawn profile
wslctl world spawn stage
wslctl world spawn apply
```

`mainhall` の入口ポータルはゲート面に合わせて定義する
（`residence` は `z=-9` 面、`factory` は `x=-9` 面）。
`factory` 入口のみ `check-destination-safety: false` とし、
着地点安全判定による遷移拒否を防ぐ。
