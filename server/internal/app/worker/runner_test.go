package worker

import (
	"testing"
	"time"
)

func TestRetryDelayUsesBoundedQuadraticBackoff(t *testing.T) {
	cases := []struct {
		attempts int
		want     time.Duration
	}{
		{attempts: 0, want: time.Minute},
		{attempts: 1, want: time.Minute},
		{attempts: 2, want: 4 * time.Minute},
		{attempts: 6, want: 36 * time.Minute},
		{attempts: 99, want: 36 * time.Minute},
	}

	for _, tc := range cases {
		if got := retryDelay(tc.attempts); got != tc.want {
			t.Fatalf("retryDelay(%d) = %s, want %s", tc.attempts, got, tc.want)
		}
	}
}

func TestSameExpiresAtMatchesMillisecondPrecision(t *testing.T) {
	value := time.Date(2026, 5, 23, 12, 0, 0, 123456789, time.UTC)
	encoded := time.Date(2026, 5, 23, 12, 0, 0, 123000000, time.UTC).Format(time.RFC3339Nano)

	if !sameExpiresAt(&value, encoded) {
		t.Fatal("expiry task payload should match database time at millisecond precision")
	}
	if sameExpiresAt(&value, time.Date(2026, 5, 23, 12, 0, 0, 124000000, time.UTC).Format(time.RFC3339Nano)) {
		t.Fatal("different expiry timestamp must not match")
	}
	if sameExpiresAt(nil, encoded) {
		t.Fatal("nil instance expiry must not match a versioned lifecycle task")
	}
	if !sameExpiresAt(nil, "") {
		t.Fatal("empty encoded expiry keeps backward-compatible unversioned task matching")
	}
}

func TestParsePayloadIgnoresInvalidJSON(t *testing.T) {
	raw := `{"instance_no":" INS-1 ","expires_at":"2026-05-23T12:00:00Z","notification_no":"NTF-1"}`
	payload := parsePayload(&raw)
	if payload.InstanceNo != " INS-1 " || payload.ExpiresAt == "" || payload.NotificationNo != "NTF-1" {
		t.Fatalf("payload not decoded as expected: %#v", payload)
	}

	invalid := "{"
	if got := parsePayload(&invalid); got != (taskPayload{}) {
		t.Fatalf("invalid payload should decode to empty payload, got %#v", got)
	}
}
