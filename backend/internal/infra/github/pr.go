package github

import (
	"context"
	"fmt"
	"strings"

	gogithub "github.com/google/go-github/v69/github"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// PRRepo は domain.PRRepository を go-github で実装する。
// newClient フィールドはテスト時に httptest 用クライアントへ差し替えられる。
type PRRepo struct {
	newClient func(ctx context.Context, token string) *gogithub.Client
}

// NewPRRepo は本番用の PRRepo を返す。
func NewPRRepo() *PRRepo {
	return &PRRepo{newClient: NewClient}
}

// GetPR は PR のメタデータ（タイトル・base/head ref・head SHA）を取得する。
func (r *PRRepo) GetPR(ctx context.Context, token, owner, repo string, number int) (*domain.PRInfo, error) {
	cli := r.newClient(ctx, token)
	pr, _, err := cli.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("github: get PR %s/%s#%d: %w", owner, repo, number, err)
	}
	return &domain.PRInfo{
		Owner:   owner,
		Repo:    repo,
		Number:  number,
		Title:   pr.GetTitle(),
		BaseRef: pr.GetBase().GetRef(),
		HeadRef: pr.GetHead().GetRef(),
		HeadSHA: pr.GetHead().GetSHA(),
	}, nil
}

// ListChangedGoFiles は PR の変更ファイルから .go のみを返す。ページネーション対応。
func (r *PRRepo) ListChangedGoFiles(ctx context.Context, token, owner, repo string, number int) ([]domain.ChangedFile, error) {
	cli := r.newClient(ctx, token)
	var out []domain.ChangedFile
	opts := &gogithub.ListOptions{PerPage: 100}
	for {
		files, resp, err := cli.PullRequests.ListFiles(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("github: list PR files %s/%s#%d: %w", owner, repo, number, err)
		}
		for _, f := range files {
			name := f.GetFilename()
			if !strings.HasSuffix(name, ".go") {
				continue
			}
			out = append(out, domain.ChangedFile{
				Filename:  name,
				Status:    f.GetStatus(),
				Additions: f.GetAdditions(),
				Deletions: f.GetDeletions(),
				Patch:     f.GetPatch(),
			})
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return out, nil
}
