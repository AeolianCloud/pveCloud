package realname

import (
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	smartalipay "github.com/smartwalle/alipay"
	tccommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	tcprofile "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcfaceid "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/faceid/v20180301"
)

const (
	ProviderAlipay = "alipay"
	ProviderWechat = "wechat"

	statusPending  = "pending"
	statusApproved = "approved"
	statusRejected = "rejected"
)

type ProviderConfig struct {
	Provider        string
	AppID           string
	GatewayURL      string
	AppPrivateKey   string
	AlipayPublicKey string
	ReturnURL       string
	NotifyURL       string
	CallbackBaseURL string
	SecretID        string
	SecretKey       string
	Region          string
	Endpoint        string
	RuleID          string
	RedirectURL     string
}

type CreateSessionInput struct {
	ApplicationNo string
	RealName      string
	IDType        string
	IDNumber      string
}

type Session struct {
	Provider              string
	ProviderApplicationID string
	ActionType            string
	RedirectURL           string
	ExpiresAt             *time.Time
	TraceID               string
	ResponseDigest        string
}

type Result struct {
	ProviderStatus string
	FinalStatus    string
	ResultCode     string
	ResultMessage  string
	ResponseDigest string
	TraceID        string
}

type CallbackRequest struct {
	Method      string
	Headers     http.Header
	Query       url.Values
	Form        url.Values
	RawBody     []byte
	ContentType string
}

type Callback struct {
	ProviderApplicationID string
	PayloadDigest         string
	ReplayKey             string
	Timestamp             *time.Time
}

type Client struct {
	httpClient *http.Client
}

type UnavailableError struct {
	Message string
}

func (e *UnavailableError) Error() string {
	if strings.TrimSpace(e.Message) != "" {
		return strings.TrimSpace(e.Message)
	}
	return "实名供应商暂不可用"
}

type InvalidCallbackError struct {
	Message string
}

func (e *InvalidCallbackError) Error() string {
	if strings.TrimSpace(e.Message) != "" {
		return strings.TrimSpace(e.Message)
	}
	return "实名供应商回调无效"
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{httpClient: httpClient}
}

func (c *Client) CreateSession(ctx context.Context, cfg ProviderConfig, input CreateSessionInput) (Session, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case ProviderAlipay:
		return c.createAlipaySession(ctx, cfg, input)
	case ProviderWechat:
		return c.createWechatSession(ctx, cfg, input)
	default:
		return Session{}, &UnavailableError{Message: "实名供应商不支持"}
	}
}

func (c *Client) QueryResult(ctx context.Context, cfg ProviderConfig, providerApplicationID string) (Result, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case ProviderAlipay:
		return c.queryAlipayResult(ctx, cfg, providerApplicationID)
	case ProviderWechat:
		return c.queryWechatResult(ctx, cfg, providerApplicationID)
	default:
		return Result{}, &UnavailableError{Message: "实名供应商不支持"}
	}
}

func (c *Client) ParseCallback(ctx context.Context, cfg ProviderConfig, req CallbackRequest) (Callback, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case ProviderAlipay:
		return c.parseAlipayCallback(ctx, cfg, req)
	case ProviderWechat:
		return c.parseWechatCallback(ctx, cfg, req)
	default:
		return Callback{}, &InvalidCallbackError{Message: "实名供应商不支持"}
	}
}

func IsUnavailable(err error) bool {
	var target *UnavailableError
	return errors.As(err, &target)
}

func IsInvalidCallback(err error) bool {
	var target *InvalidCallbackError
	return errors.As(err, &target)
}

