package analyzer

import (
	"runtime"
	"strconv"
)

// maxBuildParallelism は依存パッケージを同時にコンパイルする数の上限。
// go の既定は CPU 数だが、compile は 1 プロセスで数百 MB 使う（クラウド SDK では 500MB 超）。
// CPU 数だけ並べるとコンテナのメモリ上限に張り付いてスラッシングするため絞る。
const maxBuildParallelism = 4

// buildParallelism は go コマンドの -p に渡す値を返す。
func buildParallelism() int {
	return min(runtime.NumCPU(), maxBuildParallelism)
}

// goflagsWithParallelism は既存の GOFLAGS の値に -p を足した値を返す。
func goflagsWithParallelism(existing string) string {
	p := "-p=" + strconv.Itoa(buildParallelism())
	if existing == "" {
		return p
	}
	return existing + " " + p
}
