package logging

import (
	"strings"
	"testing"
)

func TestSanitizeFrontendLogTextMasksSensitiveValues(t *testing.T) {
	input := `Authorization: Bearer abc.def.ghi password=plain "refresh_token":"secret" cookie=sessionid`

	got := sanitizeFrontendLogText(input, 500)

	if !strings.Contains(got, frontendLogMaskedValue) {
		t.Fatalf("expected masked marker in %q", got)
	}
	for _, forbidden := range []string{"abc.def.ghi", "plain", "secret", "sessionid"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("expected %q to be redacted from %q", forbidden, got)
		}
	}
}

func TestSanitizeFrontendLogURLMasksSensitiveQueryValues(t *testing.T) {
	input := "/reset?accessToken=abc123&safe=ok&password=pwd#step"

	got := sanitizeFrontendLogURL(input, 500)

	if !strings.Contains(got, "safe=ok") {
		t.Fatalf("expected non-sensitive query value to remain in %q", got)
	}
	for _, forbidden := range []string{"abc123", "pwd"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("expected %q to be redacted from %q", forbidden, got)
		}
	}
	if strings.Count(got, frontendLogMaskedValue) < 2 {
		t.Fatalf("expected sensitive query values to be masked in %q", got)
	}
}