func (c *Client) createAlipaySession(ctx context.Context, cfg ProviderConfig, input CreateSessionInput) (Session, error) {
	apiName := "alipay.user.certify.open.initialize"
	notifyURL := strings.TrimSpace(cfg.NotifyURL)
	if notifyURL == "" {
		notifyURL = defaultProviderCallbackURL(cfg.CallbackBaseURL, ProviderAlipay)
	}
	identityParam, err := json.Marshal(map[string]string{
		"identity_type": "CERT_INFO",
		"cert_type":     "IDENTITY_CARD",
		"cert_name":     input.RealName,
		"cert_no":       input.IDNumber,
	})
	if err != nil {
		return Session{}, err
	}
	merchantConfig, err := json.Marshal(map[string]string{
		"return_url": cfg.ReturnURL,
	})
	if err != nil {
		return Session{}, err
	}
	bizContent, err := json.Marshal(map[string]string{
		"outer_order_no":  input.ApplicationNo,
		"biz_code":        "FACE",
		"identity_param":  string(identityParam),
		"merchant_config": string(merchantConfig),
	})
	if err != nil {
		return Session{}, err
	}
	client := c.newAlipayClient(cfg)
	values, err := client.URLValues(genericAliPayParam{
		apiName:      apiName,
		extJSONName:  "biz_content",
		extJSONValue: string(bizContent),
		params: map[string]string{
			"notify_url": notifyURL,
		},
	})
	if err != nil {
		return Session{}, &UnavailableError{Message: "支付宝请求签名失败"}
	}
	body, err := c.postAlipay(ctx, cfg, apiName, values)
	if err != nil {
		return Session{}, err
	}
	var response struct {
		Payload struct {
			Code      string `json:"code"`
			Msg       string `json:"msg"`
			SubCode   string `json:"sub_code"`
			SubMsg    string `json:"sub_msg"`
			CertifyID string `json:"certify_id"`
		} `json:"alipay_user_certify_open_initialize_response"`
		Sign string `json:"sign"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return Session{}, &UnavailableError{Message: "支付宝实名初始化响应解析失败"}
	}
	if response.Payload.Code != smartalipay.K_SUCCESS_CODE {
		return Session{}, &UnavailableError{Message: firstNonEmpty(response.Payload.SubMsg, response.Payload.Msg, "支付宝实名初始化失败")}
	}
	if strings.TrimSpace(response.Payload.CertifyID) == "" {
		return Session{}, &UnavailableError{Message: "支付宝实名初始化未返回认证会话"}
	}
	redirectValues, err := client.URLValues(genericAliPayParam{
		apiName:      "alipay.user.certify.open.certify",
		extJSONName:  "biz_content",
		extJSONValue: marshalJSON(map[string]string{"certify_id": response.Payload.CertifyID}),
	})
	if err != nil {
		return Session{}, &UnavailableError{Message: "支付宝跳转地址签名失败"}
	}
	expiresAt := time.Now().Add(30 * time.Minute)
	return Session{
		Provider:              ProviderAlipay,
		ProviderApplicationID: response.Payload.CertifyID,
		ActionType:            "redirect",
		RedirectURL:           buildAlipayRedirectURL(cfg, redirectValues),
		ExpiresAt:             &expiresAt,
		TraceID:               response.Payload.CertifyID,
		ResponseDigest:        sha256Hex(body),
	}, nil
}

func (c *Client) queryAlipayResult(ctx context.Context, cfg ProviderConfig, providerApplicationID string) (Result, error) {
	apiName := "alipay.user.certify.open.query"
	client := c.newAlipayClient(cfg)
	values, err := client.URLValues(genericAliPayParam{
		apiName:      apiName,
		extJSONName:  "biz_content",
		extJSONValue: marshalJSON(map[string]string{"certify_id": providerApplicationID}),
	})
	if err != nil {
		return Result{}, &UnavailableError{Message: "支付宝请求签名失败"}
	}
	body, err := c.postAlipay(ctx, cfg, apiName, values)
	if err != nil {
		return Result{}, err
	}
	var response struct {
		Payload struct {
			Code    string `json:"code"`
			Msg     string `json:"msg"`
			SubCode string `json:"sub_code"`
			SubMsg  string `json:"sub_msg"`
			Passed  string `json:"passed"`
		} `json:"alipay_user_certify_open_query_response"`
		Sign string `json:"sign"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return Result{}, &UnavailableError{Message: "支付宝实名查询响应解析失败"}
	}
	if response.Payload.Code != smartalipay.K_SUCCESS_CODE {
		return Result{}, &UnavailableError{Message: firstNonEmpty(response.Payload.SubMsg, response.Payload.Msg, "支付宝实名结果暂不可确认")}
	}
	result := Result{
		ProviderStatus: statusPending,
		FinalStatus:    statusPending,
		ResultCode:     strings.TrimSpace(response.Payload.Passed),
		ResultMessage:  "支付宝实名核验处理中",
		ResponseDigest: sha256Hex(body),
		TraceID:        providerApplicationID,
	}
	switch strings.ToUpper(strings.TrimSpace(response.Payload.Passed)) {
	case "T":
		result.ProviderStatus = statusApproved
		result.FinalStatus = statusApproved
		result.ResultCode = "PASSED"
		result.ResultMessage = "支付宝实名核验通过"
	case "F":
		result.ProviderStatus = statusRejected
		result.FinalStatus = statusRejected
		result.ResultCode = "REJECTED"
		result.ResultMessage = "支付宝实名核验未通过"
	default:
		result.ResultCode = firstNonEmpty(result.ResultCode, "PENDING")
	}
	return result, nil
}

