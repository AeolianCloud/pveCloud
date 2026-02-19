package pveclient

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ErrCircuitOpen 表示熔断器处于打开状态，当前请求被快速失败。
var ErrCircuitOpen = errors.New("pve http client circuit is open")

// HTTPClientConfig 控制真实 PVE HTTP 客户端的行为参数。
type HTTPClientConfig struct {
	BaseURL                 string
	APIKey                  string
	APISecret               string
	Timeout                 time.Duration
	MaxRetries              int
	RetryBackoff            time.Duration
	CircuitFailureThreshold int
	CircuitOpenDuration     time.Duration
}

// HttpPVEClient 通过 HTTP 调用 PVE 能力后端，内置签名、超时、重试和熔断。
type HttpPVEClient struct {
	baseURL                 string
	apiKey                  string
	apiSecret               string
	httpClient              *http.Client
	maxRetries              int
	retryBackoff            time.Duration
	circuitFailureThreshold int
	circuitOpenDuration     time.Duration
	breaker                 *circuitBreaker
}

// circuitBreaker 是最小化实现的熔断器。
type circuitBreaker struct {
	mu        sync.Mutex
	failures  int
	openUntil time.Time
}

// apiEnvelope 用于解析 PVE 后端统一响应格式。
type apiEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// NewHTTPClient 创建可用于生产对接的 PVE HTTP 客户端。
func NewHTTPClient(cfg HTTPClientConfig) *HttpPVEClient {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 8 * time.Second
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = 200 * time.Millisecond
	}
	if cfg.CircuitFailureThreshold <= 0 {
		cfg.CircuitFailureThreshold = 5
	}
	if cfg.CircuitOpenDuration <= 0 {
		cfg.CircuitOpenDuration = 30 * time.Second
	}

	return &HttpPVEClient{
		baseURL:                 strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:                  cfg.APIKey,
		apiSecret:               cfg.APISecret,
		httpClient:              &http.Client{Timeout: cfg.Timeout},
		maxRetries:              cfg.MaxRetries,
		retryBackoff:            cfg.RetryBackoff,
		circuitFailureThreshold: cfg.CircuitFailureThreshold,
		circuitOpenDuration:     cfg.CircuitOpenDuration,
		breaker:                 &circuitBreaker{},
	}
}

