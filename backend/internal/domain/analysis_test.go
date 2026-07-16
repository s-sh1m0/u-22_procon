package domain

import "testing"

func TestNormalizeChangedForBase(t *testing.T) {
	tests := []struct {
		name    string
		changed []ChangedFile
		want    []ChangedFile
	}{
		{
			name:    "空リストは空を返す",
			changed: nil,
			want:    []ChangedFile{},
		},
		{
			name: "added は base 側に存在しないため除外",
			changed: []ChangedFile{
				{Filename: "pkg/a/new.go", Status: FileStatusAdded},
				{Filename: "pkg/b/mod.go", Status: FileStatusModified},
			},
			want: []ChangedFile{
				{Filename: "pkg/b/mod.go", Status: FileStatusModified},
			},
		},
		{
			name: "renamed は Filename を PreviousFilename に置換",
			changed: []ChangedFile{
				{Filename: "pkg/a/new.go", PreviousFilename: "pkg/a/old.go", Status: FileStatusRenamed},
			},
			want: []ChangedFile{
				{Filename: "pkg/a/old.go", PreviousFilename: "pkg/a/old.go", Status: FileStatusRenamed},
			},
		},
		{
			name: "renamed でも PreviousFilename が空なら Filename を維持",
			changed: []ChangedFile{
				{Filename: "pkg/a/new.go", Status: FileStatusRenamed},
			},
			want: []ChangedFile{
				{Filename: "pkg/a/new.go", Status: FileStatusRenamed},
			},
		},
		{
			name: "modified と removed はそのまま",
			changed: []ChangedFile{
				{Filename: "pkg/a/mod.go", Status: FileStatusModified},
				{Filename: "pkg/b/rm.go", Status: FileStatusRemoved},
			},
			want: []ChangedFile{
				{Filename: "pkg/a/mod.go", Status: FileStatusModified},
				{Filename: "pkg/b/rm.go", Status: FileStatusRemoved},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeChangedForBase(tt.changed)
			if len(got) != len(tt.want) {
				t.Fatalf("len=%d want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i].Filename != tt.want[i].Filename {
					t.Errorf("[%d] Filename=%q want %q", i, got[i].Filename, tt.want[i].Filename)
				}
				if got[i].PreviousFilename != tt.want[i].PreviousFilename {
					t.Errorf("[%d] PreviousFilename=%q want %q", i, got[i].PreviousFilename, tt.want[i].PreviousFilename)
				}
				if got[i].Status != tt.want[i].Status {
					t.Errorf("[%d] Status=%q want %q", i, got[i].Status, tt.want[i].Status)
				}
			}
		})
	}
}