func (c *Client) parseAlipayCallback(_ context.Context, cfg ProviderConfig, req CallbackRequest) (Callback, error) {
	values := mergeCallbackValues(req.Query, req.Form, jsonBodyValues(req.RawBody))
	certifyID := firstNonEmpty(values.Get("certify_id"), values.Get("certifyId"), values.Get("provider_application_id"))
	if strings.TrimSpace(certifyID) == "" {
		return Callback{}, &InvalidCallbackError{Message: "支付宝回调缺少认证会话"}
	}
	client := c.newAlipayClient(cfg)
	ok, err := client.VerifySign(values)
	if err != nil {
		return Callback{}, &InvalidCallbackError{Message: "支付宝回调验签失败"}
	}
	if !ok {
		return Callback{}, &InvalidCallbackError{Message: "支付宝回调签名无效"}
	}
	timestamp, err := parseAlipayCallbackTime(firstNonEmpty(values.Get("notify_time"), values.Get("gmt_create"), values.Get("timestamp")))
	if err != nil {
		return Callback{}, &InvalidCallbackError{Message: "支付宝回调时间无效"}
	}
	return Callback{
		ProviderApplicationID: certifyID,
		PayloadDigest:         sha256Hex([]byte(values.Encode())),
		ReplayKey:             certifyID,
		Timestamp:             &timestamp,
	}, nil
}

func (c *Client) createWechatSession(ctx context.Context, cfg ProviderConfig, input CreateSessionInput) (Session, error) {
	client, err := c.newTencentFaceIDClient(cfg)
	if err != nil {
		return Session{}, err
	}
	req := tcfaceid.NewDetectAuthRequest()
	req.RuleId = tencentString(cfg.RuleID)
	req.IdCard = tencentString(input.IDNumber)
	req.Name = tencentString(input.RealName)
	req.RedirectUrl = tencentString(cfg.RedirectURL)
	req.Extra = tencentString("application_no=" + input.ApplicationNo)
	resp, err := client.DetectAuthWithContext(ctx, req)
	if err != nil {
		return Session{}, wrapTencentUnavailable(err, "微信实名初始化失败")
	}
	if resp == nil || resp.Response == nil || strings.TrimSpace(valueOf(resp.Response.BizToken)) == "" || strings.TrimSpace(valueOf(resp.Response.Url)) == "" {
		return Session{}, &UnavailableError{Message: "微信实名初始化未返回核验会话"}
	}
	expiresAt := time.Now().Add(2 * time.Hour)
	body := marshalJSON(resp.Response)
	return Session{
		Provider:              ProviderWechat,
		ProviderApplicationID: valueOf(resp.Response.BizToken),
		ActionType:            "redirect",
		RedirectURL:           valueOf(resp.Response.Url),
		ExpiresAt:             &expiresAt,
		TraceID:               firstNonEmpty(valueOf(resp.Response.RequestId), valueOf(resp.Response.BizToken)),
		ResponseDigest:        sha256Hex([]byte(body)),
	}, nil
}

