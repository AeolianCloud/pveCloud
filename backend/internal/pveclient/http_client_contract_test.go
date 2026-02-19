package pveclient

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPClientCreateInstanceSignsRequest(t *testing.T) {
	t.Parallel()

	secret := "contract-secret"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/instances" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		payload, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()

		timestamp := r.Header.Get("X-PVE-API-TIMESTAMP")
		signature := r.Header.Get("X-PVE-API-SIGNATURE")
		if timestamp == "" || signature == "" {
			t.Fatalf("missing signature headers")
		}

		expected := calcSignature(secret, r.Method, r.URL.Path, timestamp, payload)
		if signature != expected {
			t.Fatalf("signature mismatch expected=%s got=%s", expected, signature)
		}

		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"task_id":"t-1","pve_task_id":"pve-1","instance_id":"vm-1","description":"queued"}}`))
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{
		BaseURL:             server.URL,
		APIKey:              "api-key",
		APISecret:           secret,
		Timeout:             3 * time.Second,
		MaxRetries:          0,
		RetryBackoff:        10 * time.Millisecond,
		CircuitOpenDuration: 3 * time.Second,
	})

	result, err := client.CreateInstance(context.Background(), CreateInstanceReq{Name: "vm-a", CPU: 2, MemoryMB: 2048})
	if err != nil {
		t.Fatalf("CreateInstance error: %v", err)
	}
	if result.TaskID != "t-1" || result.InstanceID != "vm-1" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestHTTPClientRetryOnServerError(t *testing.T) {
	t.Parallel()

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&calls, 1)
		if current < 3 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"message":"temporary upstream error"}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"task_id":"t-retry","pve_task_id":"pve-retry","description":"ok"}}`))
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{
		BaseURL:             server.URL,
		APIKey:              "api-key",
		APISecret:           "secret",
		Timeout:             3 * time.Second,
		MaxRetries:          2,
		RetryBackoff:        5 * time.Millisecond,
		CircuitOpenDuration: 2 * time.Second,
	})

	_, err := client.StartInstance(context.Background(), "inst-1")
	if err != nil {
		t.Fatalf("StartInstance should succeed after retries: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

func TestHTTPClientCircuitBreakerOpens(t *testing.T) {
	t.Parallel()

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"always fail"}`))
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{
		BaseURL:                 server.URL,
		APIKey:                  "api-key",
		APISecret:               "secret",
		Timeout:                 3 * time.Second,
		MaxRetries:              0,
		RetryBackoff:            5 * time.Millisecond,
		CircuitFailureThreshold: 2,
		CircuitOpenDuration:     1 * time.Minute,
	})

	_, _ = client.StopInstance(context.Background(), "inst-1")
	_, _ = client.StopInstance(context.Background(), "inst-1")

	_, err := client.StopInstance(context.Background(), "inst-1")
	if err == nil {
		t.Fatalf("expected circuit open error")
	}
	if err != ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected only 2 upstream calls before circuit opened, got %d", got)
	}
}

func calcSignature(secret, method, path, timestamp string, payload []byte) string {
	bodyHash := sha256.Sum256(payload)
	signPayload := method + "\n" + path + "\n" + timestamp + "\n" + hex.EncodeToString(bodyHash[:])
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signPayload))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestHTTPClientParseDirectJSONBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := NodeStatus{Node: "node-a", CPUUsage: 0.5, MemoryUsage: 0.6, DiskUsage: 0.7}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{
		BaseURL:      server.URL,
		APIKey:       "api-key",
		APISecret:    "secret",
		Timeout:      3 * time.Second,
		MaxRetries:   0,
		RetryBackoff: 5 * time.Millisecond,
	})

	status, err := client.GetNodeStatus(context.Background(), "node-a")
	if err != nil {
		t.Fatalf("GetNodeStatus error: %v", err)
	}
	if status.Node != "node-a" {
		t.Fatalf("unexpected node status: %+v", status)
	}
}
