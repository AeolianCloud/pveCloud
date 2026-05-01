package audit

import "context"

type auditContextKey struct{}

/**
 * RequestContext 表示普通操作日志需要的请求级上下文。
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
 * WithRequestContext 合并并写入审计请求上下文。
 *
 * @param ctx 请求上下文
 * @param next 新增或覆盖的审计上下文
 * @return context.Context 写入后的请求上下文
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
	return context.WithValue(ctx, auditContextKey{}, current)
}

/**
 * RequestContextFrom 从请求上下文读取审计上下文。
 *
 * @param ctx 请求上下文
 * @return RequestContext 审计请求上下文
 */
func RequestContextFrom(ctx context.Context) RequestContext {
	if ctx == nil {
		return RequestContext{}
	}
	value, ok := ctx.Value(auditContextKey{}).(RequestContext)
	if !ok {
		return RequestContext{}
	}
	return value
}