func (c *Client) queryWechatResult(ctx context.Context, cfg ProviderConfig, providerApplicationID string) (Result, error) {
	client, err := c.newTencentFaceIDClient(cfg)
	if err != nil {
		return Result{}, err
	}
	req := tcfaceid.NewGetDetectInfoEnhancedRequest()
	req.BizToken = tencentString(providerApplicationID)
	req.RuleId = tencentString(cfg.RuleID)
	req.InfoType = tencentString("1")
	resp, err := client.GetDetectInfoEnhancedWithContext(ctx, req)
	if err != nil {
		if rejected, ok := mapTencentRejectedResult(err, providerApplicationID); ok {
			return rejected, nil
		}
		return Result{}, wrapTencentUnavailable(err, "微信实名结果暂不可确认")
	}
	result := Result{
		ProviderStatus: statusPending,
		FinalStatus:    statusPending,
		ResultCode:     "PENDING",
		ResultMessage:  "微信实名核验处理中",
		ResponseDigest: sha256Hex([]byte(marshalJSON(resp.Response))),
		TraceID:        firstNonEmpty(valueOf(resp.Response.RequestId), providerApplicationID),
	}
	if resp == nil || resp.Response == nil || resp.Response.Text == nil {
		return result, nil
	}
	text := resp.Response.Text
	errCode := valueOfInt64(text.ErrCode)
	compareStatus := valueOfInt64(text.Comparestatus)
	liveStatus := valueOfInt64(text.LiveStatus)
	switch {
	case errCode == 0 && compareStatus == 0 && liveStatus == 0:
		result.ProviderStatus = statusApproved
		result.FinalStatus = statusApproved
		result.ResultCode = "PASSED"
		result.ResultMessage = "微信实名核验通过"
	case errCode != 0 || compareStatus != 0 || liveStatus != 0:
		result.ProviderStatus = statusRejected
		result.FinalStatus = statusRejected
		result.ResultCode = firstNonEmpty(nonZeroCode(errCode), nonZeroCode(compareStatus), nonZeroCode(liveStatus), "REJECTED")
		result.ResultMessage = firstNonEmpty(valueOf(text.ErrMsg), valueOf(text.Comparemsg), valueOf(text.LiveMsg), "微信实名核验未通过")
	}
	return result, nil
}

func (c *Client) parseWechatCallback(_ context.Context, cfg ProviderConfig, req CallbackRequest) (Callback, error) {
	values := mergeCallbackValues(req.Query, req.Form, jsonBodyValues(req.RawBody))
	providerApplicationID := firstNonEmpty(values.Get("provider_application_id"), values.Get("biz_token"), values.Get("BizToken"), values.Get("token"))
	if strings.TrimSpace(providerApplicationID) == "" {
		return Callback{}, &InvalidCallbackError{Message: "微信回调缺少核验会话"}
	}
	timestampValue := firstNonEmpty(values.Get("timestamp"), values.Get("ts"), req.Headers.Get("X-Real-Name-Timestamp"))
	timestamp, err := parseUnixOrRFC3339(timestampValue)
	if err != nil || timestamp == nil {
		return Callback{}, &InvalidCallbackError{Message: "微信回调时间无效"}
	}
	signature := firstNonEmpty(values.Get("signature"), values.Get("sign"), req.Headers.Get("X-Real-Name-Signature"))
	if !verifyWechatCallbackSignature(cfg.SecretKey, timestampValue, signature, values) {
		return Callback{}, &InvalidCallbackError{Message: "微信回调签名无效"}
	}
	return Callback{
		ProviderApplicationID: providerApplicationID,
		PayloadDigest:         sha256Hex([]byte(canonicalCallbackValues(values))),
		ReplayKey:             providerApplicationID,
		Timestamp:             timestamp,
	}, nil
}

func (c *Client) newAlipayClient(cfg ProviderConfig) *smartalipay.AliPay {
	isProduction := !strings.Contains(strings.ToLower(strings.TrimSpace(cfg.GatewayURL)), "alipaydev")
	client := smartalipay.New(strings.TrimSpace(cfg.AppID), strings.TrimSpace(cfg.AlipayPublicKey), strings.TrimSpace(cfg.AppPrivateKey), isProduction)
	client.Client = c.httpClient
	return client
}

func (c *Client) postAlipay(ctx context.Context, cfg ProviderConfig, apiName string, values url.Values) ([]byte, error) {
	requestURL := strings.TrimSpace(cfg.GatewayURL)
	if requestURL == "" {
		requestURL = buildAlipayGatewayURL(cfg)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &UnavailableError{Message: "支付宝实名服务请求失败"}
	}
	defer resp.Body.Close()
	body, err := readResponseBody(resp)
	if err != nil {
		return nil, &UnavailableError{Message: "支付宝实名服务响应读取失败"}
	}
	if resp.StatusCode >= 400 {
		return nil, &UnavailableError{Message: "支付宝实名服务返回异常状态"}
	}
	if err := verifyAlipayResponseSignature(body, apiName, cfg.AlipayPublicKey); err != nil {
		return nil, &UnavailableError{Message: "支付宝实名服务响应验签失败"}
	}
	return body, nil
}

