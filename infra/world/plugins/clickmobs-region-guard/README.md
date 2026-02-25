# ClickMobsRegionGuard

WorldGuard のリージョンIDに基づいて `clickmobs.pickup` / `clickmobs.place` を制御する
Paper プラグイン。

`status_bossbar` を有効にすると、プレイヤー移動イベントを契機に
現在のリージョン状態を `bossbar` で表示する。

## build

```console
docker run --rm \
  -v "$PWD/infra/world/plugins/clickmobs-region-guard:/workspace" \
  -w /workspace \
  maven:3.9.9-eclipse-temurin-21 \
  mvn -DskipTests package
```

生成物:

- `target/clickmobs-region-guard-0.1.0.jar`
