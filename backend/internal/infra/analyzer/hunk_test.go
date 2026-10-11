package analyzer

import (
	"testing"
)

func TestParseHunks_Empty(t *testing.T) {
	got := ParseHunks("")
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestParseHunks_SingleHunk(t *testing.T) {
	patch := `@@ -10,6 +10,8 @@ func Foo() {
 	unchanged1
 	unchanged2
+	added1
+	added2
 	unchanged3
 	unchanged4`

	got := ParseHunks(patch)
	want := []ChangedLineRange{{Start: 12, End: 13}}
	assertRanges(t, got, want)
}

func TestParseHunks_MultipleHunks(t *testing.T) {
	patch := `@@ -5,4 +5,5 @@ package main
 func A() {
+	newLineInA
 }

@@ -20,3 +21,4 @@ func B() {
 	existing
+	newLineInB
 }`

	got := ParseHunks(patch)
	want := []ChangedLineRange{
		{Start: 6, End: 6},
		{Start: 22, End: 22},
	}
	assertRanges(t, got, want)
}

func TestParseHunks_AddedFile(t *testing.T) {
	patch := `@@ -0,0 +1,5 @@
+package main
+
+func New() {
+	return
+}`

	got := ParseHunks(patch)
	want := []ChangedLineRange{{Start: 1, End: 5}}
	assertRanges(t, got, want)
}

func TestParseHunks_DeleteOnly(t *testing.T) {
	patch := `@@ -10,5 +10,3 @@ func Foo() {
 	keep
-	removed1
-	removed2
 	keep2`

	// 削除位置（HEAD の 10 行目と 11 行目の間）を空範囲で返す。
	// nil を返すとファイル単位フォールバックになり全関数が changed になるため、非 nil であることも確認する。
	got := ParseHunks(patch)
	if got == nil {
		t.Fatal("expected non-nil ranges for delete-only hunk, got nil")
	}
	want := []ChangedLineRange{{Start: 11, End: 10}}
	assertRanges(t, got, want)
}

func TestParseHunks_DeleteAtEndOfHunk(t *testing.T) {
	patch := `@@ -10,3 +10,1 @@ func Foo() {
 	keep
-	removed1
-	removed2`

	got := ParseHunks(patch)
	want := []ChangedLineRange{{Start: 11, End: 10}}
	assertRanges(t, got, want)
}

func TestParseHunks_DeleteThenAddSeparately(t *testing.T) {
	patch := `@@ -1,5 +1,5 @@
 line1
-removed
 line2
+added
 line3`

	got := ParseHunks(patch)
	want := []ChangedLineRange{
		{Start: 2, End: 1},
		{Start: 3, End: 3},
	}
	assertRanges(t, got, want)
}

func TestParseHunks_NoNewlineMarker(t *testing.T) {
	// "\ No newline at end of file" は行番号を進めないので、後続の +行は 2 行目から始まる。
	patch := `@@ -1,2 +1,4 @@
 line1
-}
\ No newline at end of file
+}
+
+func New() {}`

	got := ParseHunks(patch)
	want := []ChangedLineRange{{Start: 2, End: 4}}
	assertRanges(t, got, want)
}

func TestParseHunks_MixedAddDelete(t *testing.T) {
	patch := `@@ -10,5 +10,5 @@ func Foo() {
 	keep
-	oldLine
+	newLine
 	keep2
 	keep3`

	got := ParseHunks(patch)
	want := []ChangedLineRange{{Start: 11, End: 11}}
	assertRanges(t, got, want)
}

func TestParseHunks_ConsecutiveAdditions(t *testing.T) {
	patch := `@@ -1,3 +1,7 @@
 line1
+add1
+add2
 line2
+add3
+add4
 line3`

	got := ParseHunks(patch)
	want := []ChangedLineRange{
		{Start: 2, End: 3},
		{Start: 5, End: 6},
	}
	assertRanges(t, got, want)
}

func assertRanges(t *testing.T, got, want []ChangedLineRange) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d ranges %v, want %d ranges %v", len(got), got, len(want), want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("range[%d]: got %v, want %v", i, got[i], want[i])
		}
	}
}
