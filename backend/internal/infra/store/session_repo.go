package store

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// SessionRepo は domain.SessionRepository の SQLite 実装。
// GitHub アクセストークンは AES-GCM で暗号化してから保存する。
type SessionRepo struct {
	db   *sql.DB
	aead cipher.AEAD
}

// NewSessionRepo は secret から SHA-256 で 32 バイト鍵を派生し、AES-GCM AEAD を構築する。
func NewSessionRepo(db *sql.DB, secret []byte) (*SessionRepo, error) {
	if len(secret) == 0 {
		return nil, errors.New("session secret must not be empty")
	}
	key := sha256.Sum256(secret)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("aes new cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	return &SessionRepo{db: db, aead: aead}, nil
}

func (r *SessionRepo) encrypt(plaintext string) (string, error) {
	nonce := make([]byte, r.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	ct := r.aead.Seal(nil, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(append(nonce, ct...)), nil
}

func (r *SessionRepo) decrypt(encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}
	ns := r.aead.NonceSize()
	if len(raw) < ns {
		return "", errors.New("ciphertext too short")
	}
	nonce, ct := raw[:ns], raw[ns:]
	pt, err := r.aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("aead open: %w", err)
	}
	return string(pt), nil
}

func (r *SessionRepo) Save(ctx context.Context, s *domain.Session) error {
	encrypted, err := r.encrypt(s.GitHubToken)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO sessions (id, github_token, user_login, created_at) VALUES (?, ?, ?, ?)`,
		string(s.ID), encrypted, s.UserLogin, s.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (r *SessionRepo) FindByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	var (
		encrypted string
		login     string
		createdAt time.Time
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT github_token, user_login, created_at FROM sessions WHERE id = ?`,
		string(id),
	).Scan(&encrypted, &login, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query session: %w", err)
	}
	token, err := r.decrypt(encrypted)
	if err != nil {
		return nil, err
	}
	return &domain.Session{
		ID:          id,
		GitHubToken: token,
		UserLogin:   login,
		CreatedAt:   createdAt,
	}, nil
}

func (r *SessionRepo) Delete(ctx context.Context, id domain.SessionID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, string(id))
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
