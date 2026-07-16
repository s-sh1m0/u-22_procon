package domain

import "time"

// SessionID はセッションの一意識別子
type SessionID string

// Session はログインセッションのエンティティ。
// GitHubToken は平文で保持し、永続化時の暗号化は infra 層の責務とする。
type Session struct {
	ID          SessionID
	GitHubToken string
	UserLogin   string
	CreatedAt   time.Time
}
