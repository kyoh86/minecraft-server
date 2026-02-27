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
    - `name` / `world_type` / `seed` / `deletable`
    - `world_type` は任意。未指定時は `normal` として扱う
    - `display_name`（任意）で `mainhall` ゲートの表示名を指定できる（未指定時は `name`）
    - `display_color`（任意）で `mainhall` ゲート表示色を指定できる（未指定時は `gold`）
    - `display_color` は Minecraft の色名（`green` など）または `#RRGGBB` を使用する
    - 基底ワールドは常に `normal` 環境として作成する
    - `name` / `display_name` / `dimensions.<dim>.name` は指定時に空文字不可
    - `dimensions` を記述する場合は `nether` または `end` の少なくとも一方が必須
    - `dimensions.nether` / `dimensions.end` を記述した次元のみ managed world として生成・リンクする
    - `dimensions.<dim>.name` は任意。未指定時は `<name>_nether` / `<name>_the_end`
    - `seed` の優先順位は `dimensions.<dim>.seed` > ルート `seed` > random
    - `dimensions.<dim>.seed` は任意。未指定（キー省略）の場合は次順位へフォールバックする
    - `seed` / `dimensions.<dim>.seed` は空文字を許可しない（random を使う場合は未指定）
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
- `worlds/mainhall/worldguard.regions.yml.tmpl`
    - `mainhall` 用 `WorldGuard` リージョン定義テンプレート
- `worlds/_defaults/worldguard.regions.yml.tmpl`
    - `mainhall` 以外のワールド向け既定 `WorldGuard` リージョン定義テンプレート
    - `mc-ctl world spawn stage` が runtime の `plugins/WorldGuard/worlds/<name>/regions.yml` へ描画する
- `worlds/mainhall/portals.yml.tmpl`
    - `Multiverse-Portals` 用のポータル定義テンプレート
    - `gate_<world>` / `gate_<world>_to_mainhall` を `.WorldItems` ループで生成する
    - `mc-ctl world spawn stage` が runtime の `plugins/Multiverse-Portals/portals.yml` へ描画する
- `worlds/mainhall/hub_layout.mcfunction.tmpl`
    - `mainhall` ハブのレイアウトテンプレート
    - `WorldItems` をループしてゲートと看板を生成する
    - `mc-ctl world spawn stage` が runtime の datapack に描画する
- `infra/world/schematics/hub.schem`
    - `mainhall` 以外のワールドで使用する Hub 建築スキーマ
    - `mc-ctl world spawn stage` / `mc-ctl world spawn apply` が runtime の
      `plugins/FastAsyncWorldEdit/schematics/hub.schem` へ同期する
- `infra/world/plugins/hub-terraform/src/main/resources/config.yml`
    - `HubTerraform` の地表判定設定
    - 非地形ブロック除外は `terrain.non_terrain.exact`（個別）/ `suffix` / `contains` で管理する
    - `exact` には液体・作業台系・装飾/加工済み建材（例: 黒曜石系、石レンガ系、金ブロック、TNT）を含め、自然地形判定から除外する
    - `contains` で語を指定して系統除外できる（例: `SANDSTONE` で砂岩系全般）
    - `surface.probe.*` で profile 用サンプリングの半径・刻み・下限を管理する
    - プラグイン内部に同等の除外リストは持たず、挙動は config のみを正として決定する
- `worlds/env.schema.json`
    - `world.env.yml` 用 JSON Schema
    - `default` はスキーマ補完ではなく、`mc-ctl` 実装上の既定動作を明示するために記述する
- `worlds/policy.schema.json`
    - `world.policy.yml` 用 JSON Schema

## Datapack とセットアップ

- Datapack 配置元: `datapacks/world-base`
- Datapack 出力先: `runtime/world/mainhall/datapacks/world-base`
- `mainhall` の地形生成は `infra/docker-compose.yml` の `LEVEL_TYPE=FLAT` で制御する
- `secrets/world/paper-global.yml` では `unsupported-settings.allow-unsafe-end-portal-teleportation: true` を設定し、
  マルチワールド環境で End ポータル遷移が安全判定で阻害されるのを防ぐ
- `mainhall` のセットアップは `minecraft:overworld` を対象に実行する
- それ以外のワールドは `minecraft:<world>` を対象に実行する

## プリミティブ操作

