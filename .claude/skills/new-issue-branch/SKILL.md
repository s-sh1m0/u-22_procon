---
name: new-issue-branch
description: 新しい作業を始めるときに、issue 起票 → develop からブランチ作成までを一気通貫で行う。
---

# new-issue-branch

新しい作業を始めるときの定型手順。
ユーザーが「〇〇したい」と依頼してきたら、このスキルに従って準備する。

## 手順

1. **作業内容をユーザーと確認**
   - タイトル（簡潔に）
   - 種類: `feature` / `fix` / `chore` / `docs`
   - 概要（slug 用）: 英小文字 + ハイフン（例: `claude-config`）
   - 必要なら詳細（背景・受け入れ条件）も聞く

2. **既存 issue を確認**
   - `gh issue list --repo s-sh1m0/u-22_procon --search "<キーワード>"` で重複が無いか確認
   - 重複があったらそれを使う

3. **Issue 起票**
   ```bash
   gh issue create --repo s-sh1m0/u-22_procon \
     --title "<種類>: <タイトル>" \
     --body "..." \
     --label "<種類>"
   ```
   - ラベルが無ければ `gh label create` で作成
   - 既存ラベル: `bug`, `documentation`, `enhancement`, `chore`, `backend`, `frontend`, `infra`, `auth` 他

4. **ブランチ作成**
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b "<種類>_<issue番号>_<概要>"
   ```

5. **準備完了をユーザーに報告**
   - 起票した issue URL
   - 切ったブランチ名

## 注意

- `develop` から切る（`main` ではない）
- ユーザーの明示的な指示なしに自動で push しない
- ブランチ名規約は `/CLAUDE.md` および `/README.md` に従う
