# CLAUDE.md - AI 運用ルール

このファイルは Claude Code（および類似 AI エージェント）向けの作業ルールです。
**プロジェクト仕様（概要・技術スタック・機能）は [`README.md`](./README.md) を参照してください。**

---

## 0. 大原則

- **プロジェクト仕様は `README.md` が正**。仕様が衝突したら README を優先する
- 詳細ルールは [`.claude/rules/`](./.claude/rules/) に分割している。該当領域を触る前に読むこと
- スキル（再利用可能な手順書）は [`.claude/skills/`](./.claude/skills/) に置く

---

## 1. 回答時のルール
- 挨拶・前置き・絵文字禁止。結論をシンプルに伝える。
- 指摘すべきことは率直に指摘する。

## 2. ブランチ・コミット・PR

- `main` / `develop` への直接コミット禁止
- 作業ブランチは **`develop` から** 切る（`main` からではない）
- タスクを一つ終わらせるごとにコミットする。
- ブランチ名: `{種類}_{issue番号}_{概要}`
  - 種類: `feature` / `fix` / `chore` / `docs`
  - 例: `feature_1_project-structure`, `chore_22_claude-config`
- コミットメッセージ規約: [`.claude/rules/commit-style.md`](./.claude/rules/commit-style.md)
- PR は `develop` 宛に作成し、`main` へは `develop` からマージする

---

## 3. 作業フロー（標準）

1. ユーザーから依頼を受ける
2. 関連 issue がなければ起票（`gh issue create`）
3. `develop` から作業ブランチを切る
4. 実装 → ローカル動作確認
5. コミット（[`.claude/rules/commit-style.md`](./.claude/rules/commit-style.md) に従う）
6. push して PR を作成（`develop` 宛）
7. ユーザーにレビュー依頼

**ユーザーが明示的に「コミットして」「PR 作って」と言うまで、勝手に push / PR 作成しないこと。**

---

## 4. 言語・領域別ルール

- Go コード（コーディング規約）: [`.claude/rules/go-style.md`](./.claude/rules/go-style.md)
- バックエンド アーキテクチャ（オニオン + リポジトリ DIP）: [`.claude/rules/backend-architecture.md`](./.claude/rules/backend-architecture.md)
- TypeScript / React: [`.claude/rules/frontend-style.md`](./.claude/rules/frontend-style.md)

---

## 5. パッケージ管理

- **フロントエンドの `npm install` / `npx <CLI>` は必ず Docker コンテナ内で実行する**（ホスト直接実行禁止）
  ```bash
  docker compose run --rm frontend npm install <パッケージ>
  docker compose run --rm frontend npx shadcn@latest add button
  ```
- 理由: 環境一貫性のため（コンテナ側を「正」とする）。`Dockerfile.dev` は `node` ユーザ (UID 1000) で動作し、`compose.yml` の bind mount で結果がホストにも反映される
- 詳細は [`.claude/rules/frontend-style.md`](./.claude/rules/frontend-style.md) の「パッケージ管理」節を参照

---

## 6. 困ったとき

- 仕様が曖昧 → ユーザーに質問する（推測で実装しない）
- ルールが矛盾 → README > CLAUDE.md > `.claude/rules/` の優先順
- ルールに無いケース → ユーザーに確認 → 合意後に該当ルールへ追記
