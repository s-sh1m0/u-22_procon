package analyzer

import (
	"strconv"
	"strings"
)

// ChangedLineRange は HEAD 側の変更行範囲（inclusive）を表す。
//
// End = Start-1 の空範囲は「HEAD の End 行目と Start 行目の間で行が削除された」ことを表す。
// 空範囲でも FunctionChanged の重なり判定式がそのまま使え、削除位置が関数の内側
// （宣言行〜閉じ括弧行の間）にあるときだけ重なりと判定される。
type ChangedLineRange struct {
	Start int
	End   int
}

// ParseHunks は unified diff テキストから HEAD 側の変更位置を抽出する。
//   - +行の連続 → その行範囲
//   - 直後に +行が続かない -行の連続（削除のみ） → 削除位置を表す空範囲（ChangedLineRange 参照）
//
// patch が空なら nil を返す（呼び出し側は「変更位置不明」として扱う）。
// patch が空でなければ、変更位置が無くても非 nil の空スライスを返す。
func ParseHunks(patch string) []ChangedLineRange {
	if patch == "" {
		return nil
	}

	lines := strings.Split(patch, "\n")
	ranges := []ChangedLineRange{}
	rangeStart := -1
	rangeEnd := -1
	deleted := false // 直前が -行の連続で、まだ +行に置き換わっていない

	newLine := 0 // HEAD 側の現在行番号

	flush := func() {
		if rangeStart > 0 {
			ranges = append(ranges, ChangedLineRange{Start: rangeStart, End: rangeEnd})
			rangeStart = -1
		}
		if deleted {
			ranges = append(ranges, ChangedLineRange{Start: newLine, End: newLine - 1})
			deleted = false
		}
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			flush()
			newLine = parseNewStart(line)
			continue
		}

		if newLine == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(line, `\`):
			// "\ No newline at end of file" は直前行への注記で、行番号を進めない
		case strings.HasPrefix(line, "+"):
			deleted = false // 削除行は +行で置き換えられた（変更位置は +行の範囲で表す）
			if rangeStart < 0 {
				rangeStart = newLine
			}
			rangeEnd = newLine
			newLine++
		case strings.HasPrefix(line, "-"):
			if rangeStart > 0 {
				ranges = append(ranges, ChangedLineRange{Start: rangeStart, End: rangeEnd})
				rangeStart = -1
			}
			deleted = true
			// 削除行は HEAD 側の行番号を進めない
		default:
			flush()
			newLine++
		}
	}

	flush()
	return ranges
}

// parseNewStart は @@ ヘッダから HEAD 側の開始行番号を取得する。
// 形式: @@ -oldStart[,oldCount] +newStart[,newCount] @@
func parseNewStart(header string) int {
	idx := strings.Index(header, "+")
	if idx < 0 {
		return 0
	}
	rest := header[idx+1:]
	end := strings.IndexAny(rest, ", @")
	if end < 0 {
		return 0
	}
	n, err := strconv.Atoi(rest[:end])
	if err != nil {
		return 0
	}
	return n
}
