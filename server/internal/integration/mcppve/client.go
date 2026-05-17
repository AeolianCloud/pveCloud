package mcppve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
)

type Client struct {
	baseURL    *url.URL
	token      string
	httpClient *http.Client
	enabled    bool
}

type CreateVMRequest struct {
	VMID            uint     `json:"vmid"`
	Name            string   `json:"name"`
	Cores           int      `json:"cores"`
	Memory          int      `json:"memory"`
	Storage         string   `json:"storage"`
	DiskSource      string   `json:"diskSource"`
	DiskFormat      string   `json:"diskFormat,omitempty"`
	DiskInterface   string   `json:"diskInterface,omitempty"`
	Network         string   `json:"network,omitempty"`
	CIUser          string   `json:"ciUser,omitempty"`
	SSHKeys         string   `json:"sshKeys,omitempty"`
	IPConfig0       string   `json:"ipConfig0,omitempty"`
	Nameserver      string   `json:"nameserver,omitempty"`
	SearchDomain    string   `json:"searchDomain,omitempty"`
	SnippetsStorage string   `json:"snippetsStorage,omitempty"`
	CIPackages      []string `json:"ciPackages,omitempty"`
	AptMirror       string   `json:"aptMirror,omitempty"`
}

type AsyncAccepted struct {
	Location          string
	OperationLocation string
	OperationID       string
}

type VM struct {
	VMID   uint   `json:"vmid"`
	Name   string `json:"name"`
	Status string `json:"status"`
	CPUs   int    `json:"cpus"`
	Mem    int64  `json:"mem"`
	MaxMem int64  `json:"maxmem"`
	Raw    map[string]any
}

type Operation struct {
	ID               string          `json:"id"`
	Status           string          `json:"status"`
	ResourceLocation string          `json:"resourceLocation"`
	Error            *OperationError `json:"error"`
}

type OperationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error *OperationError `json:"error"`
}

