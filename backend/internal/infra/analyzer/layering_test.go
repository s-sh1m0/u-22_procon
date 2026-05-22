package analyzer

import (
	"testing"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

func pkgNode(id, pkg string) domain.Node {
	return domain.Node{ID: domain.NodeID(id), Package: pkg}
}

func TestDetectLayeringViolations(t *testing.T) {
	const (
		domainPkg = "github.com/x/y/internal/domain"
		usecase   = "github.com/x/y/internal/usecase"
		infra     = "github.com/x/y/internal/infra/store"
		api       = "github.com/x/y/internal/api"
		pkgUtil   = "github.com/x/y/internal/pkg/logger"
	)

	tests := []struct {
		name  string
		graph *domain.Graph
		want  []domain.LayerViolation
	}{
		{
			name:  "nil graph",
			graph: nil,
			want:  nil,
		},
		{
			name: "domain to infra is a violation",
			graph: &domain.Graph{
				Nodes: []domain.Node{pkgNode("d", domainPkg), pkgNode("i", infra)},
				Edges: []domain.Edge{{From: "d", To: "i", Status: domain.DiffStatusAdded}},
			},
			want: []domain.LayerViolation{
				{From: "d", To: "i", FromLayer: "domain", ToLayer: "infra", IsNew: true},
			},
		},
		{
			name: "usecase to infra is a violation (existing edge -> not new)",
			graph: &domain.Graph{
				Nodes: []domain.Node{pkgNode("u", usecase), pkgNode("i", infra)},
				Edges: []domain.Edge{{From: "u", To: "i", Status: domain.DiffStatusExisting}},
			},
			want: []domain.LayerViolation{
				{From: "u", To: "i", FromLayer: "domain", ToLayer: "infra", IsNew: false},
			},
		},
		{
			name: "infra to domain is allowed (outer depends inward)",
			graph: &domain.Graph{
				Nodes: []domain.Node{pkgNode("i", infra), pkgNode("d", domainPkg)},
				Edges: []domain.Edge{{From: "i", To: "d", Status: domain.DiffStatusAdded}},
			},
			want: []domain.LayerViolation{},
		},
		{
			name: "api to infra is allowed (outer to inner)",
			graph: &domain.Graph{
				Nodes: []domain.Node{pkgNode("a", api), pkgNode("i", infra)},
				Edges: []domain.Edge{{From: "a", To: "i", Status: domain.DiffStatusExisting}},
			},
			want: []domain.LayerViolation{},
		},
		{
			name: "infra to api is a violation (infra must not reach ui)",
			graph: &domain.Graph{
				Nodes: []domain.Node{pkgNode("i", infra), pkgNode("a", api)},
				Edges: []domain.Edge{{From: "i", To: "a", Status: domain.DiffStatusAdded}},
			},
			want: []domain.LayerViolation{
				{From: "i", To: "a", FromLayer: "infra", ToLayer: "ui", IsNew: true},
			},
		},
		{
			name: "pkg util edges are ignored",
			graph: &domain.Graph{
				Nodes: []domain.Node{pkgNode("d", domainPkg), pkgNode("p", pkgUtil)},
				Edges: []domain.Edge{{From: "d", To: "p", Status: domain.DiffStatusAdded}},
			},
			want: []domain.LayerViolation{},
		},
		{
			name: "same layer is allowed",
			graph: &domain.Graph{
				Nodes: []domain.Node{pkgNode("d1", domainPkg), pkgNode("d2", usecase)},
				Edges: []domain.Edge{{From: "d1", To: "d2", Status: domain.DiffStatusAdded}},
			},
			want: []domain.LayerViolation{},
		},
		{
			name: "results sorted by From then To",
			graph: &domain.Graph{
				Nodes: []domain.Node{pkgNode("d", domainPkg), pkgNode("i1", infra), pkgNode("i2", infra)},
				Edges: []domain.Edge{
					{From: "d", To: "i2", Status: domain.DiffStatusAdded},
					{From: "d", To: "i1", Status: domain.DiffStatusExisting},
				},
			},
			want: []domain.LayerViolation{
				{From: "d", To: "i1", FromLayer: "domain", ToLayer: "infra", IsNew: false},
				{From: "d", To: "i2", FromLayer: "domain", ToLayer: "infra", IsNew: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectLayeringViolations(tt.graph)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d violations, want %d: %+v", len(got), len(tt.want), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("violation[%d] = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
