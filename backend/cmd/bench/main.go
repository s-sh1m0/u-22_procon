// bench は解析パイプラインの各フェーズの所要時間を計測するCLIツール。
// 使い方:
//
//	bench -token <GitHub PAT> -owner <owner> -repo <repo> -pr <PR番号>        # head 側のみ（従来）
//	bench -token <GitHub PAT> -owner <owner> -repo <repo> -pr <PR番号> -both  # worker と同じ base+head 並列 → Merge → Cluster
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
	"github.com/s-sh1m0/u-22_procon/backend/internal/infra/analyzer"
	"github.com/s-sh1m0/u-22_procon/backend/internal/infra/cluster"
	ghinfra "github.com/s-sh1m0/u-22_procon/backend/internal/infra/github"
)

func main() {
	token := flag.String("token", os.Getenv("GITHUB_TOKEN"), "GitHub personal access token")
	owner := flag.String("owner", "s-sh1m0", "repo owner")
	repo := flag.String("repo", "u-22_procon", "repo name")
	prNum := flag.Int("pr", 21, "PR number")
	both := flag.Bool("both", false, "base+head を並列構築し Merge → Cluster まで実行する（worker と同じパイプライン）")
	flag.Parse()

	if *token == "" {
		log.Fatal("GitHub token required: -token <token> or GITHUB_TOKEN env")
	}

	ctx := context.Background()

	mode := "head-only"
	if *both {
		mode = "both"
	}
	log.Printf("=== bench start: %s/%s#%d (%s) ===", *owner, *repo, *prNum, mode)
	tTotal := time.Now()

	// 1. PR メタ情報取得
	t0 := time.Now()
	prRepo := ghinfra.NewPRRepo()
	prInfo, err := prRepo.GetPR(ctx, *token, *owner, *repo, *prNum)
	if err != nil {
		log.Fatalf("GetPR: %v", err)
	}
	log.Printf("[1] GetPR took %s (head=%s base=%s)", time.Since(t0), shortSHA(prInfo.HeadSHA), shortSHA(prInfo.BaseSHA))

	// 2. 変更ファイル取得
	t1 := time.Now()
	changed, err := prRepo.ListChangedGoFiles(ctx, *token, *owner, *repo, *prNum)
	if err != nil {
		log.Fatalf("ListChangedGoFiles: %v", err)
	}
	log.Printf("[2] ListChangedGoFiles took %s (%d files)", time.Since(t1), len(changed))
	for _, f := range changed {
		fmt.Printf("    %s (%s)\n", f.Filename, f.Status)
	}

	// runBoth/runHeadOnly は clone した一時ディレクトリを defer で削除する。
	// ここで直接 log.Fatalf すると os.Exit で defer が飛び一時ディレクトリが
	// 残留するため、error を受け取ってから終了する（bench は A/B 計測で
	// 繰り返し回すのでリークが溜まりやすい）。
	if *both {
		err = runBoth(ctx, *prInfo, *token, changed)
	} else {
		err = runHeadOnly(ctx, *prInfo, *token, changed)
	}
	if err != nil {
		log.Fatalf("bench failed: %v", err)
	}

	log.Printf("=== total: %s ===", time.Since(tTotal))
}

// runHeadOnly は head 側のみを clone → Build → Cluster する（従来の bench 挙動）。
// エラー時も defer した Cleanup が走るよう、log.Fatalf せず error を返す。
func runHeadOnly(ctx context.Context, prInfo domain.PRInfo, token string, changed []domain.ChangedFile) error {
	// 3. clone + FastLoad + 近傍Load
	t0 := time.Now()
	sourceTree := analyzer.NewSourceTree(analyzer.NewGitCloner(), analyzer.NewGoPackageLoader())
	prepared, err := sourceTree.Prepare(ctx, prInfo, token, changed)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer func() { _ = prepared.Cleanup() }()
	log.Printf("[3] Prepare (clone+load) took %s", time.Since(t0))
	log.Printf("    packages loaded: %d, changed pkgs: %v", len(prepared.Packages), prepared.ChangedPackages)

	// 4. コールグラフ構築
	t1 := time.Now()
	cgBuilder := analyzer.NewGoCallGraphBuilder()
	graph, err := cgBuilder.Build(ctx, prepared.Packages, prepared.ChangedPackages, prepared.ChangedLines, prepared.RepoRoot)
	if err != nil {
		return fmt.Errorf("build: %w", err)
	}
	log.Printf("[4] Build callgraph took %s (nodes=%d edges=%d)", time.Since(t1), len(graph.Nodes), len(graph.Edges))

	// 5. クラスタリング
	t2 := time.Now()
	clusterer := cluster.NewLouvainClusterer()
	if _, err := clusterer.Cluster(ctx, graph); err != nil {
		return fmt.Errorf("cluster: %w", err)
	}
	log.Printf("[5] Cluster took %s", time.Since(t2))

	// 6. グラフの中身確認（外部パッケージが混じっていないかチェック）
	fmt.Println("\n--- nodes (package) ---")
	pkgCount := make(map[string]int)
	for _, n := range graph.Nodes {
		pkgCount[n.Package]++
	}
	for pkg, cnt := range pkgCount {
		fmt.Printf("  %s (%d nodes)\n", pkg, cnt)
	}
	return nil
}

