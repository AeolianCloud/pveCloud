package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/AeolianCloud/pveCloud/server/internal/pkg/errors"
)

type Envelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

func Error(c *gin.Context, err error) {
	appErr := apperrors.From(err)
	if appErr == nil {
		appErr = apperrors.ErrInternal
	}

	c.JSON(appErr.HTTPStatus, Envelope{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    nil,
	})
}
