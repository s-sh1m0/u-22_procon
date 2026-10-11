package analyzer

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

// maxBuildParallelism は依存パッケージを同時にコンパイルする数の上限。
// go の既定は CPU 数だが、compile は 1 プロセスで数百 MB 使う（クラウド SDK では 500MB 超）。
// CPU 数だけ並べるとコンテナのメモリ上限に張り付いてスラッシングするため絞る。
const maxBuildParallelism = 4

// buildParallelism は go コマンドの -p に渡す値を返す。
func buildParallelism() int {
	return min(runtime.NumCPU(), maxBuildParallelism)
}

// warmExportData は patterns とその依存パッケージを go list -export でコンパイルし、
// export data をビルドキャッシュに作っておく。
//
// go/packages は依存の型情報を export data から読むため（loadMode を参照）、内部で
// go list -export を実行し、キャッシュに無い依存をコンパイルする。これを go/packages に
// 任せると次の問題が起きる。
//   - base/head や同時に走る別ジョブが、同じ依存を同時にコンパイルする
//   - タイムアウトで go を kill しても子の compile が残り、数分間 CPU とメモリを使い続ける
//
// そこで先にこの関数でコンパイルを済ませ、go/packages 側はキャッシュに当たるだけにする。
// buildSem で同時に 1 つだけ実行し、後から実行した側は先の結果をキャッシュから使う。
// キャンセル時は compile を含むプロセスグループごと kill する。
// フラグは go/packages の go list 呼び出しと揃える（-pgo などが違うとキャッシュに当たらない）。
func (l *GoPackageLoader) warmExportData(ctx context.Context, dir string, patterns []string) error {
	tWait := time.Now()
	if l.buildSem != nil {
		select {
		case l.buildSem <- struct{}{}:
			defer func() { <-l.buildSem }()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	waited := time.Since(tWait)

	args := []string{
		"list", "-e", "-compiled=true", "-test=false", "-export=true", "-deps=true",
		"-find=false", "-buildvcs=false", "-pgo=off", "-f={{.ImportPath}}", "--",
	}
	cmd := exec.CommandContext(ctx, "go", append(args, patterns...)...)
	cmd.Dir = dir
	cmd.Env = analyzerEnv()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.WaitDelay = 5 * time.Second
	killProcessGroupOnCancel(cmd)

	tBuild := time.Now()
	if err := cmd.Run(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("go list -export: %w: %s", err, tailString(stderr.String(), 1024))
	}
	log.Printf("analyzer: warm export data (%d pkgs) took %s (waited %s)", len(patterns), time.Since(tBuild), waited)
	return nil
}

// tailString は s の末尾 n バイトを返す（エラーメッセージに長い stderr を全部載せないため）。
func tailString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "..." + s[len(s)-n:]
}

// goflagsWithParallelism は既存の GOFLAGS の値に -p を足した値を返す。
func goflagsWithParallelism(existing string) string {
	p := "-p=" + strconv.Itoa(buildParallelism())
	if existing == "" {
		return p
	}
	return existing + " " + p
}
