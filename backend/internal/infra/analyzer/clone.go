package analyzer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

// ErrCloneFailed はgit cloneが失敗したときのsentinel error。
var ErrCloneFailed = errors.New("analyzer: git clone failed")

// CloneRequest はcloneに必要なパラメータ。
type CloneRequest struct {
	Owner, Repo, SHA string
	Token            string // 空文字 = anonymous（公開リポジトリ）
}

// ClonedRepo はclone済みのローカルリポジトリを表す。
type ClonedRepo struct {
	RootDir string       // EvalSymlinks適用済みの絶対パス
	Cleanup func() error // 解析完了後にdeferで呼ぶ
}

// Cloner はリモートGitHubリポジトリを特定SHAで浅くcloneする抽象。
type Cloner interface {
	Clone(ctx context.Context, req CloneRequest) (*ClonedRepo, error)
}

// GitCloner はexec.Command("git", ...)でCloneを実装する本番実装。
type GitCloner struct {
	// GitBin はgitバイナリのパス。空文字のとき "git" を使う。
	GitBin string
	// TmpDir はclone先の親ディレクトリ。空文字のとき os.TempDir() を使う。
	TmpDir string
}

// NewGitCloner は本番用のGitClonerを返す。
func NewGitCloner() *GitCloner {
	return &GitCloner{}
}

// Clone は指定SHAを depth=1 でローカルにcloneする。
// tokenが空のとき匿名cloneになる（公開リポジトリ専用）。
// tokenはURL埋め込みを避けてhttp.extraheader経由で渡す（gitの設定ファイルに残らない）。
func (c *GitCloner) Clone(ctx context.Context, req CloneRequest) (*ClonedRepo, error) {
	gitBin := c.GitBin
	if gitBin == "" {
		gitBin = "git"
	}
	tmpDir := c.TmpDir
	if tmpDir == "" {
		tmpDir = os.TempDir()
	}

	dir, err := os.MkdirTemp(tmpDir, "reviewarena-clone-*")
	if err != nil {
		return nil, fmt.Errorf("analyzer: create tempdir: %w", err)
	}

	cleanup := func() error {
		err := os.RemoveAll(dir)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}

	if err := c.runGit(ctx, gitBin, dir, "init"); err != nil {
		_ = cleanup()
		return nil, err
	}

	remoteURL := fmt.Sprintf("https://github.com/%s/%s.git", req.Owner, req.Repo)
	if err := c.runGit(ctx, gitBin, dir, "remote", "add", "origin", remoteURL); err != nil {
		_ = cleanup()
		return nil, err
	}

	// tokenはhttp.extraheaderで渡す（コマンド引数経由なのでgit configに永続しない）
	fetchArgs := []string{"fetch", "--depth=1", "--no-tags", "--filter=blob:none"}
	if req.Token != "" {
		fetchArgs = append([]string{"-c", "http.extraheader=Authorization: bearer " + req.Token, "-c", "credential.helper="}, fetchArgs...)
	}
	fetchArgs = append(fetchArgs, "origin", req.SHA)
	if err := c.runGit(ctx, gitBin, dir, fetchArgs...); err != nil {
		_ = cleanup()
		return nil, err
	}

	if err := c.runGit(ctx, gitBin, dir, "checkout", "FETCH_HEAD"); err != nil {
		_ = cleanup()
		return nil, err
	}

	// EvalSymlinksでシンボリックリンクを解決しておく（macOS /tmp が /private/tmp になるケースなど）
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		_ = cleanup()
		return nil, fmt.Errorf("analyzer: eval symlinks %s: %w", dir, err)
	}

	return &ClonedRepo{
		RootDir: resolved,
		Cleanup: cleanup,
	}, nil
}

func (c *GitCloner) runGit(ctx context.Context, gitBin, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, gitBin, args...)
	cmd.Dir = dir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: git %v: %s", ErrCloneFailed, args, stderr.String())
	}
	return nil
}
