package requestcontext

import "context"

type contextKey struct{}

/**
 * RequestContext 表示请求级上下文。
 */
type RequestContext struct {
	AdminID          *uint64
	AdminUsername    string
	AdminDisplayName string
	SessionID        string
	RequestID        string
	RequestMethod    string
	RequestPath      string
	IP               string
	UserAgent        string
}

/**
 * WithRequestContext 合并并写入请求上下文。
 */
func WithRequestContext(ctx context.Context, next RequestContext) context.Context {
	current := RequestContextFrom(ctx)
	if next.AdminID != nil {
		current.AdminID = next.AdminID
	}
	if next.AdminUsername != "" {
		current.AdminUsername = next.AdminUsername
	}
	if next.AdminDisplayName != "" {
		current.AdminDisplayName = next.AdminDisplayName
	}
	if next.SessionID != "" {
		current.SessionID = next.SessionID
	}
	if next.RequestID != "" {
		current.RequestID = next.RequestID
	}
	if next.RequestMethod != "" {
		current.RequestMethod = next.RequestMethod
	}
	if next.RequestPath != "" {
		current.RequestPath = next.RequestPath
	}
	if next.IP != "" {
		current.IP = next.IP
	}
	if next.UserAgent != "" {
		current.UserAgent = next.UserAgent
	}
	return context.WithValue(ctx, contextKey{}, current)
}

/**
 * RequestContextFrom 从请求上下文读取请求上下文信息。
 */
func RequestContextFrom(ctx context.Context) RequestContext {
	if ctx == nil {
		return RequestContext{}
	}
	value, ok := ctx.Value(contextKey{}).(RequestContext)
	if !ok {
		return RequestContext{}
	}
	return value
}
