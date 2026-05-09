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
	BaseSHA string
}

// Analysis はPR解析の結果エンティティ
type Analysis struct {
	ID           AnalysisID
	PR           PRInfo
	Result       *ClusterResult
	ChangedFiles []DiffFile
	CreatedAt    time.Time
}

// ChangedFile はPRで変更されたファイル1件の情報。
// Patch は GitHub API が返す unified diff テキスト（バイナリや巨大ファイルでは空になり得る）。
type ChangedFile struct {
	Filename         string // リポジトリルートからの相対パス
	PreviousFilename string // renamed のとき base 側のパス
	Status           string // "added" | "modified" | "removed" | "renamed"
	Additions        int
	Deletions        int
	Patch            string
}

// DiffFile はPRで変更されたファイルの before/after 全文を保持する。
type DiffFile struct {
	Filename      string // head 側のパス（removed は base 側のパス）
	PreviousName  string // renamed のとき base 側のパス
	Status        string // "added" | "modified" | "removed" | "renamed"
	Additions     int
	Deletions     int
	BeforeContent string // base SHA でのファイル全文（added のとき ""）
	AfterContent  string // head SHA でのファイル全文（removed のとき ""）
}
