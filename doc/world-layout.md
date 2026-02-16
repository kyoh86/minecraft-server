# ワールド構成と再現方針

## 目的

以下の4ワールド構成を、手順ではなく定義ファイルで再現できるようにする。

- `mainhall` : 導線・集約用（`LEVEL`）
- `residence` : 通常プレイ・居住用
- `resource` : 定期再生成前提の資源用
- `factory` : 自動化・装置用

## 定義場所

- `setup/wsl/worlds/schema.json`
  - `world.env.yml` 用 JSON Schema
- `setup/wsl/worlds/<name>/world.env.yml`
  - 対象は Multiverse 管理ワールド（`residence` / `resource` / `factory`）
  - 先頭に `# yaml-language-server: $schema=../schema.json` を記述
  - `name` / `environment` / `world_type` / `seed` / `deletable`

## Datapack と初期化 function

- Datapack 配置元: `setup/wsl/datapacks/world-base`
- Datapack 配置先: `setup/wsl/runtime/world/mainhall/datapacks/world-base`
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
- `wslctl world setup [--world <name>]`
  - world 初期化 function を実行する
- `wslctl world regenerate <name>`
  - world を削除して再生成する（`deletable: true` のみ）

## 補足

`world_settings.mcfunction` は互換用エントリとして残し、
`mcserver:worlds/mainhall/init` を呼び出す。