func verifyAlipayResponseSignature(body []byte, apiName string, publicKey string) error {
	if strings.TrimSpace(publicKey) == "" {
		return errors.New("alipay public key is empty")
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	content := payload[alipayResponseNodeName(apiName)]
	if len(content) == 0 {
		content = payload["error_response"]
	}
	if len(content) == 0 {
		return errors.New("alipay response node not found")
	}
	var signature string
	if raw := payload["sign"]; len(raw) > 0 {
		if err := json.Unmarshal(raw, &signature); err != nil {
			return err
		}
	}
	if strings.TrimSpace(signature) == "" {
		return errors.New("alipay response sign not found")
	}
	return verifyAlipaySignature(content, signature, publicKey)
}

func alipayResponseNodeName(apiName string) string {
	return strings.ReplaceAll(strings.TrimSpace(apiName), ".", "_") + "_response"
}

func verifyAlipaySignature(content []byte, signature string, publicKey string) error {
	signatureBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(signature))
	if err != nil {
		return err
	}
	key, err := parseRSAPublicKey(publicKey)
	if err != nil {
		return err
	}
	hash := crypto.SHA256.New()
	_, _ = hash.Write(content)
	return rsa.VerifyPKCS1v15(key, crypto.SHA256, hash.Sum(nil), signatureBytes)
}

func parseRSAPublicKey(value string) (*rsa.PublicKey, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, errors.New("empty public key")
	}
	var der []byte
	if block, _ := pem.Decode([]byte(trimmed)); block != nil {
		der = block.Bytes
	} else {
		compact := strings.NewReplacer(" ", "", "\n", "", "\r", "", "\t", "").Replace(trimmed)
		decoded, err := base64.StdEncoding.DecodeString(compact)
		if err != nil {
			return nil, err
		}
		der = decoded
	}
	parsed, err := x509.ParsePKIXPublicKey(der)
	if err == nil {
		if key, ok := parsed.(*rsa.PublicKey); ok {
			return key, nil
		}
		return nil, errors.New("public key is not rsa")
	}
	if key, pkcs1Err := x509.ParsePKCS1PublicKey(der); pkcs1Err == nil {
		return key, nil
	}
	return nil, err
}

func buildAlipayRedirectURL(cfg ProviderConfig, values url.Values) string {
	base := strings.TrimSpace(cfg.GatewayURL)
	if base == "" {
		base = buildAlipayGatewayURL(cfg)
	}
	if strings.Contains(base, "?") {
		return base + "&" + values.Encode()
	}
	return base + "?" + values.Encode()
}

func buildAlipayGatewayURL(cfg ProviderConfig) string {
	if strings.Contains(strings.ToLower(strings.TrimSpace(cfg.GatewayURL)), "alipaydev") {
		return "https://openapi.alipaydev.com/gateway.do"
	}
	return "https://openapi.alipay.com/gateway.do"
}

func (c *Client) newTencentFaceIDClient(cfg ProviderConfig) (*tcfaceid.Client, error) {
	credential := tccommon.NewCredential(strings.TrimSpace(cfg.SecretID), strings.TrimSpace(cfg.SecretKey))
	profile := tcprofile.NewClientProfile()
	profile.HttpProfile.ReqTimeout = int(c.httpClient.Timeout / time.Second)
	if profile.HttpProfile.ReqTimeout <= 0 {
		profile.HttpProfile.ReqTimeout = 10
	}
	if endpoint := strings.TrimSpace(cfg.Endpoint); endpoint != "" {
		profile.HttpProfile.Endpoint = endpoint
	}
	client, err := tcfaceid.NewClient(credential, firstNonEmpty(cfg.Region, "ap-guangzhou"), profile)
	if err != nil {
		return nil, &UnavailableError{Message: "微信实名客户端初始化失败"}
	}
	return client, nil
}

