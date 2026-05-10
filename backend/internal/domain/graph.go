package domain

// NodeID はグラフノードの一意識別子
type NodeID string

// DiffStatus はPR差分におけるノード/エッジの状態を表す。
type DiffStatus string

const (
	// DiffStatusExisting は base/head 両方に存在する。
	DiffStatusExisting DiffStatus = "existing"
	// DiffStatusAdded は head にのみ存在する（PRで追加された）。
	DiffStatusAdded DiffStatus = "added"
	// DiffStatusRemoved は base にのみ存在する（PRで削除された）。
	DiffStatusRemoved DiffStatus = "removed"
)

// Node は関数・メソッドをグラフノードとして表す
type Node struct {
	ID         NodeID
	Name       string     // 関数名
	Package    string     // パッケージパス
	File       string     // ファイルパス
	Line       int        // 定義行
	Changed    bool       // このPRで変更されたか（ファイル単位）
	DiffStatus DiffStatus // PR差分におけるノードの状態
}

// Edge はノード間の呼び出し関係を表す
type Edge struct {
	From   NodeID
	To     NodeID
	Status DiffStatus // PR差分におけるエッジの状態
}

// Graph は呼び出しグラフ全体
type Graph struct {
	Nodes []Node
	Edges []Edge
}

// Cluster はLouvainによって検出されたノードのグループ
type Cluster struct {
	ID    int
	Label string
	Nodes []NodeID
}

// Cycle はコールグラフ上の循環参照（強連結成分）を表す。
// IsNew が true の場合、Base には存在せず Head で新たに導入された循環。
type Cycle struct {
	ID    int      // 循環の連番
	Nodes []NodeID // 循環を構成するノード（最低 2 件、ID 昇順でソート済み）
	IsNew bool     // PR で新規に発生した循環か
}

// ClusterResult はクラスタリング結果
type ClusterResult struct {
	Clusters []Cluster
	Graph    Graph
	Cycles   []Cycle // 検出された循環参照（IsNew で新規/既存を区別）
}
