package auth

import (
	"testing"
	"time"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
)

func TestAdminAuthMeBuildsStateFromCurrentPermissions(t *testing.T) {
	service := &AdminAuthService{}
	session := admindto.SessionSummary{
		SessionID: "adm_session",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	state := service.Me(
		models.AdminUser{ID: 1, Username: "admin", DisplayName: "Root", Status: adminStatusActive},
		[]uint64{1},
		[]string{"page.dashboard"},
		session,
	)

	if state.Admin.ID != 1 || state.Session.SessionID != "adm_session" {
		t.Fatalf("unexpected auth state: %#v", state)
	}
	if len(state.Menus) != 1 || state.Menus[0].Key != "dashboard" {
		t.Fatalf("expected dashboard menu only, got %#v", state.Menus)
	}
}
