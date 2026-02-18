# Codex 運用ガイド

この文書は、このリポジトリで Codex を使って作業するための最小運用手順を定義する。

## 1. Bot 読み込み手順

1. リポジトリ直下で Codex セッションを開始する。
2. 次を実行して、運用指示ファイルが存在することを確認する。

```console
ls -l AGENTS.md
```

3. 次を実行して、主要ディレクトリが想定どおり配置されていることを確認する。

```console
ls -d cmd infra worlds datapacks doc runtime
```

## 2. Skills の運用

このリポジトリ作業で使う Skills は、セッションに提示された `AGENTS.md` の一覧を正とする。

- `denops-author`
- `gogh-cli`
- `skill-creator`
- `skill-installer`

### 導入方針

- 上記がセッションで利用可能なら、そのまま使う。
- 不足がある場合は `skill-installer` を使って補完する。
- スキル本文は必要箇所だけ参照し、無関係なファイルは読まない。

## 3. MCP 接続確認手順

このリポジトリでは、Codex セッション上の MCP リソース列挙を接続確認の基準とする。

確認項目:

1. `list_mcp_resources` が実行できること
2. `list_mcp_resource_templates` が実行できること
3. 返却結果が空でも、呼び出し自体が成功すること

## 4. 検証フロー（再現手順）

以下を順に実行し、Codex がこのリポジトリで操作可能であることを確認する。

```console
make setup-init
make server-up
make world-ensure
make world-setup
make world-spawn-profile
make world-spawn-stage
make world-spawn-apply
make server-ps
```

判定基準:

- `make` の各ステップが終了コード `0` で完了する
- `make server-ps` で `mc-world` が `healthy` と表示される

