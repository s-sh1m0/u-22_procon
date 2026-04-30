package domain

import "time"

// AnalysisID は解析結果の一意識別子
type AnalysisID string

// PRInfo はGitHub PRの基本情報
type PRInfo struct {
	Owner   string
	Repo    string
	Number  int
	Title   string
	BaseRef string
	HeadRef string
	HeadSHA string
}

// Analysis はPR解析の結果エンティティ
type Analysis struct {
	ID        AnalysisID
	PR        PRInfo
	Result    *ClusterResult
	CreatedAt time.Time
}

// ChangedFile はPRで変更されたファイル1件の情報。
// Patch は GitHub API が返す unified diff テキスト（バイナリや巨大ファイルでは空になり得る）。
type ChangedFile struct {
	Filename  string // リポジトリルートからの相対パス
	Status    string // "added" | "modified" | "removed" | "renamed"
	Additions int
	Deletions int
	Patch     string
}
