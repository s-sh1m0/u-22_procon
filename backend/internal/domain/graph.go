package domain

// NodeID はグラフノードの一意識別子
type NodeID string

// Node は関数・メソッドをグラフノードとして表す
type Node struct {
	ID      NodeID
	Name    string // 関数名
	Package string // パッケージパス
	File    string // ファイルパス
	Line    int    // 定義行
	Changed bool   // このPRで変更されたか
}

// Edge はノード間の呼び出し関係を表す
type Edge struct {
	From NodeID
	To   NodeID
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

// ClusterResult はクラスタリング結果
type ClusterResult struct {
	Clusters []Cluster
	Graph    Graph
}
