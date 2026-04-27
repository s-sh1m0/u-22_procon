# .claude/skills/

Claude Code 向けの再利用可能な手順書（skill）を置くディレクトリ。
各スキルは独立したサブディレクトリに `SKILL.md` を持つ。

## スキル一覧

| スキル | 概要 |
|---|---|
| [`new-issue-branch`](./new-issue-branch/SKILL.md) | 新規作業開始時の issue 起票 → develop ブランチ作成まで |

## SKILL.md の書式

```markdown
---
name: <skill-name>
description: <一行で何をするスキルか>
---

# <skill-name>

<目的・前提>

## 手順

1. ...
2. ...

## 注意

- ...
```

## 追加方法

1. このディレクトリ配下に `<skill-name>/SKILL.md` を作成
2. 上記のスキル一覧に 1 行追加
3. ユーザーがスラッシュコマンドや自然言語で呼び出して使う
