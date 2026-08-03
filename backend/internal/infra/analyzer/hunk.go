package analyzer

import (
	"strconv"
	"strings"
)

// ChangedLineRange は HEAD 側の変更行範囲（inclusive）を表す。
type ChangedLineRange struct {
	Start int
	End   int
}

// ParseHunks は unified diff テキストから HEAD 側の変更行（+行）の範囲を抽出する。
// patch が空なら nil を返す。
func ParseHunks(patch string) []ChangedLineRange {
	if patch == "" {
		return nil
	}

	lines := strings.Split(patch, "\n")
	var ranges []ChangedLineRange
	rangeStart := -1
	rangeEnd := -1

	newLine := 0 // HEAD 側の現在行番号

	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			// flush
			if rangeStart > 0 {
				ranges = append(ranges, ChangedLineRange{Start: rangeStart, End: rangeEnd})
				rangeStart = -1
			}
			newLine = parseNewStart(line)
			continue
		}

		if newLine == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(line, "+"):
			if rangeStart < 0 {
				rangeStart = newLine
			}
			rangeEnd = newLine
			newLine++
		case strings.HasPrefix(line, "-"):
			// flush
			if rangeStart > 0 {
				ranges = append(ranges, ChangedLineRange{Start: rangeStart, End: rangeEnd})
				rangeStart = -1
			}
			// 削除行は HEAD 側の行番号を進めない
		default:
			// flush
			if rangeStart > 0 {
				ranges = append(ranges, ChangedLineRange{Start: rangeStart, End: rangeEnd})
				rangeStart = -1
			}
			newLine++
		}
	}

	if rangeStart > 0 {
		ranges = append(ranges, ChangedLineRange{Start: rangeStart, End: rangeEnd})
	}
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
