package analyzer

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGoflagsWithParallelism(t *testing.T) {
	p := "-p=" + strconv.Itoa(buildParallelism())
	tests := []struct {
		name     string
		existing string
		want     string
	}{
		{name: "GOFLAGS なし", existing: "", want: p},
		{name: "既存の GOFLAGS を残す", existing: "-mod=mod", want: "-mod=mod " + p},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := goflagsWithParallelism(tt.existing); got != tt.want {
				t.Errorf("goflagsWithParallelism(%q) = %q, want %q", tt.existing, got, tt.want)
			}
		})
	}
}

// analyzerEnv は GOFLAGS を 1 つだけ含み、既存の値に -p を足したものになること。
func TestAnalyzerEnv_GOFLAGS(t *testing.T) {
	t.Setenv("GOFLAGS", "-mod=mod")

	var got []string
	for _, kv := range analyzerEnv() {
		if v, ok := strings.CutPrefix(kv, "GOFLAGS="); ok {
			got = append(got, v)
		}
	}
	want := "-mod=mod -p=" + strconv.Itoa(buildParallelism())
	if len(got) != 1 || got[0] != want {
		t.Errorf("GOFLAGS entries = %q, want [%q]", got, want)
	}
}

func TestWarmExportData_Fixture(t *testing.T) {
	dir := writeFixture(t, map[string]string{
		"go.mod": "module example.com/fixture\n\ngo 1.21\n",
		"pkg/a/a.go": `package a

import (
	"strings"

	"example.com/fixture/pkg/b"
)

func F() string { return strings.Repeat("x", b.G()) }
`,
		"pkg/b/b.go": `package b

func G() int { return 42 }
`,
	})

	loader := NewGoPackageLoader()
	if err := loader.warmExportData(context.Background(), dir, []string{"example.com/fixture/pkg/a"}); err != nil {
		t.Fatalf("warmExportData: %v", err)
	}
}

// 別の warmExportData が実行中（セマフォ取得済み）のとき、待っている側は ctx の
// タイムアウトで諦めてエラーを返すこと。
func TestWarmExportData_TimeoutWhileWaiting(t *testing.T) {
	loader := NewGoPackageLoader()
	loader.buildSem <- struct{}{} // 実行中の別ジョブを模す
	defer func() { <-loader.buildSem }()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := loader.warmExportData(ctx, t.TempDir(), []string{"example.com/x"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}
