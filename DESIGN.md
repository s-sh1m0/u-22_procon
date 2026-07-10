# DESIGN.md - UI / UX デザイン指針

このドキュメントは ReviewArena のフロントエンド（`frontend/`）で UI に関わる作業を行うときの指針をまとめる。

**コンポーネント追加・スタイル変更・新画面作成のときは作業前に本ドキュメントを参照すること。**

コーディング規約（命名・型・パッケージ管理など）は [`.claude/rules/frontend-style.md`](./.claude/rules/frontend-style.md) を見る。本ドキュメントは「見た目とユーザー体験の決まり」のみ扱う。

---

## 0. このドキュメントの読み方

3 層構造になっており、上ほど安定で下ほど変わりやすい。

| セクション | 安定度 | 変更時の作法 |
|---|---|---|
| 1. 原則 | 高 | プロジェクト方向性が変わるとき以外いじらない |
| 2. 規約 | 中 | 作業を進める中で見直しが必要なら更新 |
| 3. 現在の実装 | 低 | UI 変更に追従して都度更新する |

実装と本ドキュメントが食い違ったら、まず実装が正しいか確認し、正しければ本ドキュメント側を直す。

---

## 1. デザイン原則

### 1.1 開発者ツールである
- 装飾より読みやすさを優先する。グラデーション・派手な影・装飾的アイコンは避ける
- 情報密度を優先する。余白を取りすぎない（ただし窮屈にもしない）
- 絵文字は使わない（プロジェクト全体ルール）

### 1.2 状態は明示する
すべての非同期データ表示で **loading / empty / error / success** の 4 状態を分岐する。「データなし」を「ローディング中」と区別がつかない見た目にしない。

### 1.3 中立トーン
- 中立色（warm neutral）を基調に、強調色はクラスタ・差分のような意味のある場所にだけ使う
- 「デフォルトは目立たない、意味があるところだけ目立つ」を徹底

### 1.4 アクセシビリティ最低限
- 操作要素の `button` には `aria-label`（テキスト無しの場合）
- 装飾的な SVG / アイコンには `aria-hidden`
- キーボード操作で詰まないこと（フォーカスリングは消さない）

---

## 2. 規約

### 2.1 レイヤー構造

| 層 | 場所 | 役割 |
|---|---|---|
| プリミティブ | `src/components/ui/` | shadcn/ui で生成。汎用的な部品 |
| レイアウト | `src/components/layout/` | 画面骨格（AppShell・ヘッダー・サイドバー） |
| 機能コンポーネント | `src/components/<domain>/` | ドメイン特化（graph・diff・auth など） |
| ページ | `src/pages/` | ルート単位の組み立て |

新規 UI を作るときは、まずプリミティブで足りないか確認する。足りないものを追加する場合は shadcn CLI（コンテナ内）で生成し、`ui/` に置く。

### 2.2 配色

- **中立色** には Tailwind の `stone` を使う（warm neutral 基調）。`gray` / `slate` / `zinc` は使わない
- **強調色** はトークン化されているものを優先（`--color-brand` 等）。Tailwind の任意色（`bg-[#abcdef]` のような hex 直書き）を新規追加しない
- **クラスタ色** は `src/lib/clusterColors.ts` のパレットを必ず経由する。直接色を書かない
- **意味付きの色**:
  - 変更マーカー: `amber`（`changed=true` の関数・ファイル）
  - エラー: `red`
  - 成功: `emerald` / `teal` 系（brand と整合）

### 2.3 タイポグラフィ

| 用途 | フォント |
|---|---|
| 本文・ラベル | sans（`--font-sans`） |
| 識別子・パス・コード・行番号 | mono（`--font-mono`） |
| 見出し | sans（`--font-heading`、現状 sans と同じ） |

ID やファイルパスは必ず mono にする。視認性と「これはコード由来の文字列」の合図を兼ねる。

### 2.4 サイズ・余白

- パネルや行のパディング: `px-4 py-3` を基本
- 角丸: 通常 `rounded`、強い区切り `rounded-md`、ピル状ラベル `rounded-full`
- 線: `border-stone-100` / `border-stone-200` の 2 段階
- 影: `shadow-sm` 程度。重い影は使わない

### 2.5 4 状態の表現パターン

| 状態 | パターン |
|---|---|
| loading | アニメーション付きスピナー + 短い説明テキスト |
| empty | `text-stone-400` で状況を 1 行説明（"このファイルに変更はありません" 等） |
| error | カード or バナー型で `red` 系背景、再試行導線（リンク or ボタン） |
| success | データを淡々と表示（成功通知の固有 UI は不要） |

### 2.6 命名・配置

- 機能領域ごとにディレクトリを切る（`graph/` `diff/` `auth/` `branding/` 等）
- 1 ファイル 1 コンポーネント
- 関連 hooks や型は同ディレクトリに置く

---

## 3. 現在の実装

> **このセクションは変わりやすい。** 実装変更の都度更新する。網羅性より「現状の最低限の手がかり」を狙う。

### 3.1 配色トークン

`src/index.css` の `@theme` で定義:

| トークン | 値 | 用途 |
|---|---|---|
| `--color-brand` | `#0f766e` (teal-700) | ロゴ・主要アクセント |
| `--color-brand-soft` | `#ccfbf1` (teal-100) | brand 背景 |
| `--color-bg` | `#fafaf9` (stone-50) | アプリ背景 |
| `--color-ink` | `#1c1917` (stone-900) | 主要テキスト |
| `--color-ink-2` | `#44403c` | 補助テキスト |
| `--color-ink-3` | `#78716c` | キャプション |

クラスタ用 8 色パレットは `src/lib/clusterColors.ts`。

### 3.2 フォント

- sans: Inter Tight / Geist Variable
- mono: JetBrains Mono / ui-monospace

### 3.3 レイアウト骨格

`AppShell.tsx`:
```
┌─────────────────────────┐
│ topBar                  │
├──────────────┬──────────┤
│ children     │ rightPanel
│ (graph 等)   │ (詳細)   │
└──────────────┴──────────┘
```

- 右パネルは詳細表示時のみマウント
- 右パネル幅は内容に応じて切替（通常 80、diff 表示時 700）

### 3.4 主要コンポーネント

| コンポーネント | 役割 |
|---|---|
| `PRMetaBar` | 上部固定バー。PR タイトル・ブランチ情報・ユーザー情報 |
| `DependencyGraph` | `@xyflow/react` ベースの呼び出しグラフ |
| `FunctionNode` / `FileNode` | グラフ上のノード。リング優先度: 選択(stone-900) > 新規循環(red-500) > 追加(emerald-500) > 変更(amber-400)。新規循環ノードは右下に赤い ↻ バッジ |
| `ClusterGroup` | 展開中クラスタの背景コンテナ。子ノードを `parentId` で内包 |
| `SuperClusterNode` | 折りたたみ中クラスタのスーパーノード。クリックで展開 |
| `FunctionDetailsPanel` | 右パネル。クラスタラベル・レイヤー・GitHub 該当行リンク・diff |
| `GraphSearch` | グラフ左上の検索ボックス。関数名/パッケージをインクリメンタル検索→候補選択で展開+選択+中央寄せ（`fitView`）。折りたたみ中クラスタ内のノードも親を開いて到達する |
| `GraphControls` | 右下フローティングパネル（frosted glass: `bg-white/90 backdrop-blur-md`）。ラベル付きセクションに分割: 表示モード切替（変更影響/全体）・クラスタモード（Louvain/Package/File）・展開/折畳アクションボタン・変更ノードステップナビゲーター。セグメントコントロールは pill 型（active=`bg-white shadow-sm`、inactive=`text-stone-500`）。ナビゲーターはレビュー優先度順に変更ノードを1つずつ巡回し、クラスタ自動展開+選択+中央寄せを行う |
| `FocusLegend` | フォーカスモード中に左上（検索ボックスの下）に出る凡例（呼び出し元/先の方向色と件数） |
| `DiffViewer` | `@monaco-editor/react` で before/after 表示 |
| `ClusterSidebar` | クラスタ一覧（実装位置: `layout/`） |

### 3.6 グラフのフォーカスモードとエッジ方向

- **エッジ矢印**: すべてのエッジに矢印マーカー（caller → callee の向き）を付ける。色は diff ステータスの stroke に揃える（added=emerald / removed=stone 破線 / existing=薄 stone）。
- **フォーカスモード**: 関数ノードを選択すると、直接の caller / callee と接続エッジだけを強調し、それ以外のノード・エッジを減光する（`DIMMED_OPACITY`）。Pane クリックで解除。
- **方向色**（`src/lib/graphFocus.ts` に集約。グラフ機能上の意味色として hex 直書きを許容）:
  - 呼び出し元（upstream caller, 選択ノードを呼ぶ側）: blue `#3b82f6`
  - 呼び出し先（downstream callee, 選択ノードが呼ぶ側）: orange `#f97316`
- フォーカス計算（1-hop 近傍）は純粋関数 `computeFocus` に切り出し、レイアウト/描画から分離している。

### 3.7 循環参照ハイライトとレビュー優先度

- **循環参照ハイライト**: 新規循環（`is_new`）に属するノードは red-500 リング + 右下 ↻ バッジ、その内部辺（同一循環内の from→to）は red-500 の太線で強調する。既存循環は強調しない（PR で新たに生じた構造リスクのみ可視化する方針）。クラスタ折りたたみ時は循環がスーパーノード内に隠れるため、展開して確認する（段階的開示）。`CycleAlert` は新規循環の索引としてクリックでフォーカスする導線を兼ねる。
- **レビュー優先度**: ノード選択時に詳細パネルへ `優先度 高/中/低` バッジを出す（グラフ上には出さない＝情報過多を避ける）。判定は `src/lib/reviewPriority.ts` の純粋関数: 高=新規循環に含まれる or（PR 変更ノードかつ被呼び出し数 `in-degree >= 3`）、中=その他の PR 変更ノード、低=未変更。色は 高=red / 中=amber / 低=stone。
- **GitHub 行リンク**: 詳細パネルから定義行の GitHub blob URL を新規タブで開く（`src/lib/githubLinks.ts`）。`removed` は base SHA、それ以外は head SHA を ref に使う。