func wrapTencentUnavailable(err error, fallback string) error {
	var sdkErr *tcerr.TencentCloudSDKError
	if errors.As(err, &sdkErr) {
		return &UnavailableError{Message: firstNonEmpty(sdkErr.GetMessage(), fallback)}
	}
	return &UnavailableError{Message: fallback}
}

func mapTencentRejectedResult(err error, providerApplicationID string) (Result, bool) {
	var sdkErr *tcerr.TencentCloudSDKError
	if !errors.As(err, &sdkErr) {
		return Result{}, false
	}
	code := strings.TrimSpace(sdkErr.GetCode())
	message := strings.TrimSpace(sdkErr.GetMessage())
	if isTerminalBizTokenError(code, message) {
		payload := marshalJSON(map[string]string{"code": code, "message": message, "request_id": sdkErr.GetRequestId()})
		return Result{
			ProviderStatus: statusRejected,
			FinalStatus:    statusRejected,
			ResultCode:     firstNonEmpty(code, "INVALID_BIZ_TOKEN"),
			ResultMessage:  firstNonEmpty(message, "微信实名核验未通过"),
			ResponseDigest: sha256Hex([]byte(payload)),
			TraceID:        firstNonEmpty(sdkErr.GetRequestId(), providerApplicationID),
		}, true
	}
	return Result{}, false
}

func isTerminalBizTokenError(code string, message string) bool {
	switch strings.TrimSpace(code) {
	case "InvalidParameterValue.BizTokenIllegal", "InvalidParameterValue.BizTokenExpired":
		return true
	case "InvalidParameter":
		return strings.Contains(strings.ToLower(message), "biztoken")
	default:
		return false
	}
}

type genericAliPayParam struct {
	apiName      string
	params       map[string]string
	extJSONName  string
	extJSONValue string
}

func (p genericAliPayParam) APIName() string { return p.apiName }

func (p genericAliPayParam) Params() map[string]string { return p.params }

func (p genericAliPayParam) ExtJSONParamName() string { return p.extJSONName }

func (p genericAliPayParam) ExtJSONParamValue() string { return p.extJSONValue }

func defaultProviderCallbackURL(base string, provider string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return ""
	}
	return base + "/" + strings.Trim(strings.ToLower(provider), "/")
}

func mergeCallbackValues(groups ...url.Values) url.Values {
	result := url.Values{}
	for _, group := range groups {
		for key, values := range group {
			for _, value := range values {
				result.Add(key, value)
			}
		}
	}
	return result
}

func canonicalCallbackValues(values url.Values) string {
	result := url.Values{}
	for key, items := range values {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		if normalizedKey == "sign" || normalizedKey == "signature" {
			continue
		}
		for _, item := range items {
			result.Add(key, item)
		}
	}
	return result.Encode()
}

func verifyWechatCallbackSignature(secret string, timestamp string, signature string, values url.Values) bool {
	secret = strings.TrimSpace(secret)
	signature = strings.TrimSpace(signature)
	timestamp = strings.TrimSpace(timestamp)
	if secret == "" || signature == "" || timestamp == "" {
		return false
	}
	signature = strings.TrimPrefix(signature, "sha256=")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("\n"))
	mac.Write([]byte(canonicalCallbackValues(values)))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.ToLower(signature)), []byte(expected))
}

func jsonBodyValues(raw []byte) url.Values {
	if len(raw) == 0 {
		return url.Values{}
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return url.Values{}
	}
	result := url.Values{}
	for key, value := range payload {
		result.Set(key, fmt.Sprint(value))
	}
	return result
}

func parseAlipayCallbackTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errors.New("empty callback time")
	}
	for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339, "2006-01-02T15:04:05Z07:00"} {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errors.New("invalid callback time")
}

func parseUnixOrRFC3339(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	if unixSeconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		parsed := time.Unix(unixSeconds, 0)
		return &parsed, nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return &parsed, nil
	}
	return nil, errors.New("invalid timestamp")
}

func readResponseBody(resp *http.Response) ([]byte, error) {
	return io.ReadAll(io.LimitReader(resp.Body, 1<<20))
}

func marshalJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(raw)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func valueOf(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func valueOfInt64(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func nonZeroCode(value int64) string {
	if value == 0 {
		return ""
	}
	return strconv.FormatInt(value, 10)
}

func tencentString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
