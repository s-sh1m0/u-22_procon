# フロントエンド コーディング規約

React 19 + TypeScript + Vite。

## 基本

- バージョンは README の技術スタックを正とする
- フォーマッタ: Prettier（保存時自動）
- リンタ: ESLint（導入予定）
- 型チェック: `tsc --noEmit` を CI で必須化する想定

## 動作確認（重要）

- **UI / 画面に影響する変更を含む場合は実機ブラウザで動作確認する**
- 確認手順:
  1. `docker compose up -d` で dev server を起動
  2. 変更画面を開いて golden path（正常系）を一通り操作
  3. 関連画面 / 既存機能のリグレッションも触って確認
  4. 必要に応じて DevTools コンソールでエラー / 警告が出ていないかチェック
- 理由: `npm run lint`・`tsc -b`・Prettier はコード品質の保証であり、「機能として動く」ことは保証しない。インタラクション・レイアウト崩れ・状態遷移バグは実機でしか出ない
- **AI エージェントが実装した場合**: CLI 環境では実機確認できないため、PR の Test plan に明示し、ユーザー側で確認してもらう

## パッケージ管理（重要）

- **`npm install` / `npm add` / `npm uninstall` / `npx` は必ず Docker コンテナ内で実行する**
- 理由: 環境一貫性のため。ホストとコンテナで Node のバージョンや lock の解決結果が食い違うと再現性が崩れる。常にコンテナ側を「正」とする
- 前提: `frontend/Dockerfile.dev` は `node` ユーザ (UID 1000) で実行し、`compose.yml` で `./frontend:/app` を bind mount している。これにより install 結果（`package.json` / `package-lock.json` / `node_modules/`）はホストにも反映され、ホスト IDE の補完も効く
- 実行方法:
  ```bash
  # 単発（コンテナを使い捨て）
  docker compose run --rm frontend npm install <パッケージ>

  # frontend コンテナが起動済みなら exec でも可
  docker compose exec frontend npm install <パッケージ>
  ```
- shadcn CLI 等の npx 経由ツールも同様に `docker compose run --rm frontend npx shadcn@latest add button` でコンテナから実行する

## ファイル構成

- `frontend/src/` 配下にコード
- コンポーネント: 1 ファイル 1 コンポーネント
- 関連する hooks / 型は同ディレクトリに置く

## TypeScript

- `any` 禁止。やむを得ない場合は `unknown` か `// @ts-expect-error` でコメント付き
- 型は `type` を基本に、宣言マージが必要なときだけ `interface`
- API レスポンス型は backend 側のスキーマと整合させる（生成ツール導入予定）

## React

- 関数コンポーネントのみ（class コンポーネント禁止）
- props 型は **必ず明示**
- `useEffect` は依存配列を正確に。lint 警告を握りつぶさない
- 状態管理: ローカルは `useState` / `useReducer`、サーバ状態は `@tanstack/react-query` を使う
- グローバル状態が必要になったら都度議論（Context / Zustand 等）

## スタイリング

- `tailwindcss` + `shadcn/ui` を基本とする
- カスタム CSS は最終手段
- shadcn コンポーネントは生成して `src/components/ui/` に置く

## グラフ可視化

- `@xyflow/react`（v12 系）を使う。**`reactflow`（v11）は使わない**
- ノード / エッジの型はプロジェクト共通の型として定義する

## エディタ

- diff 表示は `@monaco-editor/react`
