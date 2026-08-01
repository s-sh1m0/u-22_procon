# DiffGraph

Go プロジェクト向けの GitHub PR レビュー支援ツール。大規模 PR を自動的に論理的なクラスタへ分割し、レビュアーが構造を保ったまま読み進められるようにする。

> AI（Claude Code）向けの運用ルールは [`CLAUDE.md`](./CLAUDE.md) を参照してください。

---

## クイックスタート（5 分で起動）

### 前提条件

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) がインストール済みであること
  - メモリ割当を **6 GB 以上** に設定しておく（Settings → Resources → Memory）
- GitHub アカウントを持っていること

### 1. リポジトリをクローン

```bash
git clone https://github.com/s-sh1m0/u-22_procon.git
cd u-22_procon
```

### 2. GitHub OAuth App を作成

1. <https://github.com/settings/developers> を開く
2. **OAuth Apps** → **New OAuth App** をクリック
3. 以下を入力して **Register application** をクリック:

   | 項目 | 値 |
   |---|---|
   | Application name | `DiffGraph`（任意） |
   | Homepage URL | `http://localhost:20080` |
   | Authorization callback URL | `http://localhost:20080/auth/github/callback` |

4. 作成後の画面で **Client ID** をコピーする
5. **Generate a new client secret** をクリックし、表示された **Client Secret** をコピーする

### 3. セットアップスクリプトを実行

```bash
./setup.sh
```

対話形式で Client ID / Client Secret を入力すると `.env` が生成される。

### 4. 起動

```bash
docker compose -f compose.prod.yml up -d
```

初回はビルドに数分かかる。完了したら **<http://localhost:20080>** を開く。

### 5. 停止 / データ削除

```bash
# 停止
docker compose -f compose.prod.yml down

# データも含めて完全削除（DB・キャッシュすべて消える）
docker compose -f compose.prod.yml down -v
```

---

## トラブルシューティング

| 症状 | 対処 |
|---|---|
| ログインできない / OAuth エラーが出る | GitHub OAuth App の **Authorization callback URL** が `http://localhost:20080/auth/github/callback` になっているか確認する。`.env` の `GITHUB_CALLBACK_URL` も同じ値にする |
| 解析が終わらない / OOM で落ちる | Docker Desktop のメモリ割当を **6 GB 以上** に増やす（Settings → Resources → Memory） |
| ポート 20080 が競合する | `.env` に `PORT=<別のポート>` を追加し、OAuth App の Homepage URL / Callback URL のポートも合わせて変更する |
| `./setup.sh` が Permission denied | `chmod +x setup.sh` を実行してから再度試す |

---

<details>
<summary>プロジェクト仕様（開発者向け）</summary>

## プロジェクト概要

### 解決する問題

GitHub の PR レビュー UI は変更行ベースの 1 次元 diff であり、以下の問題がある:

- 100 ファイル超の大規模 PR では、レビュアーがコードの構造を頭の中で再構築しないといけない
- ファイル間の依存関係が見えず、「この変更がどこに影響するか」を手動で調べる必要がある
- どこから読み始めるべきかの指針がない
- 結果として大規模 PR は形式的にしかレビューされない、または分割が要求されて開発フローが滞る

### ターゲットユーザー

- Go プロジェクトを GitHub で開発しているチーム
- 特に「中〜大規模 PR をレビューする立場」のシニア開発者・テックリード
- OSS メンテナ

### コアバリュー

「**100 ファイルの PR を、4 つの意味のあるクラスタに自動分割する**」これに尽きる。他の機能はすべてこのコア体験の補強。

### 非ゴール（コンテスト期間中）

以下は明示的にやらない。実装途中で誘惑されても手を出さない。

- 多言語対応（Go 以外は対応しない）※コンテスト後の拡張候補
- リアルタイムコラボレーション（複数レビュアーの同時編集、CRDT 同期）※コンテスト後の拡張候補
- AST ノードベースのコメント永続化（rebase 耐性）
- セルフホスト用の本格的な認証・マルチテナント
- VSCode 拡張、ブラウザ拡張
- モバイル対応
- LLM による意味的解析（クラスタの命名のみ将来的にオプション）

