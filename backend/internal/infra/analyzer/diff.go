package analyzer

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// diffCollector は CollectDiffFiles が依存する PRRepository の最小インターフェース。
type diffCollector interface {
	GetFileContent(ctx context.Context, token, owner, repo, path, ref string) ([]byte, error)
}

// CollectDiffFiles は変更ファイル一覧に対して GitHub Contents API で before/after 全文を取得する。
// 並列度は 8 に制限する。
func CollectDiffFiles(ctx context.Context, repo diffCollector, token string, pr domain.PRInfo, changed []domain.ChangedFile) ([]domain.DiffFile, error) {
	type result struct {
		idx  int
		file domain.DiffFile
		err  error
	}

	results := make([]result, len(changed))
	sem := make(chan struct{}, 8)

	eg, egCtx := errgroup.WithContext(ctx)
	for i, cf := range changed {
		i, cf := i, cf
		eg.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			df := domain.DiffFile{
				Filename:     cf.Filename,
				PreviousName: cf.PreviousFilename,
				Status:       cf.Status,
				Additions:    cf.Additions,
				Deletions:    cf.Deletions,
			}

			// before コンテンツを取得（added は不要）
			if cf.Status != "added" {
				basePath := cf.Filename
				if cf.PreviousFilename != "" {
					basePath = cf.PreviousFilename
				}
				before, err := repo.GetFileContent(egCtx, token, pr.Owner, pr.Repo, basePath, pr.BaseSHA)
				if err != nil {
					return fmt.Errorf("get before content %s: %w", basePath, err)
				}
				if before != nil {
					df.BeforeContent = string(before)
				}
			}

			// after コンテンツを取得（removed は不要）
			if cf.Status != "removed" {
				after, err := repo.GetFileContent(egCtx, token, pr.Owner, pr.Repo, cf.Filename, pr.HeadSHA)
				if err != nil {
					return fmt.Errorf("get after content %s: %w", cf.Filename, err)
				}
				if after != nil {
					df.AfterContent = string(after)
				}
			}

			results[i] = result{idx: i, file: df}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	out := make([]domain.DiffFile, len(changed))
	for _, r := range results {
		out[r.idx] = r.file
	}
	return out, nil
}
