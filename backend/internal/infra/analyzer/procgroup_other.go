//go:build !unix

package analyzer

import "os/exec"

// killProcessGroupOnCancel は Unix 以外では何もしない（キャンセル時は exec の既定どおり
// cmd 本体だけを kill する）。解析サーバは Linux コンテナでのみ動かす。
func killProcessGroupOnCancel(_ *exec.Cmd) {}
