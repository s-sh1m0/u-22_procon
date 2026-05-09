package cluster

import (
	"testing"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

func TestLabelCluster_Empty(t *testing.T) {
	got := labelCluster(7, nil, nil)
	if got != "cluster 7" {
		t.Errorf("empty cluster: got %q, want %q", got, "cluster 7")
	}
}

func TestLabelCluster_BasicFormat(t *testing.T) {
	nodes := []domain.Node{
		{ID: "a", Name: "AnalyzePR", Package: "example.com/internal/usecase"},
	}
	got := labelCluster(0, nodes, nil)
	want := "usecase.AnalyzePR"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLabelCluster_PrefersChangedOverHigherInDegree(t *testing.T) {
	// b は in-degree が高いが Changed=false。a が Changed=true。
	// → 「Changed 優先」により a が代表になる。
	nodes := []domain.Node{
		{ID: "a", Name: "Important", Package: "pkg", Changed: true},
		{ID: "b", Name: "Hub", Package: "pkg", Changed: false},
	}
	edges := []domain.Edge{
		{From: "a", To: "b"},
		{From: "ext", To: "b"},
	}
	got := labelCluster(0, nodes, edges)
	want := "pkg.Important"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLabelCluster_PrefersExportedAmongChanged(t *testing.T) {
	// 両方 Changed だが a だけが exported。
	nodes := []domain.Node{
		{ID: "a", Name: "Public", Package: "pkg", Changed: true},
		{ID: "b", Name: "private", Package: "pkg", Changed: true},
	}
	edges := []domain.Edge{
		{From: "x", To: "b"}, // private の方が in-degree 高いが exported 優先
	}
	got := labelCluster(0, nodes, edges)
	want := "pkg.Public"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLabelCluster_PicksHighestInDegree(t *testing.T) {
	// 全員 Changed かつ exported。in-degree 最大の c が代表になる。
	nodes := []domain.Node{
		{ID: "a", Name: "AAA", Package: "pkg", Changed: true},
		{ID: "b", Name: "BBB", Package: "pkg", Changed: true},
		{ID: "c", Name: "CCC", Package: "pkg", Changed: true},
	}
	edges := []domain.Edge{
		{From: "a", To: "c"},
		{From: "b", To: "c"},
		{From: "a", To: "b"},
	}
	got := labelCluster(0, nodes, edges)
	want := "pkg.CCC"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLabelCluster_TieBreakAlphabetical(t *testing.T) {
	// 全員 in-degree 0、同条件。アルファベット昇順で Alpha が選ばれる。
	nodes := []domain.Node{
		{ID: "a", Name: "Beta", Package: "pkg"},
		{ID: "b", Name: "Alpha", Package: "pkg"},
		{ID: "c", Name: "Gamma", Package: "pkg"},
	}
	got := labelCluster(0, nodes, nil)
	want := "pkg.Alpha"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLabelCluster_AllUnchangedFallsThrough(t *testing.T) {
	// Changed=false しかない。フィルタを通り抜けて全ノードから選ぶ。
	nodes := []domain.Node{
		{ID: "a", Name: "Foo", Package: "pkg", Changed: false},
		{ID: "b", Name: "Bar", Package: "pkg", Changed: false},
	}
	edges := []domain.Edge{
		{From: "x", To: "a"}, // a の in-degree が高い
	}
	got := labelCluster(0, nodes, edges)
	want := "pkg.Foo"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLabelCluster_AllAnonymousFallsBackToPackage(t *testing.T) {
	// 名前が全部空 → 短縮パッケージ名のみ
	nodes := []domain.Node{
		{ID: "a", Name: "", Package: "github.com/example/pkg/foo"},
		{ID: "b", Name: "", Package: "github.com/example/pkg/foo"},
	}
	got := labelCluster(0, nodes, nil)
	want := "foo"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLabelCluster_NoPackageJustName(t *testing.T) {
	// パッケージ情報なし → 関数名のみ
	nodes := []domain.Node{
		{ID: "a", Name: "Solo", Package: ""},
	}
	got := labelCluster(0, nodes, nil)
	want := "Solo"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLabelCluster_NoPackageNoName(t *testing.T) {
	// パッケージも関数名も空 → fallback
	nodes := []domain.Node{
		{ID: "a", Name: "", Package: ""},
	}
	got := labelCluster(3, nodes, nil)
	want := "cluster 3"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
