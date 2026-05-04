package cluster

import (
	"strings"
	"testing"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

func TestLabelCluster_CommonPathPrefix(t *testing.T) {
	nodes := []domain.Node{
		{ID: "a", File: "pkg/foo/a.go"},
		{ID: "b", File: "pkg/foo/b.go"},
		{ID: "c", File: "pkg/foo/sub/c.go"},
	}
	label := labelCluster(0, nodes, nil)
	if !strings.Contains(label, "pkg/foo") {
		t.Errorf("want label to contain 'pkg/foo', got %q", label)
	}
}

func TestLabelCluster_PackageFallback(t *testing.T) {
	nodes := []domain.Node{
		{ID: "a", File: "/x/a.go", Package: "github.com/example/foo"},
		{ID: "b", File: "/y/b.go", Package: "github.com/example/foo"},
		{ID: "c", File: "/z/c.go", Package: "github.com/example/bar"},
	}
	label := labelCluster(0, nodes, nil)
	if !strings.Contains(label, "github.com/example/foo") {
		t.Errorf("want most-common package in label, got %q", label)
	}
}

func TestLabelCluster_BaseFallback(t *testing.T) {
	nodes := []domain.Node{
		{ID: "a", File: "/a/handler.go", Package: "pkga"},
		{ID: "b", File: "/b/handler.go", Package: "pkgb"},
		{ID: "c", File: "/c/handler.go", Package: "pkgc"},
	}
	label := labelCluster(0, nodes, nil)
	if !strings.Contains(label, "handler.go") {
		t.Errorf("want basename fallback 'handler.go', got %q", label)
	}
}

func TestLabelCluster_NoChangedNoSuffix(t *testing.T) {
	nodes := []domain.Node{
		{ID: "a", Name: "Foo", File: "pkg/a/a.go", Changed: false},
	}
	label := labelCluster(0, nodes, nil)
	if strings.Contains(label, "#") {
		t.Errorf("want no '#Func' suffix when Changed=false, got %q", label)
	}
}

func TestLabelCluster_RepresentativeChanged(t *testing.T) {
	// 3 ノード: a(Changed, 高次数), b(Changed, 低次数), c(Changed=false)
	nodes := []domain.Node{
		{ID: "a", Name: "High", File: "pkg/foo/a.go", Changed: true},
		{ID: "b", Name: "Low", File: "pkg/foo/b.go", Changed: true},
		{ID: "c", Name: "Unc", File: "pkg/foo/c.go", Changed: false},
	}
	// a はクラスタ内エッジ 2 本、b は 1 本
	edges := []domain.Edge{
		{From: "a", To: "b"},
		{From: "a", To: "c"},
	}
	label := labelCluster(0, nodes, edges)
	if !strings.Contains(label, "#High") {
		t.Errorf("want '#High' (highest degree Changed node), got %q", label)
	}
}
