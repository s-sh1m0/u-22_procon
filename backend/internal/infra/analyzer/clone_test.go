package analyzer

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupLocalBareRepo はt.TempDir()内にgit bareリポジトリを作り、
// コミット1件を含む別の作業ディレクトリからpushしたSHAを返す。
// file:// URLでGitClonerのテストに使う。
func setupLocalBareRepo(t *testing.T) (bareURL string, commitSHA string) {
	t.Helper()

	bareDir := t.TempDir()
	if err := exec.Command("git", "init", "--bare", bareDir).Run(); err != nil {
		t.Fatalf("git init --bare: %v", err)
	}

	workDir := t.TempDir()
	runIn := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
		}
	}

	runIn(workDir, "init")
	runIn(workDir, "config", "user.email", "test@example.com")
	runIn(workDir, "config", "user.name", "Test")

	// go.mod とシンプルなGoファイルを用意する
	if err := os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module example.com/test\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	runIn(workDir, "add", ".")
	runIn(workDir, "commit", "-m", "init")
	runIn(workDir, "remote", "add", "origin", bareDir)
	runIn(workDir, "push", "origin", "HEAD:refs/heads/main")

	// コミットSHAを取得
	out, err := exec.Command("git", "-C", workDir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}
	sha := strings.TrimSpace(string(out))

	return "file://" + bareDir, sha
}

func TestGitCloner_Clone_Local(t *testing.T) {
	bareURL, sha := setupLocalBareRepo(t)

	cloner := &GitCloner{}
	// file:// URLのときにfetchが動くようにremote URLをbareURLに向ける
	// GitClonerはhttps://github.com/... を組み立てるので、テスト用にオーバーライドが必要。
	// そこで fetchArgs の組み立てを直接呼ぶためにrunGitをラップして手動実行する。
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		if err := cloner.runGit(context.Background(), "git", dir, args...); err != nil {
			t.Fatalf("runGit %v: %v", args, err)
		}
	}

	run("init")
	run("remote", "add", "origin", bareURL)
	run("fetch", "--depth=1", "--no-tags", "origin", sha)
	run("checkout", "FETCH_HEAD")

	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}

	// go.mod が存在することを確認
	if _, err := os.Stat(filepath.Join(resolved, "go.mod")); err != nil {
		t.Errorf("go.mod not found in clone: %v", err)
	}
	if _, err := os.Stat(filepath.Join(resolved, "main.go")); err != nil {
		t.Errorf("main.go not found in clone: %v", err)
	}

	// Cleanup で消えることを確認
	cloned := &ClonedRepo{RootDir: resolved, Cleanup: func() error { return os.RemoveAll(dir) }}
	if err := cloned.Cleanup(); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected dir to be removed, got: %v", err)
	}
}

