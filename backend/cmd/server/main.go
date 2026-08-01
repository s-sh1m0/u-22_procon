package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/s-sh1m0/u-22_procon/backend/internal/api"
	"github.com/s-sh1m0/u-22_procon/backend/internal/infra/analyzer"
	"github.com/s-sh1m0/u-22_procon/backend/internal/infra/cluster"
	githubinfra "github.com/s-sh1m0/u-22_procon/backend/internal/infra/github"
	"github.com/s-sh1m0/u-22_procon/backend/internal/infra/job"
	"github.com/s-sh1m0/u-22_procon/backend/internal/infra/store"
	"github.com/s-sh1m0/u-22_procon/backend/internal/usecase"
)

func main() {
	clientID := mustEnv("GITHUB_CLIENT_ID")
	clientSecret := mustEnv("GITHUB_CLIENT_SECRET")
	callbackURL := mustEnv("GITHUB_CALLBACK_URL")
	sessionSecret := mustEnv("SESSION_SECRET")
	frontendURL := envOr("FRONTEND_URL", "")
	staticDir := envOr("STATIC_DIR", "")
	port := envOr("PORT", "8080")
	dbPath := envOr("DB_PATH", "/app/data/reviewarena.db")

	db, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	sessions, err := store.NewSessionRepo(db, []byte(sessionSecret))
	if err != nil {
		log.Fatalf("init session repo: %v", err)
	}

	analysisRepo := store.NewAnalysisRepo(db)
	jobRepo := store.NewJobRepo(db)

	// 再起動で処理されなくなった pending/running ジョブをエラーに更新する。
	// トークンは DB に保存していないため再キューイングは不可。ユーザーに再送信を促す。
	if n, err := jobRepo.MarkStaleJobsError(context.Background()); err != nil {
		log.Printf("warn: mark stale jobs: %v", err)
	} else if n > 0 {
		log.Printf("startup: marked %d stale job(s) as error (server restarted)", n)
	}

	prRepo := githubinfra.NewPRRepo()
	sourceTree := analyzer.NewSourceTree(analyzer.NewGitCloner(), analyzer.NewGoPackageLoader())
	cgBuilder := analyzer.NewGoCallGraphBuilder()
	clusterer := cluster.NewLouvainClusterer()

	queueBuf := envIntOr("JOB_QUEUE_BUFFER", 32)
	queue := job.NewQueue(queueBuf)

	// 1ジョブの解析時間上限。暴走解析がワーカーを占有し続けるのを防ぐ安全弁。
	jobTimeout := time.Duration(envIntOr("JOB_TIMEOUT_SECONDS", 180)) * time.Second
	worker := job.NewWorker(queue, jobRepo, analysisRepo, prRepo, sourceTree, cgBuilder, clusterer,
		job.WithJobTimeout(jobTimeout))

	// 同時に走らせる解析ワーカー数。直列実行だと1ジョブが詰まると後続（他ユーザー）が
	// 全員待たされるためプール化する。1ジョブは base/head を内部で並列構築し CPU/メモリを
	// 多く使う（型チェック＋SSA/CHA）ため、既定値は CPU 数と 4 の小さい方に抑える。
	// メモリ枯渇を避けるには compose.yml の GOMEMLIMIT/mem_limit と合わせて調整すること。
	workerCount := envIntOr("JOB_WORKERS", min(runtime.NumCPU(), 4))
	if workerCount < 1 {
		workerCount = 1
	}

	analyzeUC := usecase.NewAnalyzePRUseCase(analysisRepo, jobRepo, prRepo, queue)
	getUC := usecase.NewGetAnalysisUseCase(analysisRepo, jobRepo)

	parsedCallback, err := url.Parse(callbackURL)
	if err != nil {
		log.Fatalf("invalid GITHUB_CALLBACK_URL: %v", err)
	}
	secureCookies := parsedCallback.Scheme == "https"

	oauth := githubinfra.NewOAuthConfig(clientID, clientSecret, callbackURL)
	authHandler := api.NewAuthHandler(oauth, sessions, frontendURL, secureCookies)
	analysisHandler := api.NewAnalysisHandler(analyzeUC, getUC)
	jobHandler := api.NewJobHandler(getUC)
	diffHandler := api.NewDiffHandler(getUC)
	router := api.NewRouter(authHandler, analysisHandler, jobHandler, diffHandler, sessions, frontendURL, staticDir)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	eg, egCtx := errgroup.WithContext(ctx)

	log.Printf("startup: launching %d analysis worker(s) (job timeout %s)", workerCount, jobTimeout)
	for i := 0; i < workerCount; i++ {
		eg.Go(func() error {
			return worker.Run(egCtx)
		})
	}

	eg.Go(func() error {
		if err := router.Start(":" + port); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		<-egCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := router.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
		queue.Close()
		return nil
	})

	if err := eg.Wait(); err != nil && err != context.Canceled {
		log.Fatalf("server error: %v", err)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env %s is not set", key)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
