# ワールド構成と再現方針

## 目的

以下の4ワールド構成を、手順ではなく定義ファイルで再現できるようにする。

- `mainhall` : 導線・集約用
- `residence` : 通常プレイ・居住用
- `resource` : 定期リセット前提の資源用
- `factory` : 自動化・装置用

## 定義場所

- `setup/wsl/worlds/schema.json`
  - `world.env.yml` 用 JSON Schema
- `setup/wsl/worlds/<name>/world.env.yml`
  - 先頭に `# yaml-language-server: $schema=../schema.json` を記述
  - ワールド名
  - 環境（`normal` など）
  - ワールドタイプ（`normal`/`flat` など）
  - seed
  - 実行する初期化 function 名
  - resettable フラグ

## 初期化ロジック

- Datapack: `setup/wsl/datapacks/world-base`
- 各ワールドの function エントリ:
  - `mcserver:worlds/mainhall/init`
  - `mcserver:worlds/residence/init`
  - `mcserver:worlds/resource/init`
  - `mcserver:worlds/factory/init`

`init.mcfunction` は分割可能で、現在は以下を分離している。

- `init/gamerules.mcfunction`
- `init/environment.mcfunction`
- `init/spawn.mcfunction`

## 実行コマンド

- `make worlds-bootstrap`
  - Datapack同期
  - reload
  - `world.env.yml` 定義に従った作成/import
  - 各worldの init function 実行

- `make world-reset WORLD=<name>`
  - 指定ワールドを unload
  - ワールドディレクトリ削除
  - 再作成/import
  - init function 再実行

## 補足

`world_settings.mcfunction` は互換用エントリとして残し、
`mcserver:worlds/mainhall/init` を呼び出す。
