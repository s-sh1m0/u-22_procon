package analyzer

import (
	"testing"
)

func TestChangedLines_FunctionChanged(t *testing.T) {
	cl := ChangedLines{
		"/repo/a.go": {
			{Start: 10, End: 15},
			{Start: 30, End: 35},
		},
		"/repo/b.go": nil, // Patch 空 → ファイル単位フォールバック
		"/repo/d.go": {
			{Start: 21, End: 20}, // 20 行目と 21 行目の間で削除
		},
		"/repo/e.go": {}, // Patch はあるが変更位置なし
	}

	tests := []struct {
		name  string
		file  string
		start int
		end   int
		want  bool
	}{
		{"overlap first range", "/repo/a.go", 8, 12, true},
		{"overlap second range", "/repo/a.go", 33, 40, true},
		{"exact match", "/repo/a.go", 10, 15, true},
		{"inside range", "/repo/a.go", 11, 14, true},
		{"between ranges", "/repo/a.go", 16, 29, false},
		{"before all ranges", "/repo/a.go", 1, 5, false},
		{"after all ranges", "/repo/a.go", 40, 50, false},
		{"nil ranges (fallback)", "/repo/b.go", 1, 100, true},
		{"unknown file", "/repo/c.go", 1, 100, false},
		{"deletion inside function", "/repo/d.go", 15, 25, true},
		{"deletion right after closing brace", "/repo/d.go", 10, 20, false},
		{"deletion right before declaration", "/repo/d.go", 21, 30, false},
		{"empty ranges (no fallback)", "/repo/e.go", 1, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cl.FunctionChanged(tt.file, tt.start, tt.end)
			if got != tt.want {
				t.Errorf("FunctionChanged(%q, %d, %d) = %v, want %v", tt.file, tt.start, tt.end, got, tt.want)
			}
		})
	}
}

func TestChangedLines_FileChanged(t *testing.T) {
	cl := ChangedLines{
		"/repo/a.go": {{Start: 10, End: 15}},
		"/repo/b.go": nil,
	}

	tests := []struct {
		file string
		want bool
	}{
		{"/repo/a.go", true},
		{"/repo/b.go", true},
		{"/repo/c.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			got := cl.FileChanged(tt.file)
			if got != tt.want {
				t.Errorf("FileChanged(%q) = %v, want %v", tt.file, got, tt.want)
			}
		})
	}
}
