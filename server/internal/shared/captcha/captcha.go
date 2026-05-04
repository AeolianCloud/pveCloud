package captcha

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

const (
	DefaultCodeLength = 4
	DefaultTTLSeconds = 120
)

// NewID 生成验证码或会话 ID。
func NewID(prefix string) (string, error) {
	value, err := RandomHex(16)
	if err != nil {
		return "", err
	}
	return prefix + value, nil
}

// RandomHex 生成指定字节长度的随机十六进制字符串。
func RandomHex(byteLength int) (string, error) {
	var bytes [16]byte
	buffer := bytes[:0]
	if byteLength > len(bytes) {
		buffer = make([]byte, byteLength)
	} else {
		buffer = bytes[:byteLength]
	}
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

// RandomCode 生成指定位数的验证码内容。
func RandomCode(length int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var builder strings.Builder
	builder.Grow(length)
	max := big.NewInt(int64(len(alphabet)))
	for i := 0; i < length; i++ {
		index, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		builder.WriteByte(alphabet[index.Int64()])
	}
	return builder.String(), nil
}

// ImageDataURL 把验证码内容渲染为 data URL。
func ImageDataURL(code string) string {
	svg := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="132" height="44" viewBox="0 0 132 44"><rect width="132" height="44" rx="8" fill="#f4f7ff"/><path d="M8 32 C30 10, 52 40, 76 18 S110 36, 124 12" fill="none" stroke="#9db6ff" stroke-width="2" opacity=".65"/><path d="M15 12 L118 35" stroke="#d2dcff" stroke-width="2" opacity=".75"/><text x="66" y="29" text-anchor="middle" font-family="Consolas, Menlo, monospace" font-size="24" font-weight="700" letter-spacing="5" fill="#2458d9">%s</text></svg>`,
		code,
	)
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}

// HashText 对验证码答案或限流标识做统一哈希。
func HashText(value string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:])
}
