package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

func TestRefineChangedByContent(t *testing.T) {
	baseRoot := t.TempDir()
	headRoot := t.TempDir()

	writeFile := func(root, rel, content string) {
		t.Helper()
		abs := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// head では A と B の間に New を追加し、B の中身を変更している。
	// GitHub の diff は New と A の共通末尾（`return nil` / `}`）を揃えることがあり、
	// その場合 +行が A の末尾にかかって A も hunk 判定で changed になる。
	writeFile(baseRoot, "pkg/a/a.go", `package a

func A() error {
	x()
	return nil
}

func B() {
	y()
}
`)
	writeFile(headRoot, "pkg/a/a.go", `package a

func A() error {
	x()
	return nil
}

func New() error {
	z()
	return nil
}

func B() {
	y2()
}
`)
	renamed := `package a

func R() {
	r()
}
`
	writeFile(baseRoot, "pkg/a/old_name.go", renamed)
	writeFile(headRoot, "pkg/a/new_name.go", renamed)

	node := func(id, file string, line, endLine int, changed bool) domain.Node {
		return domain.Node{ID: domain.NodeID(id), Name: id, File: file, Line: line, EndLine: endLine, Changed: changed}
	}
	ptr := func(n domain.Node) *domain.Node { return &n }

	tests := []struct {
		name string
		head domain.Node
		base *domain.Node // nil なら base に同じ ID が無い
		want bool
	}{
		{
			name: "隣の関数に +行がかかっただけで中身は同じ",
			head: node("A", "pkg/a/a.go", 3, 6, true),
			base: ptr(node("A", "pkg/a/a.go", 3, 6, false)),
			want: false,
		},
		{
			name: "中身が変わった",
			head: node("B", "pkg/a/a.go", 13, 15, true),
			base: ptr(node("B", "pkg/a/a.go", 8, 10, false)),
			want: true,
		},
		{
			name: "base に無い（追加された関数）",
			head: node("New", "pkg/a/a.go", 8, 11, true),
			want: true,
		},
		{
			name: "ファイル rename のみで中身は同じ",
			head: node("R", "pkg/a/new_name.go", 3, 5, true),
			base: ptr(node("R", "pkg/a/old_name.go", 3, 5, false)),
			want: false,
		},
		{
			name: "EndLine 不明は hunk 判定を維持",
			head: node("S", "pkg/a/a.go", 3, 0, true),
			base: ptr(node("S", "pkg/a/a.go", 3, 6, false)),
			want: true,
		},
		{
			name: "Changed=false は中身が違っても true にしない",
			head: node("C", "pkg/a/a.go", 13, 15, false),
			base: ptr(node("C", "pkg/a/a.go", 8, 10, false)),
			want: false,
		},
		{
			name: "行範囲がファイル外は hunk 判定を維持",
			head: node("O", "pkg/a/a.go", 14, 99, true),
			base: ptr(node("O", "pkg/a/a.go", 3, 6, false)),
			want: true,
		},
		{
			name: "ファイルを読めないときは hunk 判定を維持",
			head: node("M", "pkg/a/missing.go", 1, 3, true),
			base: ptr(node("M", "pkg/a/missing.go", 1, 3, false)),
			want: true,
		},
	}

	head := &domain.Graph{}
	base := &domain.Graph{}
	for _, tt := range tests {
		head.Nodes = append(head.Nodes, tt.head)
		if tt.base != nil {
			base.Nodes = append(base.Nodes, *tt.base)
		}
	}

	RefineChangedByContent(head, base, headRoot, baseRoot)

	got := make(map[domain.NodeID]bool, len(head.Nodes))
	for _, n := range head.Nodes {
		got[n.ID] = n.Changed
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got[tt.head.ID] != tt.want {
				t.Errorf("node %q Changed=%v, want %v", tt.head.ID, got[tt.head.ID], tt.want)
			}
		})
	}
}

func TestRefineChangedByContent_NilGraph(t *testing.T) {
	head := &domain.Graph{Nodes: []domain.Node{{ID: "A", Changed: true}}}
	RefineChangedByContent(head, nil, "", "")
	if !head.Nodes[0].Changed {
		t.Error("base が nil のときは head を変更しない")
	}
	RefineChangedByContent(nil, head, "", "") // panic しないこと
}
