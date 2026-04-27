# フロントエンド コーディング規約

React 19 + TypeScript + Vite。

## 基本

- バージョンは README の技術スタックを正とする
- フォーマッタ: Prettier（保存時自動）
- リンタ: ESLint（導入予定）
- 型チェック: `tsc --noEmit` を CI で必須化する想定

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
