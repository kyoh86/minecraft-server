# mc-ctl リリース配布設計

## 目的

`mc-ctl` をソースコードなしでも利用できるようにし、運用者が同一の手順で導入できる状態を維持する。

## 配布対象

- `mc-ctl` 本体バイナリ
- `checksums.txt`

対象OS/アーキテクチャ:

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`
- `windows/amd64`
- `windows/arm64`

## 実装

- `/.goreleaser.yml`
  - `./cmd/mc-ctl` をクロスビルド
  - `-X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}` を埋め込む
  - `checksums.txt` を生成する
- `/.github/workflows/release-on-tag.yml`
  - `v*` タグ push で実行する
  - `go test ./cmd/mc-ctl/...` 成功後に GoReleaser を実行する

## 運用手順

1. `main` にリリース対象の変更が入っていることを確認する
2. `vX.Y.Z` 形式のタグを作成して push する
3. GitHub Actions `Release on tag` の完了を確認する
4. Release ページに成果物と `checksums.txt` があることを確認する

## バージョン確認

配布バイナリは以下で埋め込み情報を確認できる。

```console
mc-ctl version
```