// runBoth は worker と同じパイプライン（base+head 並列構築 → Merge → cycles/violations
// 検出 → Cluster）を実行し、ノード/エッジ数のサマリを出力する。
// PR-F（依存の export data 化 / #82）の A/B 比較ゲートとして、新旧実装で
// このサマリのノード/エッジ数が完全一致することの確認に使う。
// エラー時も defer した Cleanup（base/head 共有の一時ディレクトリ）が走るよう、
// log.Fatalf せず error を返す。
func runBoth(ctx context.Context, prInfo domain.PRInfo, token string, changed []domain.ChangedFile) error {
	sourceTree := analyzer.NewSourceTree(analyzer.NewGitCloner(), analyzer.NewGoPackageLoader())
	cgBuilder := analyzer.NewGoCallGraphBuilder()

	// 3. base+head を並列構築（worker.buildBothGraphs と同じ構成）。
	// clone は CloneBoth で1リポジトリ共有（1回の fetch）にまとめる。
	t0 := time.Now()
	var headGraph, baseGraph *domain.Graph

	tc := time.Now()
	cw, err := sourceTree.CloneBoth(ctx, prInfo, token)
	if err != nil {
		return fmt.Errorf("clone base/head: %w", err)
	}
	defer func() { _ = cw.Cleanup() }()
	log.Printf("[3a] CloneBoth (shared clone) took %s", time.Since(tc))

	headDir, ok := cw.Dirs[prInfo.HeadSHA]
	if !ok {
		return fmt.Errorf("clone base/head: missing worktree for head %s", prInfo.HeadSHA)
	}
	baseDir, ok := cw.Dirs[prInfo.BaseSHA]
	if !ok {
		return fmt.Errorf("clone base/head: missing worktree for base %s", prInfo.BaseSHA)
	}

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		t := time.Now()
		prepared, err := sourceTree.PrepareFromDir(egCtx, headDir, changed)
		if err != nil {
			return fmt.Errorf("head prepare: %w", err)
		}
		log.Printf("[head] PrepareFromDir took %s (pkgs=%d, changed pkgs=%d)", time.Since(t), len(prepared.Packages), len(prepared.ChangedPackages))
		t = time.Now()
		g, err := cgBuilder.Build(egCtx, prepared.Packages, prepared.ChangedPackages, prepared.ChangedLines, prepared.RepoRoot)
		if err != nil {
			return fmt.Errorf("head build: %w", err)
		}
		log.Printf("[head] Build took %s (nodes=%d edges=%d)", time.Since(t), len(g.Nodes), len(g.Edges))
		headGraph = g
		return nil
	})
	eg.Go(func() error {
		t := time.Now()
		baseChanged := domain.NormalizeChangedForBase(changed)
		prepared, err := sourceTree.PrepareFromDir(egCtx, baseDir, baseChanged)
		if err != nil {
			return fmt.Errorf("base prepare: %w", err)
		}
		log.Printf("[base] PrepareFromDir took %s (pkgs=%d, changed pkgs=%d)", time.Since(t), len(prepared.Packages), len(prepared.ChangedPackages))
		t = time.Now()
		g, err := cgBuilder.Build(egCtx, prepared.Packages, prepared.ChangedPackages, prepared.ChangedLines, prepared.RepoRoot)
		if err != nil {
			return fmt.Errorf("base build: %w", err)
		}
		log.Printf("[base] Build took %s (nodes=%d edges=%d)", time.Since(t), len(g.Nodes), len(g.Edges))
		baseGraph = g
		return nil
	})
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("build base/head graphs: %w", err)
	}
	log.Printf("[3] parallel base+head build took %s", time.Since(t0))

	// 4. Merge + cycles/violations 検出
	t1 := time.Now()
	diffGraph := analyzer.MergeWithDiffStatus(baseGraph, headGraph)
	cycles := analyzer.DetectNewCycles(baseGraph, headGraph)
	violations := analyzer.DetectLayeringViolations(diffGraph)
	log.Printf("[4] Merge+Detect took %s", time.Since(t1))

	// 5. クラスタリング
	t2 := time.Now()
	clusterer := cluster.NewLouvainClusterer()
	result, err := clusterer.Cluster(ctx, diffGraph)
	if err != nil {
		return fmt.Errorf("cluster: %w", err)
	}
	log.Printf("[5] Cluster took %s", time.Since(t2))

	// 6. ノード/エッジ数サマリ（A/B 比較で完全一致を確認する対象）
	nodeCnt := map[domain.DiffStatus]int{}
	for _, n := range diffGraph.Nodes {
		nodeCnt[n.DiffStatus]++
	}
	edgeCnt := map[domain.DiffStatus]int{}
	for _, e := range diffGraph.Edges {
		edgeCnt[e.Status]++
	}
	fmt.Println("\n--- summary ---")
	fmt.Printf("head  : nodes=%d edges=%d\n", len(headGraph.Nodes), len(headGraph.Edges))
	fmt.Printf("base  : nodes=%d edges=%d\n", len(baseGraph.Nodes), len(baseGraph.Edges))
	fmt.Printf("merged: nodes=%d (existing=%d added=%d removed=%d)\n",
		len(diffGraph.Nodes), nodeCnt[domain.DiffStatusExisting], nodeCnt[domain.DiffStatusAdded], nodeCnt[domain.DiffStatusRemoved])
	fmt.Printf("        edges=%d (existing=%d added=%d removed=%d)\n",
		len(diffGraph.Edges), edgeCnt[domain.DiffStatusExisting], edgeCnt[domain.DiffStatusAdded], edgeCnt[domain.DiffStatusRemoved])
	fmt.Printf("result: clusters=%d cycles=%d violations=%d\n", len(result.Clusters), len(cycles), len(violations))
	return nil
}

// shortSHA は SHA の先頭8桁を返す（8桁未満ならそのまま）。
func shortSHA(sha string) string {
	if len(sha) >= 8 {
		return sha[:8]
	}
	return sha
}
