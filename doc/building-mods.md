# 建築デザイン用 Mod 選定（Issue #7）

## 結論

採用 Mod は **Litematica（Fabric）** とする。

- 理由1: 回路/建築を「設計図（`.litematic`）」として再利用できる
- 理由2: このサーバー構成（Paper 1.21.11）に対してクライアント側のみで導入できる
- 理由3: ワールド横断で同じ設計を持ち運びしやすい

## 候補比較

| 候補 | 主用途 | 1.21.11対応 | 導入先 | 判定 |
|---|---|---|---|---|
| Litematica (+ MaLiLib) | 設計図の読込/投影/配置支援 | 対応あり | クライアント | 採用 |
| MiniHUD (+ MaLiLib) | 補助オーバーレイ（情報表示） | 対応あり | クライアント | 補助候補 |
| WorldEdit CUI (Fabric) | WorldEdit選択範囲の可視化 | 対応あり | クライアント | 補助候補 |

## 導入前提

- サーバー側の追加導入は不要（本件はクライアント Mod）
- クライアントは Fabric Loader を使用する
- Minecraft バージョンは `1.21.11` を前提にする

## 導入手順（クライアント）

1. Fabric Loader `1.21.11` プロファイルを作成する
2. 以下の jar を `mods/` に配置する
   - `litematica-fabric-1.21.11-0.25.4.jar`
   - `malilib-fabric-1.21.11-0.27.5.jar`
3. （任意）補助表示が必要な場合のみ追加
   - `minihud-fabric-1.21.11-0.38.3.jar`
   - `WorldEditCUI-1.21.11+01.jar`
4. クライアントを起動し、サーバーへ接続する

## サンプル操作（読み込み・配置・確認）

### 1. 読み込み

1. Litematica のメニューを開く（デフォルト `M`）
2. `Schematic Placements` から `Load Schematic` を選択
3. `.litematic` ファイルを選ぶ

### 2. 配置

1. `Create Placement` で配置を作成する
2. 配置の原点を建築したい場所へ移動する
3. 必要に応じて `Rotation` / `Mirror` を調整する

### 3. 確認

1. サーバー内の基準点（例: Hub 座標）と重なるかを確認する
2. 投影と実ブロックの差分を確認する（Litematica の差分表示）
3. 問題なければ建築を進める

## 補足

- Litematica の貼り付け系機能はワールド権限やゲームモードの影響を受ける
- この環境では Survival ワールドがあるため、実作業は「投影で合わせる」運用を基本とする

## 参照

- Litematica（Modrinth）: https://modrinth.com/mod/litematica
- MaLiLib（Modrinth）: https://modrinth.com/mod/malilib
- MiniHUD（Modrinth）: https://modrinth.com/mod/minihud
- WorldEdit CUI (Fabric)（CurseForge）: https://www.curseforge.com/minecraft/mc-mods/worldeditcui-fabric
