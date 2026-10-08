package usecase

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// AuthenticateTokenUseCase は Authorization: Bearer で渡された GitHub トークンを検証し、
// DB に保存しないセッションを組み立てるユースケース（VS Code 拡張からの API 呼び出し用）。
//
// 検証結果はトークンの SHA-256 をキーに TTL の間だけメモリに保持する。
// ジョブ状態のポーリング（1.5 秒間隔）のたびに GitHub /user を叩かないため。
// トークン平文はキャッシュのキーに使わない。
type AuthenticateTokenUseCase struct {
	users domain.GitHubUserRepository
	ttl   time.Duration
	now   func() time.Time

	mu    sync.Mutex
	cache map[[sha256.Size]byte]tokenCacheEntry
}

type tokenCacheEntry struct {
	login     string
	expiresAt time.Time
}

// NewAuthenticateTokenUseCase は AuthenticateTokenUseCase を返す。
// ttl は検証結果をキャッシュする期間。
func NewAuthenticateTokenUseCase(users domain.GitHubUserRepository, ttl time.Duration) *AuthenticateTokenUseCase {
	return &AuthenticateTokenUseCase{
		users: users,
		ttl:   ttl,
		now:   time.Now,
		cache: make(map[[sha256.Size]byte]tokenCacheEntry),
	}
}

// Execute はトークンを検証し、ID を持たない（永続化しない）セッションを返す。
// トークンが無効な場合は domain.ErrInvalidToken を返す。
func (uc *AuthenticateTokenUseCase) Execute(ctx context.Context, token string) (*domain.Session, error) {
	if token == "" {
		return nil, domain.ErrInvalidToken
	}
	key := sha256.Sum256([]byte(token))
	now := uc.now()

	uc.mu.Lock()
	entry, ok := uc.cache[key]
	uc.mu.Unlock()
	if ok && now.Before(entry.expiresAt) {
		return &domain.Session{GitHubToken: token, UserLogin: entry.login, CreatedAt: now}, nil
	}

	login, err := uc.users.LoginByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("resolve github user: %w", err)
	}

	uc.mu.Lock()
	uc.pruneExpiredLocked(now)
	uc.cache[key] = tokenCacheEntry{login: login, expiresAt: now.Add(uc.ttl)}
	uc.mu.Unlock()

	return &domain.Session{GitHubToken: token, UserLogin: login, CreatedAt: now}, nil
}

// pruneExpiredLocked は期限切れのエントリを捨ててキャッシュの肥大化を防ぐ。uc.mu を保持して呼ぶ。
func (uc *AuthenticateTokenUseCase) pruneExpiredLocked(now time.Time) {
	for k, e := range uc.cache {
		if !now.Before(e.expiresAt) {
			delete(uc.cache, k)
		}
	}
}
