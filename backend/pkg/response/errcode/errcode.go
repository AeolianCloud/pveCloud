package errcode

// Code 业务错误码类型
type Code int

const (
	// 通用
	OK            Code = 0
	ServerError   Code = 10000
	InvalidParams Code = 10001
	NotFound      Code = 10002
	TooManyReqs   Code = 10003

	// 认证 / 权限
	Unauthorized  Code = 20001
	TokenExpired  Code = 20002
	TokenInvalid  Code = 20003
	Forbidden     Code = 20004

	// 用户
	UserNotFound      Code = 30001
	UserAlreadyExists Code = 30002
	PasswordWrong     Code = 30003

	// 订单
	OrderNotFound   Code = 40001
	OrderStatusErr  Code = 40002
)

// messages 错误码对应的中文描述
var messages = map[Code]string{
	OK:            "成功",
	ServerError:   "服务器内部错误",
	InvalidParams: "请求参数错误",
	NotFound:      "资源不存在",
	TooManyReqs:   "请求过于频繁，请稍后再试",

	Unauthorized:  "未授权，请先登录",
	TokenExpired:  "登录已过期，请重新登录",
	TokenInvalid:  "无效的令牌",
	Forbidden:     "无权限访问",

	UserNotFound:      "用户不存在",
	UserAlreadyExists: "用户已存在",
	PasswordWrong:     "密码错误",

	OrderNotFound:  "订单不存在",
	OrderStatusErr: "订单状态异常",
}

// Msg 返回错误码对应的中文描述，未定义则返回默认提示
func (c Code) Msg() string {
	if msg, ok := messages[c]; ok {
		return msg
	}
	return "未知错误"
}

// Int 返回 int 值，方便 JSON 序列化
func (c Code) Int() int {
	return int(c)
}
