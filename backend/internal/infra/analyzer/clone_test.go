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
