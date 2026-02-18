# Bot 制御インターフェース（Codex向け）

## 目的

`make` を介さず、Codex から機械可読なプロトコルで Bot を制御する。

## 形式

- 標準入力: JSON Lines（1行1JSON）
- 標準出力: JSON Lines（1行1JSON）

## 起動例

```console
docker run --rm -i --network infra_mcnet --user $(id -u):$(id -g) \
  -v $(pwd)/tools/bot:/work \
  -v $(pwd)/runtime:/runtime \
  -w /work \
  node:22-alpine sh -lc 'npm install --no-audit --no-fund && BOT_HOST=mc-world BOT_PORT=25565 BOT_AUTH=offline BOT_USERNAME=codexbot BOT_REPORT_DIR=/runtime/bot-reports npm run control'
```

`BOT_AUTH=offline` を使う場合は、サーバーを `ONLINE_MODE=FALSE` で起動しておく。
`ONLINE_MODE=TRUE` のまま使う場合は Bot 側も Microsoft 認証（`auth=microsoft`）が必要。

## リクエスト

- `connect`
- `disconnect`
- `status`
- `snapshot`
- `runScenario`（`name` 必須）
- `quit`

例:

```json
{"id":"1","action":"connect"}
{"id":"2","action":"runScenario","name":"smoke"}
{"id":"3","action":"snapshot"}
{"id":"4","action":"quit"}
```

## レスポンス

成功:

```json
{"id":"2","ok":true,"scenario":"smoke","result":"pass","report":"/runtime/bot-reports/....json"}
```

失敗:

```json
{"id":"2","ok":false,"error":{"code":"scenario_failed","message":"..."},"report":"/runtime/bot-reports/....json"}
```

## 常時監視（ローカル）

`smoke` を一定間隔で繰り返す:

```console
docker run --rm --network infra_mcnet --user $(id -u):$(id -g) \
  -v $(pwd)/tools/bot:/work \
  -v $(pwd)/runtime:/runtime \
  -w /work \
  node:22-alpine sh -lc 'npm install --no-audit --no-fund && BOT_HOST=mc-world BOT_PORT=25565 BOT_AUTH=offline BOT_USERNAME=codexbot BOT_REPORT_DIR=/runtime/bot-reports BOT_MONITOR_SCENARIO=smoke BOT_MONITOR_INTERVAL_SEC=300 npm run monitor'
```

## CI運用方針

- まずは手動トリガーまたは夜間で `smoke` を実行する
- 安定後に PR 必須チェックへ昇格する