- `mc-ctl world ensure`
    - world 定義に従って create/import する
    - `dimensions` に定義された次元のみ create/import する
    - `dimensions` に定義された次元に対して `mvnp link nether/end` を自動適用する
    - 上記 link 機能は `Multiverse-NetherPortals` 導入を前提とする
    - 適用後に `Multiverse` 登録と world ディレクトリ実体の両方を検証する
    - `mainhall_nether` / `mainhall_the_end` は Overworld-only 方針のため自動で drop する
- `mc-ctl world setup [--world <name>]`
    - `setup.commands` を対象次元で実行する
    - `world.policy.yml` に定義された MV 管理項目を適用する
    - `world-base` datapack を runtime へそのままコピーする
- `mc-ctl world spawn profile [--world <name>]`
    - `HubTerraform` の `hubterraform probe <world>` を使って地表Y（`motion_blocking_no_leaves`）を中心周辺の複数点で検出する
    - サンプル点は `x,z=-24..24` を `12` 刻みで走査する（25点）
    - 各サンプルの地表Yは下限 `y=63` を適用する（`63` 未満は `63` に丸める）
    - 最終 `surface_y` は `中央値` / `平均値(切り捨て)` / `40パーセンタイル` の最小値を採用する
    - `y=64` 以上では `ice` / `packed_ice` / `blue_ice` / `snow` / `snow_block` を地表候補から除外する
    - `surface_y` と `anchor_y=surface_y-32` を runtime profile に保存する
    - 各ワールドに `mcserver_spawn_anchor_<world>` marker を配置する
    - `setworldspawn` と `mvsetspawn` を同期する
- `mc-ctl world spawn stage [--world <name>]`
    - `worldguard.regions.yml.tmpl` を対象ワールド分だけ runtime に描画する
    - `--world` なし実行時は、profile を入力に `portals.yml.tmpl` を runtime に描画する
    - `--world` なし実行時は、`mainhall/hub_layout.mcfunction.tmpl` も runtime に描画する
    - `--world` 指定時は、既存 `portals.yml` の対象ワールド定義だけを更新する
    - `reload` / `wg reload` / `mv reload` を実行する
    - `mvp config enforce-portal-access false` を実行する
    - `spawn_protected` / `clickmobs_allowed` の Y 範囲は `surface_y-8 .. surface_y+12` とする（`mainhall` は `-64 .. -35`）
    - `*_to_mainhall` ポータルの Y 範囲は `surface_y .. surface_y+3` とする
    - `*_to_mainhall` ポータル面は `x=-2..0, z=2..3`（中心 `x=-1, z=2.5`）とする
- `mc-ctl world spawn apply [--world <name>]`
    - `--world` なし実行時のみ、`mainhall` で `mcserver:mainhall/hub_layout` を適用する
    - `--world` 指定時は対象ワールドのみ適用する
    - `residence/resource/factory` では profile の `surface_y` を使い、
      `hubterraform apply <world> <surface_y>` でHub周辺整地を先に実行し、
      `hub.schem` を `0,surface_y,0` へ貼り付ける
- `mc-ctl world regenerate [--world <name>]`
    - world を削除して再生成する（`deletable: true` のみ）

## 補足

`world setup` は固定値適用のみを担当し、地表Y判定やポータル座標補正は行わない。
座標依存の反映は `world spawn profile/stage/apply` のみで行う。

## hub_layout

`mainhall` の初期スポーン付近に、導線確認用のデモ建築を配置できる。

```console
mc-ctl world function run mcserver:mainhall/hub_layout
```

この function は、御殿風の簡易ハブと管理対象ワールド行きの
案内表示を設置する。
各ゲートは北側の1辺に並べ、背面を塞ぎ、フレーム中央に銅電球とレッドストーン入力を配置する。
ゲート名は `text_display` で表示し、`world.env.yml` の `display_name` / `display_color` を使う。
表示は一律で `transformation.scale=[1.4,1.4,1.4]` を適用する。
配置座標は `x=<ゲート中央>, y=-56.5, z=-7.5`（ゲート面 `z=-9` から 1.5 ブロック手前）とする。
初回ログイン時の安全スポーン補正で屋根上に出ないよう、
中心座標（`0 -51 0`）の天井を開口している。

