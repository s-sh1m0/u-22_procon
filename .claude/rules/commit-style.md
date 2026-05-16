# コミットメッセージ規約

[Conventional Commits](https://www.conventionalcommits.org/) ライクの簡易版を採用する。

## フォーマット

```
<type>: <subject>

[optional body]
```

- 1 行目（subject）は **日本語可・72 文字以内**
- 本文（body）は必要なら入れる（「なぜ」を書く。「なに」は diff を見れば分かる）

## type 一覧

| type | 用途 | 例 |
|---|---|---|
| `feat` | 新機能追加 | `feat: GitHub OAuth 認証実装` |
| `fix` | バグ修正 | `fix: トークンリフレッシュ時の競合状態を解消` |
| `chore` | 設定・環境整備、依存更新など | `chore: settings.json の編集` |
| `docs` | ドキュメントのみ | `docs: README に技術スタックを追記` |
| `refactor` | 機能変更を伴わない内部改善 | `refactor: 認証ハンドラを util に切り出し` |
| `test` | テスト追加・修正 | `test: callgraph パーサのユニットテスト追加` |
| `style` | フォーマット・空白等 | `style: gofmt 適用` |
| `perf` | パフォーマンス改善 | `perf: クラスタリング処理を並列化` |

## ルール

- 1 コミット = 1 論理変更。1つのコミットについて、小さくできるだけ小さくする。
- WIP コミットを `develop` / `main` にマージしない（squash か rebase で潰す）
- AI が補助した場合のフッターは任意。プロジェクトでは現状 **付けない方針**
- `--no-verify` は禁止。pre-commit hook が落ちたら原因を直す
