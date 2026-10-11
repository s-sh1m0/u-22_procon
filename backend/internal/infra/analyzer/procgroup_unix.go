//go:build unix

package analyzer

import (
	"os/exec"
	"syscall"
)

// killProcessGroupOnCancel は cmd を独立したプロセスグループで起動させ、ctx のキャンセル時に
// グループ全体を SIGKILL するよう設定する。go コマンドは compile などの子プロセスを起動するが、
// 親の go だけを kill すると子は止まらずに残り、CPU とメモリを使い続けるため。
func killProcessGroupOnCancel(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		// Setpgid で pgid = pid になっているので、負の pid でグループ全体に送る。
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
