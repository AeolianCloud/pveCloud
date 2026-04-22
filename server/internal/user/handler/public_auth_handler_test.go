package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/user"
	"github.com/AeolianCloud/pveCloud/server/internal/user/handler"
)

type fakeAuthService struct {
	login func(ctx context.Context, phone, password string) (user.AuthResponse, error)
}

func (f fakeAuthService) Login(ctx context.Context, phone, password string) (user.AuthResponse, error) {
	return f.login(ctx, phone, password)
}

func TestLoginReturnsToken(t *testing.T) {
	h := handler.NewAuthHandler(fakeAuthService{
		login: func(ctx context.Context, phone, password string) (user.AuthResponse, error) {
			if phone != "13800000000" {
				t.Fatalf("unexpected phone: %s", phone)
			}
			if password != "secret" {
				t.Fatalf("unexpected password: %s", password)
			}
			return user.AuthResponse{
				Token:       "signed-token",
				SubjectID:   1001,
				SubjectType: "user",
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"phone":"13800000000","password":"secret"}`))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp user.AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token != "signed-token" {
		t.Fatalf("expected token signed-token, got %s", resp.Token)
	}
	if resp.SubjectID != 1001 {
		t.Fatalf("expected subject id 1001, got %d", resp.SubjectID)
	}
}

func TestLoginRejectsInvalidJSON(t *testing.T) {
	h := handler.NewAuthHandler(fakeAuthService{
		login: func(ctx context.Context, phone, password string) (user.AuthResponse, error) {
			t.Fatal("login should not be called")
			return user.AuthResponse{}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{`))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(errorsx.ErrBadRequest.Code)) {
		t.Fatalf("expected bad request code in response, got %s", rec.Body.String())
	}
}