---

## 技術スタック

### バックエンド

| 領域 | パッケージ | バージョン | 理由 |
|---|---|---|---|
| 言語 | Go | **1.25.3** | AST 解析が公式ライブラリで強力 |
| AST 解析・呼び出しグラフ | `golang.org/x/tools` | **v0.44.0** | `go/packages`, `go/ast`, `go/callgraph` を含む |
| グラフ処理・クラスタリング | `gonum.org/v1/gonum` | **v0.17.0** | `graph/community` の Louvain 法を使用 |
| HTTP サーバー | `github.com/labstack/echo/v4` | **v4.15.1** | ミドルウェアが充実、使いやすい |
| GitHub API | `github.com/google/go-github/v69` | **v69.2.0** | 公式維持 |
| DB / キャッシュ | `modernc.org/sqlite` | **v1.49.1** | CGO 不要、デプロイが楽 |

### フロントエンド

| 領域 | パッケージ | バージョン | 理由 |
|---|---|---|---|
| UI ライブラリ | `react` | **19.2.5** | - |
| ビルドツール | `vite` | **8.0.10** | 高速、SPA に最適 |
| 言語 | `typescript` | **6.0.3** | - |
| ルーティング | `react-router-dom` | **v7.14.2** | 標準的な SPA ルーティング |
| UI コンポーネント | `shadcn/ui` + `tailwindcss` | **tailwind v4.2.4** | コード生成型、独自デザイン不要 |
| グラフ可視化 | `@xyflow/react` | **v12.10.2** | React Flow v12、ノードグラフに最適 |
| エディタ表示 | `@monaco-editor/react` | **v4.7.0** | VSCode と同じ diff 表示 |
| API 通信 | `@tanstack/react-query` | **v5.100.1** | キャッシュとローディング状態管理 |

### 選定理由メモ

- **Next.js ではなく React + Vite**: このアプリは実質 SPA（ログイン → PR 入力 → グラフ表示）。SEO・SSR 不要のため Next.js の恩恵がない。OAuth は Go バックエンドで完結させる。
- **Go バックエンド**: `go/ast`, `go/callgraph` は Go ネイティブで他言語に代替手段がない。将来の多言語対応は LSP ベースに移行する設計にしておく。
- **@xyflow/react**: reactflow v12 から改名。v11（`reactflow` パッケージ）は使わない。

---

## ブランチ戦略

- `main` ブランチは本番環境に対応する
- `develop` ブランチが統合用のブランチ。機能ブランチはここにマージする
- 作業は各 Issue に対応したブランチを `develop` から作成し、そこで行う
- ブランチ命名規則: `{種類}_{issue番号}_{概要}` （例: `feature_1_project-structure`, `fix_12_analysis-timeout`）
  - 種類: `feature`（新機能）, `fix`（バグ修正）, `chore`（設定・環境整備）, `docs`（ドキュメント）
- `main`・`develop` への直接コミットは禁止。必ず Pull Request を出してマージする

---

## 機能（3 層アーキテクチャ）

### Layer 1: AST + 依存解析エンジン（バックエンド）

- PR 前後のコードを `go/packages` でパース
- 関数・型・変数の定義と参照を抽出
- 変更された関数から「呼び出しグラフ」を双方向に辿る（caller/callee）
- diff のヒットした行が、どの関数 / 型のどこに当たるかをマッピング
- 出力: 「変更ノード」「影響を受ける可能性があるノード」のグラフ

### Layer 2: コラボレーション層（リアルタイム）

- 複数レビュアーが同時に同じ PR を見ている状態の同期（カーソル位置、開いているノード、コメント）
- コメントは行ではなく AST ノードにぶら下がる（rebase してもコメントが迷子にならない）
- スレッド機能、解決済みフラグ、絵文字リアクション

### Layer 3: ビジュアライゼーション（フロントエンド）

- `@xyflow/react` でクラスタ単位の依存グラフを可視化
- ノードクリックで `@monaco-editor/react` による diff 表示

---

## デザイン

- [こちら](https://raw.githack.com/s-sh1m0/u-22_procon/develop/design/preview.html)から確認可能

</details>
