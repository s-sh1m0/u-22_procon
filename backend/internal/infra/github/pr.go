package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	gogithub "github.com/google/go-github/v69/github"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// maxFileSize は GitHub Contents API が返す最大サイズ（1MB）。超過時は空を返す。
const maxFileSize = 1 * 1024 * 1024

// PRRepo は domain.PRRepository を go-github で実装する。
// newClient フィールドはテスト時に httptest 用クライアントへ差し替えられる。
type PRRepo struct {
	newClient func(ctx context.Context, token string) *gogithub.Client
}

// NewPRRepo は本番用の PRRepo を返す。
// レート制限対応・クライアントキャッシュ付きの ClientFactory を使う。
func NewPRRepo() *PRRepo {
	return &PRRepo{newClient: NewClientFactory()}
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
		BaseSHA: pr.GetBase().GetSHA(),
	}, nil
}

// GetFileContent は指定 ref のファイル全文を返す。ファイル不在は (nil, nil)、1MB 超は (nil, nil)。
func (r *PRRepo) GetFileContent(ctx context.Context, token, owner, repo, path, ref string) ([]byte, error) {
	cli := r.newClient(ctx, token)
	fc, _, resp, err := cli.Repositories.GetContents(ctx, owner, repo, path, &gogithub.RepositoryContentGetOptions{Ref: ref})
	if err != nil {
		var ghErr *gogithub.ErrorResponse
		if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("github: get file content %s@%s: %w", path, ref, err)
	}
	_ = resp
	if fc == nil {
		return nil, nil
	}
	if fc.GetSize() > maxFileSize {
		return nil, nil
	}
	content, err := fc.GetContent()
	if err != nil {
		return nil, fmt.Errorf("github: decode file content %s@%s: %w", path, ref, err)
	}
	return []byte(content), nil
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
				Filename:         name,
				PreviousFilename: f.GetPreviousFilename(),
				Status:           f.GetStatus(),
				Additions:        f.GetAdditions(),
				Deletions:        f.GetDeletions(),
				Patch:            f.GetPatch(),
			})
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return out, nil
}
