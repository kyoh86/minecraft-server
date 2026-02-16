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

上記 Border は初期化 function で `worldborder set 10000`（中心 `0 0`）として適用する。

## 流通導線

- プレイヤー導線は `mainhall` を中心に集約する
- アイテム流通も現時点では `mainhall` 経由を基本とする
- 座標整合が必要な独自ゲート導線は将来対応（現時点では後回し）

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

## Datapack と初期化 function

- Datapack 配置元: `setup/wsl/datapacks/world-base`
- Datapack 配置先: `setup/wsl/runtime/world/mainhall/datapacks/world-base`
- `mainhall` の地形生成は `setup/wsl/docker-compose.yml` の `LEVEL_TYPE=FLAT` で制御する
- `mainhall` の初期化コマンドは `minecraft:overworld` を対象に実行する
- 初期化 function 例:
  - `mcserver:worlds/mainhall/init`
  - `mcserver:worlds/residence/init`
  - `mcserver:worlds/resource/init`
  - `mcserver:worlds/factory/init`

初期化 function ID は `mcserver:worlds/<name>/init` の規則で決まる。

`init.mcfunction` は分割可能。

- `init/gamerules.mcfunction`
- `init/environment.mcfunction`
- `init/spawn.mcfunction`

## プリミティブ操作

- `wslctl assets stage`
  - Datapack を runtime 側へ配置する
- `wslctl server reload`
  - Datapack/function を再読み込みする
- `wslctl world ensure`
  - world 定義に従って create/import する
  - `mainhall_nether` / `mainhall_the_end` は Overworld-only 方針のため自動で drop する
- `wslctl world setup [--world <name>]`
  - world 初期化 function を実行する
  - `world.policy.yml` に定義された MV 管理項目を適用する
- `wslctl world regenerate <name>`
  - world を削除して再生成する（`deletable: true` のみ）

## 補足

`world_settings.mcfunction` は互換用エントリとして残し、
`mcserver:worlds/mainhall/init` を呼び出す。
