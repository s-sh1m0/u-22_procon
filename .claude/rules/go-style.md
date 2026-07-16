# Go コーディング規約

公式 [Effective Go](https://go.dev/doc/effective_go) と [Google Go Style Guide](https://google.github.io/styleguide/go/) に準拠する。
本ファイルでは **本プロジェクト固有の追加ルール** のみを記載する。

## 基本

- Go バージョン: `1.25.3`（README の技術スタックを正とする）
- フォーマッタ: `gofmt` / `goimports`（保存時自動）
- リンタ: `go vet` 必須。`golangci-lint` は導入予定

## ディレクトリ構成

- `backend/` 配下に Go コードを置く
- パッケージ命名: 短く、複数形を避ける（`models` ではなく `model`）

## 命名

- エクスポートする識別子は **必ず GoDoc コメント** を付ける
- エラー値: `var ErrXxx = errors.New(...)` の形で先頭定義
- インターフェースは利用側で小さく定義する（過度な事前抽象化はしない）

## エラーハンドリング

- エラーは握りつぶさない（`_` で捨てない）
- `fmt.Errorf("...: %w", err)` でラップしてコンテキストを足す
- `errors.Is` / `errors.As` で判定する

## 同時実行

- goroutine を起こすときは **必ず終了経路** を設計する（`context.Context` 経由でキャンセル可能に）
- `sync.WaitGroup` か `errgroup.Group` で待ち合わせる
- グローバル mutable 状態を避ける

## テスト

- `*_test.go` は対象ファイルと同一パッケージに置く
- テーブル駆動テストを基本とする
- 外部依存（GitHub API 等）はインタフェース化してモックする

## 本プロジェクト特有

- AST 解析: `golang.org/x/tools/go/packages` を使う（`tree-sitter` ではない）
- 呼び出しグラフ: `golang.org/x/tools/go/callgraph` を使う
- グラフ処理: `gonum.org/v1/gonum/graph/community`（Louvain 法）を使う
- HTTP: `labstack/echo/v4` のミドルウェア機構を活用する
