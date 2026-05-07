package audit

import (
	"context"

	"github.com/AeolianCloud/pveCloud/server/internal/shared/requestcontext"
)

type RequestContext = requestcontext.RequestContext

func WithRequestContext(ctx context.Context, next RequestContext) context.Context {
	return requestcontext.WithRequestContext(ctx, next)
}

func RequestContextFrom(ctx context.Context) RequestContext {
	return requestcontext.RequestContextFrom(ctx)
}
