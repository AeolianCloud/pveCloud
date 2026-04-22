package errorsx

import "net/http"

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *Error) Error() string {
	return e.Message
}

var (
	ErrInternal     = &Error{Code: "internal_error", Message: "internal error", Status: http.StatusInternalServerError}
	ErrBadRequest   = &Error{Code: "bad_request", Message: "bad request", Status: http.StatusBadRequest}
	ErrUnauthorized = &Error{Code: "unauthorized", Message: "unauthorized", Status: http.StatusUnauthorized}
	ErrForbidden    = &Error{Code: "forbidden", Message: "forbidden", Status: http.StatusForbidden}
	ErrConflict     = &Error{Code: "conflict", Message: "conflict", Status: http.StatusConflict}
)

func Status(err error) int {
	if appErr, ok := err.(*Error); ok {
		return appErr.Status
	}
	return http.StatusInternalServerError
}
