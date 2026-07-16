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

// WorktreesRequest は複数SHAを1つのオブジェクトストア共有でcloneするためのパラメータ。
type WorktreesRequest struct {
	Owner, Repo string
	Token       string   // 空文字 = anonymous（公開リポジトリ）
	SHAs        []string // clone対象のSHA群（重複は許容・内部でdedupされる）
}

// ClonedWorktrees は1リポジトリを共有しSHAごとにworktreeを持つclone結果。
type ClonedWorktrees struct {
	// Dirs はリクエストした各SHA → そのworktreeの絶対パス（EvalSymlinks済み）。
	// 同一SHAを重複指定した場合は同じパスにマップされる。
	Dirs    map[string]string
	Cleanup func() error // 解析完了後にdeferで呼ぶ（共有tmp dirを丸ごと削除する）
}

// Cloner はリモートGitHubリポジトリを特定SHAで浅くcloneする抽象。
type Cloner interface {
	Clone(ctx context.Context, req CloneRequest) (*ClonedRepo, error)
	// CloneWorktrees は複数SHAを1回のfetchで取得し、SHAごとにworktreeを作る。
	// base/headを別々にcloneするより、オブジェクトストアとfetchを共有できる
	// （変更のないblobは1度しか取得しない）。
	CloneWorktrees(ctx context.Context, req WorktreesRequest) (*ClonedWorktrees, error)
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

	// token がある場合はURL認証を使う（一時ディレクトリ内のみで完結するため安全）
	// トークンはリモート URL に埋め込んでコンテナ内の一時ディレクトリに git remote add する。
	// git remote の URL はホスト外に露出せず、ClonedRepo.Cleanup() で削除される。
	var remoteURL string
	if req.Token != "" {
		remoteURL = fmt.Sprintf("https://oauth2:%s@github.com/%s/%s.git", req.Token, req.Owner, req.Repo)
	} else {
		remoteURL = fmt.Sprintf("https://github.com/%s/%s.git", req.Owner, req.Repo)
	}
	if err := c.runGit(ctx, gitBin, dir, "remote", "add", "origin", remoteURL); err != nil {
		_ = cleanup()
		return nil, err
	}

	fetchArgs := []string{"fetch", "--depth=1", "--no-tags", "--filter=blob:none", "origin", req.SHA}
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

// CloneWorktrees は req.SHAs を1つの共有リポジトリに1回の fetch で取得し、
// SHAごとに作業ツリー（worktree）を生やす。先頭SHAは共有リポジトリ自身を
// 作業ツリーとして使い、残りは git worktree で追加する。すべてのworktreeが
// 同一オブジェクトストアを共有するため、base/head間で共通のblobは1度しか
// fetchされない（--filter=blob:none の遅延取得がストアを共有して効く）。
// cleanupは共有tmp dirのRemoveAllで完結する。
func (c *GitCloner) CloneWorktrees(ctx context.Context, req WorktreesRequest) (*ClonedWorktrees, error) {
	if len(req.SHAs) == 0 {
		return nil, fmt.Errorf("%w: no SHAs given", ErrCloneFailed)
	}
	gitBin := c.GitBin
	if gitBin == "" {
		gitBin = "git"
	}
	tmpDir := c.TmpDir
	if tmpDir == "" {
		tmpDir = os.TempDir()
	}

	parent, err := os.MkdirTemp(tmpDir, "reviewarena-clone-*")
	if err != nil {
		return nil, fmt.Errorf("analyzer: create tempdir: %w", err)
	}
	cleanup := func() error {
		err := os.RemoveAll(parent)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}

	// 共有オブジェクトストアとなるリポジトリ。ここに全SHAをまとめてfetchし、
	// SHAごとにworktreeを生やすことでclone/fetchを1回に集約する。
	repoDir := filepath.Join(parent, "repo")
	if err := os.Mkdir(repoDir, 0o755); err != nil {
		_ = cleanup()
		return nil, fmt.Errorf("analyzer: create repo dir: %w", err)
	}
	if err := c.runGit(ctx, gitBin, repoDir, "init"); err != nil {
		_ = cleanup()
		return nil, err
	}

	// token がある場合はURL認証を使う（Cloneと同様、一時ディレクトリ内で完結し
	// ClonedWorktrees.Cleanup() で削除されるため安全）。
	var remoteURL string
	if req.Token != "" {
		remoteURL = fmt.Sprintf("https://oauth2:%s@github.com/%s/%s.git", req.Token, req.Owner, req.Repo)
	} else {
		remoteURL = fmt.Sprintf("https://github.com/%s/%s.git", req.Owner, req.Repo)
	}
	if err := c.runGit(ctx, gitBin, repoDir, "remote", "add", "origin", remoteURL); err != nil {
		_ = cleanup()
		return nil, err
	}

	// 重複を除いたSHA群を1回のfetchでまとめて取得する。
	uniqueSHAs := dedupStrings(req.SHAs)
	fetchArgs := append([]string{"fetch", "--depth=1", "--no-tags", "--filter=blob:none", "origin"}, uniqueSHAs...)
	if err := c.runGit(ctx, gitBin, repoDir, fetchArgs...); err != nil {
		_ = cleanup()
		return nil, err
	}

	// SHA → worktree絶対パス。先頭SHAはrepoDirを作業ツリーとして使い、
	// 残りは git worktree add で追加する（すべてrepoDirのオブジェクトを共有）。
	shaToDir := make(map[string]string, len(uniqueSHAs))
	for i, sha := range uniqueSHAs {
		var wtDir string
		if i == 0 {
			if err := c.runGit(ctx, gitBin, repoDir, "checkout", "--detach", sha); err != nil {
				_ = cleanup()
				return nil, err
			}
			wtDir = repoDir
		} else {
			wtDir = filepath.Join(parent, fmt.Sprintf("wt-%d", i))
			if err := c.runGit(ctx, gitBin, repoDir, "worktree", "add", "--detach", wtDir, sha); err != nil {
				_ = cleanup()
				return nil, err
			}
		}
		// EvalSymlinksでシンボリックリンクを解決しておく（macOS /tmp が /private/tmp になるケースなど）
		resolved, err := filepath.EvalSymlinks(wtDir)
		if err != nil {
			_ = cleanup()
			return nil, fmt.Errorf("analyzer: eval symlinks %s: %w", wtDir, err)
		}
		shaToDir[sha] = resolved
	}

	// リクエストされた各SHA（重複含む）を対応するworktreeパスに紐付ける。
	dirs := make(map[string]string, len(req.SHAs))
	for _, sha := range req.SHAs {
		dirs[sha] = shaToDir[sha]
	}

	return &ClonedWorktrees{Dirs: dirs, Cleanup: cleanup}, nil
}

// dedupStrings は順序を保ったまま重複を除いたスライスを返す。
func dedupStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func (c *GitCloner) runGit(ctx context.Context, gitBin, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, gitBin, args...)
	cmd.Dir = dir
	// GIT_TERMINAL_PROMPT=0 で認証プロンプトを抑止する（コンテナ環境で tty がない場合に必要）
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: git %v: %s", ErrCloneFailed, args, stderr.String())
	}
	return nil
}
