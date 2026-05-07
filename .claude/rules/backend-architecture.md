# バックエンド アーキテクチャ規約

**対象: `backend/` 配下の Go コードのみ**（フロントエンドは対象外）。

軽量オニオンアーキテクチャを採用する。厳格 DDD ではないが、
**リポジトリ層の DIP は徹底する**（単体テスト容易性のため）。

---

## 1. レイヤー構成

```
backend/internal/
├── domain/      ← 中心: エンティティ + リポジトリ interface
├── usecase/     ← 業務ロジック（domain にのみ依存）
├── api/         ← 入り口: handler / router / dto / middleware
└── infra/       ← 外側: domain interface の実装群
    ├── store/      （SQLite 永続化）
    ├── github/     （GitHub API クライアント）
    ├── analyzer/   （AST / callgraph パーサ）
    ├── cluster/    （Louvain クラスタリング）
    └── job/        （ジョブキュー / worker）
```

`cmd/server/main.go` が DI コンテナの役割を担う（infra 実装を usecase に注入する）。

> **注意**: `internal/` 直下に `analyzer/`, `cluster/`, `github/`, `job/`, `handler/` という空ディレクトリが存在するが、これらは Go ファイルを含まない残骸。コードは必ず `infra/<パッケージ>/` 配下に置くこと。

---

## 2. 依存方向（厳守）

```
api ──→ usecase ──→ domain ←── infra
```

- **`domain` は何にも依存しない**（標準ライブラリのみ可）
- **`usecase` は `domain` のみに依存**（infra を直接 import 禁止）
- **`infra` は `domain` の interface を実装する**（usecase / api を import 禁止）
- **`api` は `usecase` に依存**（infra を直接呼ばない）

違反例:
- ❌ `usecase` から `infra/store` を直接 import
- ❌ `domain` から `infra` を import
- ❌ `api` から SQLite を直接叩く

---

## 3. リポジトリ DIP（最重要）

### interface は `domain/` に置く

```go
// backend/internal/domain/repository.go
type AnalysisRepository interface {
    Save(ctx context.Context, a *Analysis) error
    FindByID(ctx context.Context, id AnalysisID) (*Analysis, error)
}
```

### 実装は `infra/store/` に置く

```go
// backend/internal/infra/store/analysis_repo.go
type AnalysisRepo struct { db *sql.DB }
func (r *AnalysisRepo) Save(ctx context.Context, a *domain.Analysis) error { ... }
```

### usecase は interface に依存させる

```go
// backend/internal/usecase/analyze_pr.go
type AnalyzePRUseCase struct {
    analysisRepo domain.AnalysisRepository  // ← interface を持つ
    jobRepo      domain.JobRepository
}
```

**`*store.AnalysisRepo` のような具象型を usecase が持ってはいけない。**

---

## 4. 命名規約

| レイヤー | パッケージ名 | 型名の例 |
|---|---|---|
| domain | `domain` | `Analysis`, `Job`, `AnalysisRepository`(I/F) |
| usecase | `usecase` | `AnalyzePRUseCase` |
| api | `api` | `AnalysisHandler`, `Router` |
| infra/store | `store` | `AnalysisRepo`(I/F の実装) |
| infra/github | `github` | `Client`, `OAuthHandler` |

- domain の interface 名: `XxxRepository`
- infra/store の実装名: `XxxRepo`（`Repository` サフィックスは interface 用に予約）

---

## 5. 単体テスト方針

- **usecase のテスト**: domain interface の **モック** を注入してテストする
  - `gomock` か手書きフェイク（小規模なら手書き推奨）
  - DB を立ち上げない
- **infra/store のテスト**: 実 SQLite に対して結合テスト
  - `modernc.org/sqlite` の in-memory モード（`:memory:`）を使う
- **api のテスト**: usecase をモック化してハンドラ単体でテスト

---

## 6. 新規コード追加時のチェックリスト

- [ ] 新しい外部依存（DB テーブル、外部 API）は `infra/<新パッケージ>/` に配置したか
- [ ] usecase が必要とする操作は `domain/repository.go` の interface に追加したか
- [ ] usecase は具象型ではなく interface を受け取っているか
- [ ] `cmd/server/main.go` の DI 配線を更新したか
- [ ] usecase の単体テストを書いたか（モック注入で）

---

## 7. 非同期ジョブキューフロー

API リクエストから結果取得までの流れ:

```
POST /api/analyze
  → AnalysisHandler.Analyze
  → AnalyzePRUseCase（Job を DB 保存 → job.Queue.Enqueue）
  → 202 Accepted + jobID 返却

job.Worker.Run（goroutine）
  → Queue から取り出し → 解析実行 → Job/Analysis を DB 更新

GET /api/jobs/:id   ← クライアントがポーリング
  → JobHandler.Get → pending / running / done / error を返す

GET /api/graph/:jobId
  → 解析結果の GraphResponse を返す（done になってから呼ぶ）
```

**重要な制約**: GitHub の OAuth Token はメモリ上の `job.Queue` にのみ存在し DB には保存しない。サーバー再起動時に pending/running だったジョブは `error` に遷移する（起動時に `MarkStaleJobsError` で更新）。ユーザーに再送信を促すこと。

---

## 8. 例外

- ロギング、メトリクス、`context.Context` の操作などの「インフラ横断的関心事」は
  ヘルパーパッケージ（例: `internal/pkg/logger`）として切り出してよい。
  これは layered architecture の「supporting」扱いで、依存方向ルールの例外とする。
- 軽量な値オブジェクト（ID 型、列挙型など）は `domain` に置く。
