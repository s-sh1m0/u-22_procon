package analyzer

import (
	"strconv"
	"strings"
	"testing"
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
