package analyzer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// キャンセルすると、cmd 本体だけでなく cmd が起動した子プロセスも止まること。
// go コマンドが起動した compile がタイムアウト後も残る問題の再現。
func TestKillProcessGroupOnCancel_KillsChildren(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// sh が子の sleep を起動し、その pid を出力してから待つ。
	cmd := exec.CommandContext(ctx, "sh", "-c", "sleep 60 & echo $!; wait")
	killProcessGroupOnCancel(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	var childPID int
	if _, err := fmt.Fscan(stdout, &childPID); err != nil {
		t.Fatalf("read child pid: %v", err)
	}

	cancel()
	_ = cmd.Wait()

	deadline := time.Now().Add(5 * time.Second)
	for !processExited(childPID) {
		if time.Now().After(deadline) {
			_ = (&os.Process{Pid: childPID}).Kill()
			t.Fatalf("child process %d is still running after cancel", childPID)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// processExited は pid のプロセスが存在しないか、終了済み（ゾンビ）なら true を返す。
// 親を失った子はテストを動かすプロセスに回収されずゾンビで残ることがあるため、
// /proc/<pid>/stat の状態で判定する。
func processExited(pid int) bool {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return true
	}
	// 形式: "pid (comm) state ..."。comm に空白や括弧が入り得るため最後の ")" から読む。
	s := string(b)
	i := strings.LastIndex(s, ")")
	return i >= 0 && i+2 < len(s) && s[i+2] == 'Z'
}