`mainhall` のハブは `mc-ctl world spawn stage` が world 定義から生成した function を
`mc-ctl world spawn apply` が `mcserver:mainhall/hub_layout` を実行して構築する。
`residence` / `resource` / `factory` の小ハブは
`mc-ctl world spawn apply` が profile 座標を基準に `hub.schem` を貼り付けて構築する。
小ハブは施工前に `HubTerraform` で次を実行する。

- `x,z=-32..32` を profile の `surfaceY` 高さへ平準化する
- 外周 `x,z=-64..64` を `smoothstep` 補間で元地形へ接続する
- 補間先の高さは元地形を近傍平均した値を使い、局所的な凹凸を抑える
- 表層1層とその下層は元地形のブロック種を可能な限り継承する
- 継承時は `HubTerraform` 設定（`terrain.non_terrain.exact` / `suffix` / `contains`）に一致する非地形ブロックを除外する
- `y=64` 以上では `ice` / `packed_ice` / `blue_ice` / `snow` / `snow_block` を地表判定に使わない
- 上空クリア高さは対象範囲の実地形最大Y + 96 を基準に決定し、浮島残りを防ぐ
- 基礎は `surfaceY-16` と `OCEAN_FLOOR` の低い方まで石で充填する
- 液体の再充填は `water` のみ行い、`lava` は再充填しない
- 水面の再充填は元の水面セルを起点に、整地後の地形高を見ながら隣接方向へ伝播して欠けを埋める
- 凍結面では `ice` の直下が `water` の列も再充填シードとして扱う（`packed_ice` / `blue_ice` はシード化しない）
- 伝播シードに使う水面は海面（`y=63`）基準で、水ブロック座標 `y=62` 以下に限定し、高高度の水源で低地が過充填されるのを防ぐ
- 伝播結果が未設定のセルには水再充填を行わない
- 再充填する水ブロックの上端は固定で `y=62`（海面 `y=63`）とし、列ごとの水位ぶれを抑制する

地表判定ロジックは `HubTerraform` 内の `probe`（profile用）と terraform本体（地形解析用）の2箇所に存在する。
除外対象ブロックや高度条件を変更する場合は、両ロジックを同時に更新して挙動を一致させる。
小ハブの東西出入口には、Mob に開けられないよう圧力板入力の鉄ドア回路を配置する。
同時に `worldguard.regions.yml.tmpl` の反映結果により
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

`mainhall` の入口ポータルは北側の1辺（`z=-9` 面）に並べて定義する。
`check-destination-safety` は全ワールドで `true` を使用する。

## ClickMobs のリージョン制御

`infra/world/plugins/clickmobs-region-guard` から build された
`ClickMobsRegionGuard.jar` を導入し、
`infra/world/plugins/clickmobs-region-guard/config/config.yml` の
`allowed_region_ids` に許可リージョンIDを列挙する。
運用時に反映される設定は上記 `config/config.yml` であり、
`src/main/resources/config.yml` は jar 同梱の初期値としてのみ扱う。
同設定の `login_safety` では、ログイン時の補正動作を有効化できる。
`spawn_safe.min/max.(x|y|z)` で安全範囲を明示し、範囲外に出現したプレイヤーを
`mainhall` spawn へ補正できる。

```yaml
allowed_region_ids:
  - clickmobs_allowed

login_safety:
  enabled: true
  mainhall_world: mainhall

spawn_safe:
  min:
    x: -7
    y: -58
    z: -7
  max:
    x: 7
    y: -55
    z: 7
```

列挙したリージョン内でのみ `ClickMobs` の捕獲・設置操作を許可する。
リージョン外では `ClickMobs` 操作イベントをキャンセルする。
`infra/world/plugins/clickmobs/config/config.yml` では `whitelisted_mobs: [?all]` を固定する。
標準では `residence/resource/factory` の `clickmobs_allowed`
（Hub 周辺 `x,z=-64..64`）を許可リージョンにする。

## 保護/許可エリアの可視化

プレイ中に次の範囲を把握しやすくするため、

- 破壊不可能エリア（`spawn_protected`）
- Mob 捕獲/解放許可エリア（`clickmobs_allowed`）

WorldGuard の `greeting` / `farewell` を使って入退域時にメッセージ表示する。
加えて `ClickMobsRegionGuard` がプレイヤー移動イベントで領域判定し、
`bossbar` で現在位置の状態を表示する。

- `保護エリア（建築・破壊不可）`
- `ClickMobs許可エリア`
