package adminuser

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/AeolianCloud/pveCloud/server/internal/auth"
	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
)

type Service struct {
	db     *sql.DB
	signer *auth.JWTSigner
}

type AuthResponse struct {
	Token       string `json:"token"`
	SubjectID   uint64 `json:"subject_id"`
	SubjectType string `json:"subject_type"`
}

func NewService(db *sql.DB, signer *auth.JWTSigner) *Service {
	return &Service{db: db, signer: signer}
}

func (s *Service) Login(ctx context.Context, username, password string) (AuthResponse, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return AuthResponse{}, errorsx.ErrBadRequest
	}

	var id uint64
	var passwordHash string
	var status string
	err := s.db.QueryRowContext(ctx, `
SELECT id, password_hash, status
FROM admins
WHERE username = ?
`, username).Scan(&id, &passwordHash, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return AuthResponse{}, errorsx.ErrUnauthorized
	}
	if err != nil {
		return AuthResponse{}, err
	}
	if status != "active" || passwordHash != hashPassword(password) {
		return AuthResponse{}, errorsx.ErrUnauthorized
	}

	token, err := s.signer.Issue(auth.Claims{SubjectID: id, SubjectType: "admin"})
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		Token:       token,
		SubjectID:   id,
		SubjectType: "admin",
	}, nil
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}
