package api

import "github.com/s-sh1m0/u-22_procon/backend/internal/domain"

// AnalyzeRequest は POST /api/analyze のリクエスト
type AnalyzeRequest struct {
	PRURL string `json:"pr_url"`
}

// JobResponse は GET /api/jobs/:id のレスポンス
type JobResponse struct {
	JobID      string `json:"job_id"`
	Status     string `json:"status"`
	AnalysisID string `json:"analysis_id,omitempty"`
	Error      string `json:"error,omitempty"`
}

// MeResponse は GET /auth/me のレスポンス
type MeResponse struct {
	Login string `json:"login"`
}

// PRInfoDTO は PR 基本情報の JSON 表現
type PRInfoDTO struct {
	Owner   string `json:"owner"`
	Repo    string `json:"repo"`
	Number  int    `json:"number"`
	Title   string `json:"title"`
	BaseRef string `json:"base_ref"`
	HeadRef string `json:"head_ref"`
}

// GraphResponse は GET /api/graph/:jobId のレスポンス
type GraphResponse struct {
	PR       PRInfoDTO    `json:"pr"`
	Clusters []ClusterDTO `json:"clusters"`
	Graph    GraphDTO     `json:"graph"`
	Cycles   []CycleDTO   `json:"cycles"`
}

// CycleDTO は循環参照の JSON 表現
type CycleDTO struct {
	ID    int      `json:"id"`
	Nodes []string `json:"nodes"`
	IsNew bool     `json:"is_new"`
}

// ClusterDTO はクラスタの JSON 表現
type ClusterDTO struct {
	ID    int      `json:"id"`
	Label string   `json:"label"`
	Nodes []string `json:"nodes"`
}

// GraphDTO はグラフの JSON 表現
type GraphDTO struct {
	Nodes []NodeDTO `json:"nodes"`
	Edges []EdgeDTO `json:"edges"`
}

// NodeDTO はノードの JSON 表現
type NodeDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Package    string `json:"package"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Changed    bool   `json:"changed"`
	DiffStatus string `json:"diff_status"` // "added" | "removed" | "existing"
}

// EdgeDTO はエッジの JSON 表現
type EdgeDTO struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Status string `json:"status"` // "added" | "removed" | "existing"
}

// DiffResponse は GET /api/diff/:jobId のレスポンス
type DiffResponse struct {
	Files []DiffFileDTO `json:"files"`
}

// DiffFileDTO はファイル単位の diff 情報の JSON 表現
type DiffFileDTO struct {
	Filename      string `json:"filename"`
	PreviousName  string `json:"previous_name,omitempty"`
	Status        string `json:"status"`
	Additions     int    `json:"additions"`
	Deletions     int    `json:"deletions"`
	BeforeContent string `json:"before_content"`
	AfterContent  string `json:"after_content"`
}

// toDiffResponse は domain.Analysis を DiffResponse に変換する。
func toDiffResponse(a *domain.Analysis) DiffResponse {
	files := make([]DiffFileDTO, len(a.ChangedFiles))
	for i, f := range a.ChangedFiles {
		files[i] = DiffFileDTO{
			Filename:      f.Filename,
			PreviousName:  f.PreviousName,
			Status:        f.Status,
			Additions:     f.Additions,
			Deletions:     f.Deletions,
			BeforeContent: f.BeforeContent,
			AfterContent:  f.AfterContent,
		}
	}
	return DiffResponse{Files: files}
}

// toGraphResponse は domain.Analysis を GraphResponse に変換する。
func toGraphResponse(a *domain.Analysis) GraphResponse {
	r := a.Result
	clusters := make([]ClusterDTO, len(r.Clusters))
	for i, c := range r.Clusters {
		nodes := make([]string, len(c.Nodes))
		for j, n := range c.Nodes {
			nodes[j] = string(n)
		}
		clusters[i] = ClusterDTO{ID: c.ID, Label: c.Label, Nodes: nodes}
	}

	nodes := make([]NodeDTO, len(r.Graph.Nodes))
	for i, n := range r.Graph.Nodes {
		ds := string(n.DiffStatus)
		if ds == "" {
			ds = string(domain.DiffStatusExisting)
		}
		nodes[i] = NodeDTO{
			ID:         string(n.ID),
			Name:       n.Name,
			Package:    n.Package,
			File:       n.File,
			Line:       n.Line,
			Changed:    n.Changed,
			DiffStatus: ds,
		}
	}

	edges := make([]EdgeDTO, len(r.Graph.Edges))
	for i, e := range r.Graph.Edges {
		st := string(e.Status)
		if st == "" {
			st = string(domain.DiffStatusExisting)
		}
		edges[i] = EdgeDTO{From: string(e.From), To: string(e.To), Status: st}
	}

	cycles := make([]CycleDTO, len(r.Cycles))
	for i, c := range r.Cycles {
		nodeIDs := make([]string, len(c.Nodes))
		for j, n := range c.Nodes {
			nodeIDs[j] = string(n)
		}
		cycles[i] = CycleDTO{ID: c.ID, Nodes: nodeIDs, IsNew: c.IsNew}
	}

	return GraphResponse{
		PR: PRInfoDTO{
			Owner:   a.PR.Owner,
			Repo:    a.PR.Repo,
			Number:  a.PR.Number,
			Title:   a.PR.Title,
			BaseRef: a.PR.BaseRef,
			HeadRef: a.PR.HeadRef,
		},
		Clusters: clusters,
		Graph:    GraphDTO{Nodes: nodes, Edges: edges},
		Cycles:   cycles,
	}
}
