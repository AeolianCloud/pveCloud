package auth_test

import (
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/auth"
)

func TestIssueAndParseToken(t *testing.T) {
	signer := auth.NewJWTSigner("web-secret")
	token, err := signer.Issue(auth.Claims{SubjectID: 1001, SubjectType: "user"})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	claims, err := signer.Parse(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	if claims.SubjectID != 1001 {
		t.Fatalf("expected subject id 1001, got %d", claims.SubjectID)
	}
	if claims.SubjectType != "user" {
		t.Fatalf("expected subject type user, got %s", claims.SubjectType)
	}
}
