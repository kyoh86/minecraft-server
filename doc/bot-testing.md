# Bot 検証基盤仕様（Issue #6）

この文書は、Minecraft 内 Bot を Codex から操作して検証を自動化するための仕様を定義する。

## 目的

- ポータル・保護範囲・スポーン周辺の不具合を、再現可能な手順で確認できるようにする
- 検証結果を機械可読なレポートとして保存し、失敗時の切り分けを可能にする

## 初回スコープ

- Bot 実装は Mineflayer を採用する
- 検証用 Bot は 1 体
- 自動シナリオは 3 本
  - `world_transfer_command_roundtrip`（基盤確認用）
  - `portal_resource_to_mainhall`（ポータル不具合再現用）
  - `smoke`（主要2シナリオ連続実行）
- 結果レポートは `runtime/bot-reports/` に JSON で保存する

## 運用コマンド

- `make bot-up`
  - Bot を接続して待機状態にする
  - 接続先ワールドは固定前提を置かない
- `make bot-test SCENARIO=world_transfer_command_roundtrip`
  - シナリオを実行し、`pass/fail` を判定する
  - 実行時のみ `ONLINE_MODE=FALSE` でサーバーを再起動し、終了後は `ONLINE_MODE=TRUE` に戻す
- `make bot-test SCENARIO=smoke`
  - ポータルと転送の主要確認を連続実行する
- `make bot-down`
  - Bot を停止する
- `make bot-report-latest`
  - 最新レポートを表示する

Codex からの操作は `tools/bot/src/control.js` を使い、標準入出力で JSON Lines を扱う。
`make` は補助手段とし、主要インターフェースは機械可読プロトコルとする。

前提:

- Docker ネットワーク `infra_mcnet` が利用可能であること
- Bot 接続認証は `BOT_AUTH=offline` を前提とする
- `BOT_AUTH=offline` の実行時はサーバーを `ONLINE_MODE=FALSE` で起動すること
- Bot 名は Minecraft 制限（16文字）以内に収まる形式で生成する

## シナリオ設計方針

- 各シナリオは前提状態を自己セットアップする
- `bot-up` の直後状態に依存しない
- シナリオ開始時に必要な次元・座標へ Bot を移動してから判定する

## 初回シナリオ仕様

シナリオ名: `world_transfer_command_roundtrip`

判定項目:

1. `/mvtp resource` 後に `resource` 側高さへ遷移する
2. `/mvtp mainhall` 後に `mainhall` 側高さへ遷移する

### 不具合再現シナリオ

シナリオ名: `portal_resource_to_mainhall`

判定項目:

1. 前提セットアップ後、Bot が `resource` にいる
2. ポータル侵入後、制限時間内に `mainhall` へ遷移する
3. 遷移後座標がハブ中心の許容範囲内にある

いずれかが不成立なら `fail` とする。

## レポート仕様

出力先: `runtime/bot-reports/<timestamp>-<scenario>.json`

必須項目:

- `scenario`
- `started_at`
- `ended_at`
- `duration_ms`
- `start.dimension`
- `start.pos`
- `end.dimension`
- `end.pos`
- `checks[]`（各判定の `pass/fail`）
- `logs[]`（主要イベント）
- `result`（`pass` / `fail`）
- `error`（例外時）

## コントロールプロトコル（JSON Lines）

入力は1行1 JSON:

```json
{"id":"1","action":"connect"}
{"id":"2","action":"runScenario","name":"smoke"}
{"id":"3","action":"snapshot"}
{"id":"4","action":"quit"}
```

応答も1行1 JSON:

```json
{"id":"1","ok":true,"connected":true}
{"id":"2","ok":true,"scenario":"smoke","result":"pass","report":"..."}
{"id":"3","ok":true,"pos":{"x":0.5,"y":81,"z":-8}}
{"id":"4","ok":true,"quitting":true}
```

## ディレクトリ構成

- `tools/bot/`
- `tools/bot/src/scenarios/`
- `runtime/bot-reports/`

## 受け入れ条件

1. Bot の導入・起動・停止手順が `doc/` に記載されている
2. Codex からシナリオ実行コマンドを起動できる
3. 基盤確認シナリオ 1 本が自動実行され、`pass/fail` を出力する
4. 失敗時にレポートから座標・次元・判定結果を追跡できる

## 現状の制約

- Paper `1.21.11` に対して、Mineflayer Bot がログイン初期化で切断される場合がある
- 切断時は `runtime/bot-reports/` の JSON で失敗内容を回収する