### 3.8 変更影響ビュー（既定表示）

- 大規模グラフでは diff に絡まないノード/エッジがレビューでほぼ見られず、描画負荷と視覚ノイズの主因になる。これを避けるため、グラフは **既定で「変更影響のみ」** を表示する。
- 範囲: 変更ノード（`changed` または `diff_status !== 'existing'`）＋その **直接近傍 1-hop（呼び出し元・呼び出し先の両方向）**。いずれかの端点が変更ノードのエッジだけを残し、無関係なノード/エッジを描画前に落とす。孤立した変更ノードも残す。1-hop に限定するのはフォーカスモードと同様に情報過多を避けるため（そこから先は click-focus で辿る）。
- 全体グラフ（全体構造・レイヤリング違反の俯瞰）は `GraphControls` 右下のトグル（`変更影響のみ` / `全体`）で切替できる。切替時はクラスタ構成が変わるため選択・展開状態をリセットする。
- 絞り込みは純粋関数 `src/lib/impactFilter.ts` の `filterToImpact` に集約し、クラスタ集約前の生グラフ（`data.graph`）へ適用する（レイアウト/描画から分離）。

### 3.5 採用ライブラリ（外観に影響するもの）

- Tailwind CSS v4
- shadcn/ui（プリミティブ生成）
- `@xyflow/react` v12（グラフ）
- `@monaco-editor/react`（diff）

新規ライブラリ導入は議論してから。

### 3.8 ランディングページ（公開）

`/` は未認証でも見られる公開 LP（`pages/Landing.tsx`）。セクションは `components/landing/` に分割（`LandingNav` / `LandingHero` / `ClusterShowcase` / `HowItWorks` / `FeatureGrid` / `CtaStrip` / `LandingFooter`）。

- **認証で出し分け**: Hero と CTA strip は `useAuth()` を見て切替える。未ログイン → GitHub 連携ボタン（`/auth/github`）、ログイン済み → PR URL 入力フォーム（`PrInputForm`、`useAnalyzeMutation` + `parsePrUrl` に接続し `/analysis/:jobId` へ遷移）。
- **ルート**: `/`=LP（公開） / `/analyze`=PR 入力 Home（要認証） / `/analysis/:jobId`=結果（要認証） / `/login`。`RequireAuth` から外れる公開ページは `/` のみ。
- **実装**: Claude Design の handoff（`design/project/landing.jsx`）を Tailwind ユーティリティで再現。handoff の色は Tailwind 標準パレットにそのまま対応する（中立色=`stone`、ブランド=`teal`）。ボタンと入力欄は shadcn/ui の `Button`（`asChild` でリンク化）/ `Input` を流用。中央寄せコンテナは `components/landing/Wrap.tsx` に共通化。レスポンシブは Tailwind の `sm:`/`lg:` と `clamp()`（見出し）で担保。装飾の radial-gradient/mask とグラフ系 SVG の配色のみ inline style / リテラル hex（グラフ意味色の慣習に準拠）。
- **プロダクトマーク**: `components/branding/AppMark.tsx` が DiffGraph. のアプリアイコン（design "A4·4"：ノード群が 1 つの accent ノードへ V 字で収束する形）。`Brand.tsx` と `public/favicon.svg` もこれに統一。
- **コンテンツ整合方針**: LP の記述は実装済み機能に限定する。未実装機能・架空の統計値・存在しないプラン/ページへのリンクは載せない。
  - `ClusterShowcase`: 実際の 3 クラスタリングモード（Louvain / パッケージ / ファイル）を紹介
  - `FeatureGrid`: 影響範囲 hop 可視化・AST diff・新規循環参照検出の 3 機能
  - `HowItWorks`: PR 貼付 → 依存グラフ構築 → クラスタ構造把握の 3 ステップ
  - Footer: 実リンクのみ（GitHub / README / ページ内アンカー）

---

## 4. 変更時のチェックリスト

UI を変更するときは以下を確認する。

- [ ] 既存プリミティブ（`components/ui/`）で表現できないか確認した
- [ ] 新しい色を使うなら `index.css` のトークンか `clusterColors.ts` 経由にした
- [ ] loading / empty / error / success の 4 状態を考慮した
- [ ] mono にすべき文字列（ID・パス・コード）が sans になっていない
- [ ] 影響範囲が大きい変更は、本ドキュメントの該当セクションを更新した
- [ ] [`.claude/rules/frontend-style.md`](./.claude/rules/frontend-style.md) のコード規約に違反していない

---

## 5. 本ドキュメントの保守

機能や仕様が変わるとき:

1. 「3. 現在の実装」を実態に合わせて更新する（必須）
2. 「2. 規約」レベルの方針が変わったら、それも併せて更新する
3. 「1. 原則」が変わるのは方向転換のときだけ。慎重に
4. 大きな UI 変更の PR には本ドキュメントの更新を含めるのが望ましい
