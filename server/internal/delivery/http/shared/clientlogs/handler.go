package clientlogs

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/middleware"
	webmiddleware "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/middleware"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/cache"
	sharedcaptcha "github.com/AeolianCloud/pveCloud/server/internal/shared/captcha"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
	weblogging "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/logging"
)

const (
	clientErrorRateLimit  = int64(60)
	clientErrorRateWindow = time.Minute
)

type Handler struct {
	sourceApp string
	redis     *cache.Redis
	recorder  *weblogging.Recorder
}

func NewHandler(sourceApp string, redis *cache.Redis, recorder *weblogging.Recorder) *Handler {
	return &Handler{sourceApp: sourceApp, redis: redis, recorder: recorder}
}

func (h *Handler) Create(c *gin.Context) {
	if err := h.ensureAllowed(c.Request.Context(), c.ClientIP()); err != nil {
		response.Error(c, err)
		return
	}
	var req webdto.ClientErrorLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}

	input := weblogging.FrontendErrorInput{
		SourceApp:    h.sourceApp,
		RequestID:    req.RequestID,
		PagePath:     req.PagePath,
		ErrorType:    req.ErrorType,
		Message:      req.Message,
		Stack:        req.Stack,
		APIPath:      req.APIPath,
		HTTPStatus:   req.HTTPStatus,
		BusinessCode: req.BusinessCode,
		Browser:      req.Browser,
		OS:           req.OS,
		AppVersion:   req.AppVersion,
	}
	if h.sourceApp == "admin" {
		if adminID, ok := middleware.CurrentAdminID(c); ok {
			input.AdminID = &adminID
		}
	} else {
		if userID, ok := webmiddleware.CurrentUserID(c); ok {
			input.UserID = &userID
		}
	}
	if err := h.recorder.FrontendError(c.Request.Context(), nil, input); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{})
}

func (h *Handler) ensureAllowed(ctx context.Context, ip string) error {
	if h.redis == nil {
		return nil
	}
	segment := strings.TrimSpace(h.sourceApp)
	if segment == "" {
		segment = "unknown"
	}
	key := h.redis.Key("client_logs", "errors", segment, sharedcaptcha.HashText(ip))
	count, err := h.redis.Client().Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		if err := h.redis.Client().Expire(ctx, key, clientErrorRateWindow).Err(); err != nil {
			return err
		}
	}
	if count > clientErrorRateLimit {
		return apperrors.ErrTooManyRequests.WithMessage("错误日志上报过于频繁，请稍后再试")
	}
	return nil
}
