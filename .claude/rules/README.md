# .claude/rules/

Claude Code（および類似 AI エージェント）向けの **領域別ルール** を置くディレクトリ。
全体方針は [`/CLAUDE.md`](../../CLAUDE.md) を参照。

## ファイル一覧

| ファイル | 内容 |
|---|---|
| [`commit-style.md`](./commit-style.md) | コミットメッセージ規約 |
| [`go-style.md`](./go-style.md) | Go コーディング規約（バックエンド） |
| [`frontend-style.md`](./frontend-style.md) | TypeScript / React 規約（フロントエンド） |

## 追加ルール

新しいルールを追加するときは:

1. このディレクトリに `*.md` を追加する
2. 上記の表に 1 行追加する
3. 必要なら `/CLAUDE.md` から該当領域でリンクする