func NewClient(cfg config.MCPPVEConfig) (*Client, error) {
	base, err := url.Parse(strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"))
	if err != nil {
		return nil, fmt.Errorf("解析虚拟化接口地址失败: %w", err)
	}
	return &Client{
		baseURL:    base,
		token:      strings.TrimSpace(cfg.BearerToken),
		httpClient: &http.Client{Timeout: cfg.Timeout()},
		enabled:    cfg.Enabled,
	}, nil
}

func (c *Client) Enabled() bool {
	return c != nil && c.enabled
}

func (c *Client) Nodes(ctx context.Context) (any, error) {
	var out any
	err := c.doJSON(ctx, http.MethodGet, "/api/pve/nodes", nil, &out, nil)
	return out, err
}

func (c *Client) Node(ctx context.Context, node string) (any, error) {
	var out any
	err := c.doJSON(ctx, http.MethodGet, "/api/pve/nodes/"+url.PathEscape(node), nil, &out, nil)
	return out, err
}

func (c *Client) NodeVMs(ctx context.Context, node string) (any, error) {
	var out any
	err := c.doJSON(ctx, http.MethodGet, "/api/pve/nodes/"+url.PathEscape(node)+"/vms", nil, &out, nil)
	return out, err
}

func (c *Client) Storage(ctx context.Context) (any, error) {
	var out any
	err := c.doJSON(ctx, http.MethodGet, "/api/pve/storage", nil, &out, nil)
	return out, err
}

func (c *Client) CreateVM(ctx context.Context, node string, req CreateVMRequest) (AsyncAccepted, error) {
	var accepted AsyncAccepted
	err := c.doJSON(ctx, http.MethodPost, "/api/pve/nodes/"+url.PathEscape(node)+"/vms", req, nil, &accepted)
	return accepted, err
}

func (c *Client) VM(ctx context.Context, node string, vmid uint) (VM, error) {
	var raw map[string]any
	err := c.doJSON(ctx, http.MethodGet, "/api/pve/nodes/"+url.PathEscape(node)+"/vms/"+strconv.FormatUint(uint64(vmid), 10), nil, &raw, nil)
	if err != nil {
		return VM{}, err
	}
	return vmFromRaw(raw), nil
}

func (c *Client) StartVM(ctx context.Context, node string, vmid uint) (AsyncAccepted, error) {
	var accepted AsyncAccepted
	err := c.doJSON(ctx, http.MethodPost, "/api/pve/nodes/"+url.PathEscape(node)+"/vms/"+strconv.FormatUint(uint64(vmid), 10)+"/start", nil, nil, &accepted)
	return accepted, err
}

func (c *Client) StopVM(ctx context.Context, node string, vmid uint) (AsyncAccepted, error) {
	var accepted AsyncAccepted
	err := c.doJSON(ctx, http.MethodPost, "/api/pve/nodes/"+url.PathEscape(node)+"/vms/"+strconv.FormatUint(uint64(vmid), 10)+"/stop", nil, nil, &accepted)
	return accepted, err
}

func (c *Client) DeleteVM(ctx context.Context, node string, vmid uint) (AsyncAccepted, error) {
	var accepted AsyncAccepted
	err := c.doJSON(ctx, http.MethodDelete, "/api/pve/nodes/"+url.PathEscape(node)+"/vms/"+strconv.FormatUint(uint64(vmid), 10), nil, nil, &accepted)
	return accepted, err
}

func (c *Client) Operation(ctx context.Context, id string) (Operation, error) {
	var out Operation
	err := c.doJSON(ctx, http.MethodGet, "/api/pve/operations/"+url.PathEscape(id), nil, &out, nil)
	return out, err
}

func (c *Client) doJSON(ctx context.Context, method string, relPath string, input any, output any, accepted *AsyncAccepted) error {
	if !c.Enabled() {
		return &UnavailableError{Message: "虚拟化管理接口未启用"}
	}
	var body io.Reader
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.endpoint(relPath), body)
	if err != nil {
		return err
	}
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &UnavailableError{Message: "虚拟化管理接口请求失败"}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return parseError(resp)
	}
	if accepted != nil {
		accepted.Location = resp.Header.Get("Location")
		accepted.OperationLocation = resp.Header.Get("Operation-Location")
		accepted.OperationID = operationID(accepted.OperationLocation)
	}
	if output == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(output)
}

func (c *Client) endpoint(relPath string) string {
	copied := *c.baseURL
	copied.Path = path.Join(c.baseURL.Path, relPath)
	return copied.String()
}

func parseError(resp *http.Response) error {
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var parsed ErrorResponse
	if err := json.Unmarshal(data, &parsed); err == nil && parsed.Error != nil {
		return &UpstreamError{StatusCode: resp.StatusCode, Code: parsed.Error.Code, Message: parsed.Error.Message}
	}
	return &UpstreamError{StatusCode: resp.StatusCode, Message: "虚拟化管理接口返回错误"}
}

func operationID(location string) string {
	location = strings.TrimSpace(location)
	if location == "" {
		return ""
	}
	return path.Base(location)
}

func vmFromRaw(raw map[string]any) VM {
	vm := VM{Raw: raw}
	if value, ok := raw["vmid"].(float64); ok {
		vm.VMID = uint(value)
	}
	if value, ok := raw["name"].(string); ok {
		vm.Name = value
	}
	if value, ok := raw["status"].(string); ok {
		vm.Status = strings.TrimSpace(value)
	}
	if value, ok := raw["cpus"].(float64); ok {
		vm.CPUs = int(value)
	}
	if value, ok := raw["mem"].(float64); ok {
		vm.Mem = int64(value)
	}
	if value, ok := raw["maxmem"].(float64); ok {
		vm.MaxMem = int64(value)
	}
	return vm
}

type UnavailableError struct {
	Message string
}

func (e *UnavailableError) Error() string {
	if strings.TrimSpace(e.Message) == "" {
		return "虚拟化管理接口不可用"
	}
	return e.Message
}

type UpstreamError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *UpstreamError) Error() string {
	if strings.TrimSpace(e.Message) == "" {
		return "虚拟化管理接口返回错误"
	}
	return e.Message
}