func (h *HttpPVEClient) CreateInstance(ctx context.Context, req CreateInstanceReq) (*TaskResult, error) {
	var result TaskResult
	if err := h.doJSON(ctx, http.MethodPost, "/instances", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *HttpPVEClient) StartInstance(ctx context.Context, instanceID string) (*TaskResult, error) {
	path := "/instances/" + url.PathEscape(instanceID) + "/start"
	var result TaskResult
	if err := h.doJSON(ctx, http.MethodPost, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *HttpPVEClient) StopInstance(ctx context.Context, instanceID string) (*TaskResult, error) {
	path := "/instances/" + url.PathEscape(instanceID) + "/stop"
	var result TaskResult
	if err := h.doJSON(ctx, http.MethodPost, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *HttpPVEClient) RebootInstance(ctx context.Context, instanceID string) (*TaskResult, error) {
	path := "/instances/" + url.PathEscape(instanceID) + "/reboot"
	var result TaskResult
	if err := h.doJSON(ctx, http.MethodPost, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *HttpPVEClient) DeleteInstance(ctx context.Context, instanceID string) (*TaskResult, error) {
	path := "/instances/" + url.PathEscape(instanceID)
	var result TaskResult
	if err := h.doJSON(ctx, http.MethodDelete, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *HttpPVEClient) GetInstanceStatus(ctx context.Context, instanceID string) (*InstanceStatus, error) {
	path := "/instances/" + url.PathEscape(instanceID) + "/status"
	var result InstanceStatus
	if err := h.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *HttpPVEClient) GetInstanceMetrics(ctx context.Context, instanceID string) (*Metrics, error) {
	path := "/instances/" + url.PathEscape(instanceID) + "/metrics"
	var result Metrics
	if err := h.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *HttpPVEClient) GetConsoleToken(ctx context.Context, instanceID string) (*ConsoleInfo, error) {
	path := "/instances/" + url.PathEscape(instanceID) + "/console-token"
	var result ConsoleInfo
	if err := h.doJSON(ctx, http.MethodPost, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *HttpPVEClient) CreateSnapshot(ctx context.Context, instanceID, name string) (*TaskResult, error) {
	path := "/instances/" + url.PathEscape(instanceID) + "/snapshots"
	payload := map[string]string{"name": name}
	var result TaskResult
	if err := h.doJSON(ctx, http.MethodPost, path, payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *HttpPVEClient) RestoreSnapshot(ctx context.Context, instanceID, snapshotName string) (*TaskResult, error) {
	path := "/instances/" + url.PathEscape(instanceID) + "/snapshots/" + url.PathEscape(snapshotName) + "/restore"
	var result TaskResult
	if err := h.doJSON(ctx, http.MethodPost, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *HttpPVEClient) ListSnapshots(ctx context.Context, instanceID string) ([]*Snapshot, error) {
	path := "/instances/" + url.PathEscape(instanceID) + "/snapshots"
	var result []*Snapshot
	if err := h.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (h *HttpPVEClient) DeleteSnapshot(ctx context.Context, instanceID, snapshotName string) (*TaskResult, error) {
	path := "/instances/" + url.PathEscape(instanceID) + "/snapshots/" + url.PathEscape(snapshotName)
	var result TaskResult
	if err := h.doJSON(ctx, http.MethodDelete, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *HttpPVEClient) GetTaskStatus(ctx context.Context, taskID string) (*TaskStatus, error) {
	path := "/tasks/" + url.PathEscape(taskID)
	var result TaskStatus
	if err := h.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *HttpPVEClient) GetNodeStatus(ctx context.Context, node string) (*NodeStatus, error) {
	path := "/nodes/" + url.PathEscape(node) + "/status"
	var result NodeStatus
	if err := h.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *HttpPVEClient) doJSON(ctx context.Context, method, path string, reqBody interface{}, out interface{}) error {
	payloadBytes, err := marshalRequest(reqBody)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= h.maxRetries; attempt++ {
		if err := h.breaker.allowRequest(); err != nil {
			return err
		}

		respBody, statusCode, requestErr := h.doOnce(ctx, method, path, payloadBytes)
		if requestErr == nil {
			if statusCode >= 500 {
				h.breaker.recordFailure(h.circuitFailureThreshold, h.circuitOpenDuration)
				lastErr = fmt.Errorf("pve server error status=%d", statusCode)
			} else if statusCode >= 400 {
				h.breaker.recordSuccess()
				return fmt.Errorf("pve client error status=%d body=%s", statusCode, string(respBody))
			} else {
				h.breaker.recordSuccess()
				if err := unmarshalResponse(respBody, out); err != nil {
					return err
				}
				return nil
			}
		} else {
			h.breaker.recordFailure(h.circuitFailureThreshold, h.circuitOpenDuration)
			lastErr = requestErr
		}

		if attempt == h.maxRetries {
			break
		}
		if sleepErr := sleepWithContext(ctx, h.retryBackoff*time.Duration(1<<attempt)); sleepErr != nil {
			return sleepErr
		}
	}

	if lastErr == nil {
		return errors.New("pve request failed")
	}
	return fmt.Errorf("pve request failed after retry: %w", lastErr)
}

func (h *HttpPVEClient) doOnce(ctx context.Context, method, path string, payload []byte) ([]byte, int, error) {
	requestURL := h.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, err
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := h.signRequest(method, path, timestamp, payload)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PVE-API-KEY", h.apiKey)
	req.Header.Set("X-PVE-API-TIMESTAMP", timestamp)
	req.Header.Set("X-PVE-API-SIGNATURE", signature)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func (h *HttpPVEClient) signRequest(method, path, timestamp string, payload []byte) string {
	bodyHash := sha256.Sum256(payload)
	signPayload := method + "\n" + path + "\n" + timestamp + "\n" + hex.EncodeToString(bodyHash[:])
	mac := hmac.New(sha256.New, []byte(h.apiSecret))
	_, _ = mac.Write([]byte(signPayload))
	return hex.EncodeToString(mac.Sum(nil))
}

func marshalRequest(reqBody interface{}) ([]byte, error) {
	if reqBody == nil {
		return []byte("{}"), nil
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}
	return payload, nil
}

func unmarshalResponse(raw []byte, out interface{}) error {
	var envelope apiEnvelope
	if err := json.Unmarshal(raw, &envelope); err == nil && envelope.Message != "" {
		if envelope.Code != 0 {
			return fmt.Errorf("pve response error code=%d message=%s", envelope.Code, envelope.Message)
		}
		if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
			return nil
		}
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("unmarshal pve envelope data: %w", err)
		}
		return nil
	}

	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("unmarshal pve response body: %w", err)
	}
	return nil
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (cb *circuitBreaker) allowRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if !cb.openUntil.IsZero() && time.Now().Before(cb.openUntil) {
		return ErrCircuitOpen
	}
	if !cb.openUntil.IsZero() && time.Now().After(cb.openUntil) {
		cb.openUntil = time.Time{}
		cb.failures = 0
	}
	return nil
}

func (cb *circuitBreaker) recordFailure(threshold int, openDuration time.Duration) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	if cb.failures >= threshold {
		cb.openUntil = time.Now().Add(openDuration)
		cb.failures = 0
	}
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.openUntil = time.Time{}
}
