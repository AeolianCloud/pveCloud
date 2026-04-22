package auth

import (
	"context"
	"net/http"
	"strings"

	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
)

type contextKey string

const claimsContextKey contextKey = "auth_claims"

func Middleware(signer *JWTSigner, subjectType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				httpx.WriteError(w, errorsx.ErrUnauthorized)
				return
			}

			claims, err := signer.Parse(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				httpx.WriteError(w, errorsx.ErrUnauthorized)
				return
			}
			if claims.SubjectType != subjectType {
				httpx.WriteError(w, errorsx.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsContextKey, claims)))
		})
	}
}

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(Claims)
	return claims, ok
}

func MustUserID(ctx context.Context) uint64 {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return 0
	}
	return claims.SubjectID
}
