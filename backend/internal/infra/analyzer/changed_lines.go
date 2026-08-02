package analyzer

import (
	"os"
	"path/filepath"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// ChangedLines はファイル絶対パスから変更行範囲へのマップ。
// マップキーに存在するがスライスが nil のファイルは「変更ありだが行範囲不明」を意味し、
// ファイル単位にフォールバックする。
type ChangedLines map[string][]ChangedLineRange

// BuildChangedLines は ChangedFile リストから ChangedLines を構築する。
// repoRoot は GitHub API のファイルパスを絶対パスに変換するための基点。
// ディスク上に存在しないファイル（HEAD 側での removed 等）はスキップする。
func BuildChangedLines(changed []domain.ChangedFile, repoRoot string) ChangedLines {
	cl := make(ChangedLines, len(changed))
	for _, f := range changed {
		absPath := filepath.Clean(filepath.Join(repoRoot, f.Filename))
		if _, err := os.Stat(absPath); err != nil {
			continue
		}
		cl[absPath] = ParseHunks(f.Patch)
	}
	return cl
}

// FunctionChanged は指定ファイルの [startLine, endLine] 範囲に変更行が含まれるかを返す。
//   - ファイルが ChangedLines に存在しない → false（変更なし）
//   - ファイルが存在するが行範囲が nil → true（Patch が空でファイル単位フォールバック）
//   - 行範囲が存在する → いずれかの ChangedLineRange と重なれば true
func (cl ChangedLines) FunctionChanged(absFilePath string, startLine, endLine int) bool {
	ranges, ok := cl[absFilePath]
	if !ok {
		return false
	}
	if ranges == nil {
		return true
	}
	for _, r := range ranges {
		if r.Start <= endLine && r.End >= startLine {
			return true
		}
	}
	return false
}

// FileChanged はファイルが ChangedLines に含まれるかを返す。
// callgraph.go で Syntax() が nil の合成関数に対するフォールバックとして使う。
func (cl ChangedLines) FileChanged(absFilePath string) bool {
	_, ok := cl[absFilePath]
	return ok
}
