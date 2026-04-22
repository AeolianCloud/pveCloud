package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

var errInvalidToken = errors.New("invalid token")

type Claims struct {
	SubjectID   uint64 `json:"subject_id"`
	SubjectType string `json:"subject_type"`
}

type JWTSigner struct {
	secret []byte
}

func NewJWTSigner(secret string) *JWTSigner {
	return &JWTSigner{secret: []byte(secret)}
}

func (s *JWTSigner) Issue(claims Claims) (string, error) {
	headerPart, err := encodeJWTPart(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", err
	}

	payloadPart, err := encodeJWTPart(claims)
	if err != nil {
		return "", err
	}

	unsigned := headerPart + "." + payloadPart
	signature := s.sign(unsigned)
	return unsigned + "." + signature, nil
}

func (s *JWTSigner) Parse(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, errInvalidToken
	}

	unsigned := parts[0] + "." + parts[1]
	expectedSignature := s.sign(unsigned)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return Claims{}, errInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, errInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return Claims{}, errInvalidToken
	}
	if claims.SubjectID == 0 || claims.SubjectType == "" {
		return Claims{}, errInvalidToken
	}

	return claims, nil
}

func (s *JWTSigner) sign(unsigned string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func encodeJWTPart(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}
