// bench は解析パイプラインの各フェーズの所要時間を計測するCLIツール。
// 使い方: bench -token <GitHub PAT> -pr <PR URL>
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

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
	flag.Parse()

	if *token == "" {
		log.Fatal("GitHub token required: -token <token> or GITHUB_TOKEN env")
	}

	ctx := context.Background()

	log.Printf("=== bench start: %s/%s#%d ===", *owner, *repo, *prNum)
	tTotal := time.Now()

	// 1. PR メタ情報取得
	t0 := time.Now()
	prRepo := ghinfra.NewPRRepo()
	prInfo, err := prRepo.GetPR(ctx, *token, *owner, *repo, *prNum)
	if err != nil {
		log.Fatalf("GetPR: %v", err)
	}
	log.Printf("[1] GetPR took %s (head=%s)", time.Since(t0), prInfo.HeadSHA[:8])

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

	// 3. clone + FastLoad + 近傍Load
	t2 := time.Now()
	sourceTree := analyzer.NewSourceTree(analyzer.NewGitCloner(), analyzer.NewGoPackageLoader())
	prepared, err := sourceTree.Prepare(ctx, *prInfo, *token, changed)
	if err != nil {
		log.Fatalf("Prepare: %v", err)
	}
	defer func() { _ = prepared.Cleanup() }()
	log.Printf("[3] Prepare (clone+load) took %s", time.Since(t2))
	log.Printf("    packages loaded: %d, changed pkgs: %v", len(prepared.Packages), prepared.ChangedPackages)

	// 4. コールグラフ構築
	t3 := time.Now()
	cgBuilder := analyzer.NewGoCallGraphBuilder()
	graph, err := cgBuilder.Build(ctx, prepared.Packages, prepared.ChangedPackages, prepared.ChangedFileAbsPaths, prepared.RootDir)
	if err != nil {
		log.Fatalf("Build: %v", err)
	}
	log.Printf("[4] Build callgraph took %s (nodes=%d edges=%d)", time.Since(t3), len(graph.Nodes), len(graph.Edges))

	// 5. クラスタリング
	t4 := time.Now()
	clusterer := cluster.NewLouvainClusterer()
	_, err = clusterer.Cluster(ctx, graph)
	if err != nil {
		log.Fatalf("Cluster: %v", err)
	}
	log.Printf("[5] Cluster took %s", time.Since(t4))

	log.Printf("=== total: %s ===", time.Since(tTotal))

	// 6. グラフの中身確認（外部パッケージが混じっていないかチェック）
	fmt.Println("\n--- nodes (package) ---")
	pkgCount := make(map[string]int)
	for _, n := range graph.Nodes {
		pkgCount[n.Package]++
	}
	for pkg, cnt := range pkgCount {
		fmt.Printf("  %s (%d nodes)\n", pkg, cnt)
	}

	_ = domain.PRInfo{}
}
