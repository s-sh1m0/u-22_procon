package api

// TODO: リクエスト・レスポンスの JSON 構造体定義

// AnalyzeRequest は POST /api/analyses のリクエスト
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
