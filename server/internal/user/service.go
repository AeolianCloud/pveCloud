package user

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/auth"
	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
)

type Service struct {
	db     *sql.DB
	signer *auth.JWTSigner
	now    func() time.Time
}

type AuthResponse struct {
	Token       string `json:"token"`
	SubjectID   uint64 `json:"subject_id"`
	SubjectType string `json:"subject_type"`
}

type RegisterResponse struct {
	Token       string `json:"token"`
	UserID      uint64 `json:"user_id"`
	UserNo      string `json:"user_no"`
	SubjectType string `json:"subject_type"`
}

func NewService(db *sql.DB, signer *auth.JWTSigner) *Service {
	return &Service{
		db:     db,
		signer: signer,
		now:    time.Now,
	}
}

func (s *Service) Login(ctx context.Context, phone, password string) (AuthResponse, error) {
	phone = strings.TrimSpace(phone)
	password = strings.TrimSpace(password)
	if phone == "" || password == "" {
		return AuthResponse{}, errorsx.ErrBadRequest
	}

	var id uint64
	var passwordHash string
	var status string
	err := s.db.QueryRowContext(ctx, `
SELECT id, password_hash, status
FROM users
WHERE phone = ?
`, phone).Scan(&id, &passwordHash, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return AuthResponse{}, errorsx.ErrUnauthorized
	}
	if err != nil {
		return AuthResponse{}, err
	}
	if status != "active" || passwordHash != hashPassword(password) {
		return AuthResponse{}, errorsx.ErrUnauthorized
	}

	token, err := s.signer.Issue(auth.Claims{SubjectID: id, SubjectType: "user"})
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		Token:       token,
		SubjectID:   id,
		SubjectType: "user",
	}, nil
}

func (s *Service) Register(ctx context.Context, phone, email, password string) (RegisterResponse, error) {
	phone = strings.TrimSpace(phone)
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if phone == "" || password == "" {
		return RegisterResponse{}, errorsx.ErrBadRequest
	}

	userNo := fmt.Sprintf("U%d", s.now().UnixNano())
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
INSERT INTO users (user_no, email, phone, password_hash, status, created_at, updated_at)
VALUES (?, ?, ?, ?, 'active', ?, ?)
`, userNo, nullableString(email), phone, hashPassword(password), now, now)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return RegisterResponse{}, errorsx.ErrConflict
		}
		return RegisterResponse{}, err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return RegisterResponse{}, err
	}

	token, err := s.signer.Issue(auth.Claims{SubjectID: uint64(lastID), SubjectType: "user"})
	if err != nil {
		return RegisterResponse{}, err
	}

	return RegisterResponse{
		Token:       token,
		UserID:      uint64(lastID),
		UserNo:      userNo,
		SubjectType: "user",
	}, nil
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

type UserRow struct {
	ID        uint64 `json:"id"`
	UserNo    string `json:"user_no"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func (s *Service) ListUsers(ctx context.Context, limit int) ([]UserRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, user_no, email, phone, status, created_at
FROM users
ORDER BY id DESC
LIMIT ?
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []UserRow
	for rows.Next() {
		var u UserRow
		if err := rows.Scan(&u.ID, &u.UserNo, &u.Email, &u.Phone, &u.Status, &u.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, u)
	}
	return items, nil
}
