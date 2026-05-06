package job

import "github.com/s-sh1m0/u-22_procon/backend/internal/domain"

// Item はジョブキューの 1 要素。Token は in-memory のみで永続化しない。
type Item struct {
	JobID domain.JobID
	Token string
}

// Queue は in-process のジョブキュー。buffered channel で実装する。
type Queue struct {
	ch chan Item
}

// NewQueue は容量 buf のキューを返す。
func NewQueue(buf int) *Queue {
	return &Queue{ch: make(chan Item, buf)}
}

// Enqueue はジョブをキューに追加する。満杯の場合はブロックする。
// この関数シグネチャは usecase.JobEnqueuer インターフェースを充足する。
func (q *Queue) Enqueue(jobID domain.JobID, token string) {
	q.ch <- Item{JobID: jobID, Token: token}
}

// Recv はキューの受信チャネルを返す。
func (q *Queue) Recv() <-chan Item {
	return q.ch
}

// Close はキューをクローズし、ワーカーに終了を通知する。
func (q *Queue) Close() {
	close(q.ch)
}
