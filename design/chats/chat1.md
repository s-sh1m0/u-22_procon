# PR依存関係可視化ツール

_Started 2026-04-30 12:42 UTC_

---

## User

プルリクエストの各モジュールの依存関係可視化ツールを作ろうとしている。
githubのoauthでログイン後、PRのURLを入力することで、そのリポジトリをcloneし、バックエンド側で解析した後、変更関数をdiffと突き合わせ、モジュールのグラフみたいなのを作成しようとしている。
また、各diffを4つのクラスタに分割して見やすくする機能も作りたい

## Assistant

プルリクエストの依存関係可視化ツールですね。面白いプロジェクトです！高品質な設計のために、いくつか確認させてください。

_[tool: questions_v2]_

## User

Questions answered:
- scope: ランディング含めて全部
- main_screen: モジュール依存グラフ（ノード&エッジ）が中心
- graph_style: 階層レイアウト（DAG、上から下）
- cluster_meaning: 変更の種類（追加/削除/修正/リファクタ）
- interactions: ノードクリックでdiff詳細サイドパネル
- tone: モダンSaaS（明るくクリーン）
- color: diff定番（緑/赤）+ 補色
- sample_data: 大きめ（30+ファイル）で密度を見せる
- variations: design_canvasで2-3案を並べて比較
- tweaks: クラスタ表示の有無, ノード密度・サイズ, サイドパネル位置（右/下/オーバーレイ）
- language: 日本語

## Assistant

ありがとうございます！方針が固まりました。design_canvasで3案並べて比較する構成で作っていきます。

## 設計方針の宣言

**アプローチ:** モダンSaaSトーン × DAG階層グラフを軸に、3つのバリエーションを並べて比較
- **案A: クラシックDAG** — GitHub寄り、上から下のレイヤー構造、サイドパネル右
- **案B: スイムレーン式** — 4クラスタ（追加/削除/修正/リファクタ）を縦レーンに分けて配置
- **案C: 集約サマリービュー** — クラスタ別カード + ミニグラフ、密度重視のダッシュボード

**カラー:** diff定番の緑(追加)/赤(削除)/橙(修正)/紫(リファクタ) + ニュートラル背景
**フォント:** Inter Tight (UI) + JetBrains Mono (コード) — モダンSaaS向け
**ランディング+ログイン+PR入力+解析中+結果** までフローを通します

_[tool: update_todos]_

_[tool: copy_starter_component]_

_[tool: copy_starter_component]_

_[tool: read_file]_

_[tool: snip]_

Now let me build the project. I'll create the data, components, then assemble the design canvas.

_[tool: set_project_title]_

_[tool: write_file]_

Now let me build the graph layout helper, landing/login/input screens, and the three main analysis variants.

_[tool: write_file]_

_[tool: snip]_

_[tool: write_file]_

_[tool: write_file]_

_[tool: snip]_

_[tool: write_file]_

_[tool: snip]_

_[tool: write_file]_

_[tool: snip]_

_[tool: write_file]_

_[tool: snip]_

_[tool: done]_

_[tool: fork_verifier_agent]_

_[tool: snip]_

完成しました！design_canvas上に3案を並べた構成で、フローと結果ビューを比較できるようにしました。

## 構成

**ユーザーフロー (1080×780)**
- ① ランディング — ヒーロー + ミニ依存グラフ + 4クラスタ説明
- ② PR入力 (ログイン済み) — URL入力 + 履歴
- ③ 解析中 — 5ステップの進捗表示

**解析結果 3案 (1280×820)**
- **案A · クラシックDAG** — 4レイヤ階層レイアウト + 右サイドパネルでdiff/関数一覧。クラスタフィルタ・ノードクリックで連結ノード以外をディム
- **案B · スイムレーン** — 4クラスタを縦レーンに分け、レイヤを横ライン区切り。レーンを跨ぐ依存は破線
- **案C · ダッシュボード型** — 4象限グリッド + ヒート強度バー + 関数スパークバー。密度を一覧で見る

**カラー:** diff定番(緑+/赤−/橙~/紫↻) + teal-700ブランド  
**Tweaks:** クラスタ色、レイヤラベル、ノード密度、サイドパネル位置

各アートボードは右上ボタンでフォーカス表示できます。次のステップとしては、選択した案を深掘り（案Aと案Bのハイブリッド等）したり、解析失敗・空PR・大規模PRなどのエッジケース画面を追加できます。

## User

diffによる影響範囲の可視化も行いたい

## Assistant

_[tool: snip]_

影響範囲の可視化、いいですね。「変更されたモジュールから何段先まで影響が及ぶか」を見せる機能を追加します。

設計方針：
- **影響範囲 = 変更ノードから下流（依存される側）への伝播深さ**
- 各ノードに「影響度（直接/2hop/3hop+/影響なし）」を計算
- 案Aに「影響範囲モード」をトグル追加 — 選択モジュールの波紋を可視化
- 新しい案D（影響範囲ヒートマップ）を追加して比較できるように

_[tool: write_file]_

_[tool: snip]_

_[tool: str_replace_edit]_

_[tool: str_replace_edit]_

_[tool: snip]_

_[tool: done]_

_[tool: snip]_

_[tool: fork_verifier_agent]_

影響範囲ビュー（案D）を追加しました。

**新セクション「影響範囲の可視化」**

- **2モード切替**
  - 「変更全体の影響」: PR内の全変更モジュールを起点に、依存元へ何hop伝播するかを一気に可視化
  - 「選択モジュールの波紋」: 特定モジュールを選んで、そこから広がる影響範囲だけを見る

- **可視化要素**
  - ノードに hop バッジ（0=変更元 / 1=直接 / 2=2hop / 3+=遠い）
  - hop≥1のノードに点線の波紋リング
  - 影響を受けないノードはグレーで沈める
  - エッジも影響伝播するものだけ強調表示

- **上部の影響度分布バー**: 全モジュールのうち何件が直接/2hop/3hop+影響を受けるかを比率表示

- **右パネル**: 依存元/依存先の数、hop別の影響先一覧、関数変更詳細

これで「このPRはマージするとどこまで影響するか」がパッと見えるようになります。

