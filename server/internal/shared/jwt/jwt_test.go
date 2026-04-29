package jwt

import (
	"testing"
	"time"
)

func TestSignAndParseAdminClaims(t *testing.T) {
	claims := Claims{
		TokenType:        "admin",
		AdminID:          42,
		PermissionCodes:  []string{"dashboard:view"},
		RegisteredClaims: NewRegisteredClaims("pvecloud-admin", time.Minute),
	}

	token, err := Sign(claims, "secret")
	if err != nil {
		t.Fatalf("expected token signed, got %v", err)
	}
	parsed, err := Parse(token, "secret")
	if err != nil {
		t.Fatalf("expected token parsed, got %v", err)
	}
	if parsed.TokenType != "admin" || parsed.AdminID != 42 || parsed.Issuer != "pvecloud-admin" {
		t.Fatalf("unexpected claims: %#v", parsed)
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	token, err := Sign(Claims{RegisteredClaims: NewRegisteredClaims("issuer", time.Minute)}, "secret")
	if err != nil {
		t.Fatalf("expected token signed, got %v", err)
	}
	if _, err := Parse(token, "wrong-secret"); err == nil {
		t.Fatal("expected wrong secret rejected")
	}
}
