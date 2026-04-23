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

type AdminRow struct {
	ID        uint64 `json:"id"`
	AdminNo   string `json:"admin_no"`
	Username  string `json:"username"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func (s *Service) ListAdmins(ctx context.Context, limit int) ([]AdminRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, admin_no, username, status, created_at
FROM admins
ORDER BY id DESC
LIMIT ?
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AdminRow
	for rows.Next() {
		var a AdminRow
		if err := rows.Scan(&a.ID, &a.AdminNo, &a.Username, &a.Status, &a.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, nil
}
