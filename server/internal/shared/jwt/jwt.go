package jwt

import (
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

/**
 * Claims 定义用户端和管理端共用的 JWT 声明结构。
 */
type Claims struct {
	TokenType       string   `json:"token_type"`
	UserID          uint64   `json:"user_id,omitempty"`
	AdminID         uint64   `json:"admin_id,omitempty"`
	RoleIDs         []uint64 `json:"role_ids,omitempty"`
	PermissionCodes []string `json:"permission_codes,omitempty"`
	gojwt.RegisteredClaims
}

/**
 * Sign 使用 HS256 签发 JWT。
 *
 * @param claims JWT 声明
 * @param secret 签名密钥
 * @return string JWT 字符串
 * @return error 签发失败原因
 */
func Sign(claims Claims, secret string) (string, error) {
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

/**
 * Parse 解析并校验 JWT。
 *
 * @param tokenString JWT 字符串
 * @param secret 签名密钥
 * @return *Claims JWT 声明
 * @return error 解析或校验失败原因
 */
func Parse(tokenString string, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := gojwt.ParseWithClaims(tokenString, claims, func(token *gojwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, gojwt.WithValidMethods([]string{gojwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	// 明确限制签名算法，避免 token 被替换成非预期算法后仍通过解析。
	if !token.Valid {
		return nil, gojwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

/**
 * NewRegisteredClaims 创建带签发方和过期时间的标准声明。
 *
 * @param issuer 签发方
 * @param ttl 有效时长
 * @return jwt.RegisteredClaims 标准声明
 */
func NewRegisteredClaims(issuer string, ttl time.Duration) gojwt.RegisteredClaims {
	now := time.Now()
	return gojwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  gojwt.NewNumericDate(now),
		ExpiresAt: gojwt.NewNumericDate(now.Add(ttl)),
	}
}