// setupLocalBareRepoTwoCommits は bare リポジトリに2コミットを push し、
// それぞれの SHA を返す。commit1 は main.go のみ、commit2 は foo.go を追加する。
// CloneWorktrees の「1リポジトリ共有 + SHAごと worktree」の検証に使う。
func setupLocalBareRepoTwoCommits(t *testing.T) (bareURL, sha1, sha2 string) {
	t.Helper()

	bareDir := t.TempDir()
	if err := exec.Command("git", "init", "--bare", bareDir).Run(); err != nil {
		t.Fatalf("git init --bare: %v", err)
	}

	workDir := t.TempDir()
	runIn := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
		}
	}
	revParse := func(dir string) string {
		t.Helper()
		out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
		if err != nil {
			t.Fatalf("rev-parse: %v", err)
		}
		return strings.TrimSpace(string(out))
	}

	runIn(workDir, "init")
	runIn(workDir, "config", "user.email", "test@example.com")
	runIn(workDir, "config", "user.name", "Test")

	if err := os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module example.com/test\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(workDir, "add", ".")
	runIn(workDir, "commit", "-m", "commit1")
	sha1 = revParse(workDir)

	if err := os.WriteFile(filepath.Join(workDir, "foo.go"), []byte("package main\n\nfunc Foo() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(workDir, "add", ".")
	runIn(workDir, "commit", "-m", "commit2")
	sha2 = revParse(workDir)

	runIn(workDir, "remote", "add", "origin", bareDir)
	runIn(workDir, "push", "origin", "HEAD:refs/heads/main")

	return "file://" + bareDir, sha1, sha2
}

// TestGitCloner_CloneWorktrees_SharedObjectStore は CloneWorktrees が依存する
// git の手順（1回の fetch で複数 SHA 取得 → checkout --detach → worktree add --detach）が
// オブジェクトストアを共有した状態で正しく各 SHA のツリーを展開することを検証する。
// GitCloner は github URL を組み立てるため（既存 Clone テストと同様）、ここでは
// file:// リモートに対して runGit で同じコマンド列を実行して確認する。
// ローカル bare リポジトリは既定で部分クローン(--filter)を許可しないため、
// 既存のローカル fetch テストに倣い --filter=blob:none は付けない（worktree の
// 機構検証には影響しない）。
func TestGitCloner_CloneWorktrees_SharedObjectStore(t *testing.T) {
	bareURL, sha1, sha2 := setupLocalBareRepoTwoCommits(t)

	cloner := &GitCloner{}
	parent := t.TempDir()
	repoDir := filepath.Join(parent, "repo")
	if err := os.Mkdir(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		t.Helper()
		if err := cloner.runGit(context.Background(), "git", repoDir, args...); err != nil {
			t.Fatalf("runGit %v: %v", args, err)
		}
	}

	run("init")
	run("remote", "add", "origin", bareURL)
	// 2 SHA を1回の fetch でまとめて取得（CloneWorktrees と同じ集約）
	run("fetch", "--depth=1", "--no-tags", "origin", sha1, sha2)
	// 先頭 SHA は repoDir を作業ツリーとして使う
	run("checkout", "--detach", sha1)
	// 残りは worktree で追加（同一オブジェクトストアを共有）
	wt1 := filepath.Join(parent, "wt-1")
	run("worktree", "add", "--detach", wt1, sha2)

	// sha1 の worktree(repoDir) は main.go のみ、foo.go はない
	if _, err := os.Stat(filepath.Join(repoDir, "main.go")); err != nil {
		t.Errorf("repoDir(sha1) main.go not found: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repoDir, "foo.go")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("repoDir(sha1) should NOT contain foo.go, got: %v", err)
	}
	// sha2 の worktree は foo.go を含む
	if _, err := os.Stat(filepath.Join(wt1, "foo.go")); err != nil {
		t.Errorf("wt-1(sha2) foo.go not found: %v", err)
	}

	// worktree はオブジェクトストアを共有する（.git はファイルで repoDir を指す）
	gitInfo, err := os.Stat(filepath.Join(wt1, ".git"))
	if err != nil {
		t.Fatalf("wt-1 .git not found: %v", err)
	}
	if gitInfo.IsDir() {
		t.Errorf("wt-1 .git should be a file (linked worktree), got directory")
	}

	// 親ディレクトリ丸ごとの RemoveAll で cleanup が完結する
	if err := os.RemoveAll(parent); err != nil {
		t.Fatalf("cleanup RemoveAll: %v", err)
	}
	if _, err := os.Stat(parent); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected parent removed, got: %v", err)
	}
}

func TestDedupStrings(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"empty", nil, []string{}},
		{"no dup", []string{"a", "b"}, []string{"a", "b"}},
		{"dup preserves order", []string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{"same sha twice", []string{"x", "x"}, []string{"x"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dedupStrings(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("len got %v want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] got %q want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGitCloner_Clone_BadSHA(t *testing.T) {
	bareURL, _ := setupLocalBareRepo(t)

	cloner := &GitCloner{TmpDir: t.TempDir()}
	dir := t.TempDir()

	run := func(args ...string) error {
		return cloner.runGit(context.Background(), "git", dir, args...)
	}

	if err := run("init"); err != nil {
		t.Fatal(err)
	}
	if err := run("remote", "add", "origin", bareURL); err != nil {
		t.Fatal(err)
	}

	err := run("fetch", "--depth=1", "--no-tags", "origin", "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	if err == nil {
		t.Fatal("expected error for bad SHA, got nil")
	}
	if !errors.Is(err, ErrCloneFailed) {
		t.Errorf("expected ErrCloneFailed, got: %v", err)
	}
}

func TestGitCloner_Clone_TokenNotPersistedInConfig(t *testing.T) {
	bareURL, sha := setupLocalBareRepo(t)

	dir := t.TempDir()
	cloner := &GitCloner{}

	run := func(args ...string) {
		t.Helper()
		if err := cloner.runGit(context.Background(), "git", dir, args...); err != nil {
			t.Fatalf("runGit %v: %v", args, err)
		}
	}

	const fakeToken = "ghp_supersecrettoken12345"

	run("init")
	run("remote", "add", "origin", bareURL)
	// tokenをhttp.extraheader経由で渡す（gitの設定ファイルには残らないはず）
	if err := cloner.runGit(context.Background(), "git", dir,
		"-c", "http.extraheader=Authorization: bearer "+fakeToken,
		"-c", "credential.helper=",
		"fetch", "--depth=1", "--no-tags", "origin", sha,
	); err != nil {
		t.Fatal(err)
	}
	run("checkout", "FETCH_HEAD")

	// .git/config にtokenが含まれていないことを確認
	configData, err := os.ReadFile(filepath.Join(dir, ".git", "config"))
	if err != nil {
		t.Fatalf("read .git/config: %v", err)
	}
	if strings.Contains(string(configData), fakeToken) {
		t.Errorf("token found in .git/config: security regression!\n%s", configData)
	}
}
