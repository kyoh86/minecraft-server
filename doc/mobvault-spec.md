# MobVault 仕様（現行実装）

## 目的

`MobVault` は、モブをサーバー内で一時保管し、別サーバーで再生成できるようにするためのプラグイン。  
対象構成は `Velocity + Paper(lobby/survival)`。

このドキュメントは、現在コードに実装されている仕様のみを記載する。

## 対象範囲

- 対象プラグイン: `plugins/mobvault/src/dev/kyoh86/minecraft/mobvault/MobVaultPlugin.java`
- 配備先: `lobby` / `survival` の両 Paper サーバー
- データ保存先: `postgres`（PostgreSQL）

## データモデル

保存テーブル: `mob_vault_entries`

- `id UUID PRIMARY KEY`
- `owner_uuid UUID NOT NULL`
- `source_server TEXT NOT NULL`
- `entity_type TEXT NOT NULL`
- `payload TEXT NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

`owner_uuid` で所有者を識別し、一覧表示と引き出しは所有者本人に限定する。

## 権限

- 権限ノード: `mobvault.use`
- 既定値: `op`（`plugin.yml` で `default: op`）

## コマンド

- `/mobvault tool`
  - MobVault専用ツールを配布
- `/mobvault deposit`
  - 近傍モブを預け入れ（後述）
- `/mobvault list`
  - 預かり一覧 GUI を開く
- `/mobvault withdraw <uuid>`
  - UUID指定で引き出し（保守・直接操作用）

## ツール方式の預け入れ

専用ツール（PDCキー: `mobvault_tool`）をメインハンドに持っている場合、以下を有効化する。

- モブ右クリック: 預け入れ実行
- モブ左クリック: ダメージをキャンセルして預け入れ実行

左クリックでモブが逃げる問題を回避するため、`EntityDamageByEntityEvent` をキャンセルする。

## 預け入れルール

- 預けられる対象:
  - `LivingEntity`
  - ただし `Player` と `ARMOR_STAND` は対象外
- `Tameable` かつ `tamed=true` の場合:
  - 飼い主UUIDが実行者と一致しないと預け入れ不可

## 引き出しルール

- GUIクリックまたは `/mobvault withdraw <uuid>` で引き出し
- 引き出し対象は `id + owner_uuid` で検索
  - 所有者不一致のデータは引き出せない
- 再生成に成功したら該当行を削除

## GUI 仕様

- `/mobvault list` で開く
- 最大表示件数: `vault.list_limit`（実装上、GUIは最大45件）
- 各スロット項目:
  - 表示名: `entity_type`
  - Lore: `id`, `source_server`, `created_at`
  - クリック時に該当 `id` を引き出し

## 設定ファイル

- `setup/wsl/runtime/lobby/plugins/MobVault/config.yml`
- `setup/wsl/runtime/survival/plugins/MobVault/config.yml`

主な設定項目:

- `database.jdbc_url`
- `database.user`
- `database.password`
- `database.connect_timeout_seconds`
- `vault.source_server`
- `vault.search_radius`
- `vault.list_limit`
- `tool.material`
- `tool.display_name`

## 現時点の制約

- モブの「騎乗状態」は対象外
- 所有者オフライン時の `Tameable` 復元は限定的（オンラインプレイヤーのみ owner 再設定）
- `payload` は `Properties` テキスト形式で保持しているため、将来スキーマ管理は要検討
