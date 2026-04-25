package password

import "golang.org/x/crypto/bcrypt"

/**
 * Hash 使用 bcrypt 生成密码哈希。
 *
 * @param plain 明文密码
 * @return string bcrypt 哈希
 * @return error 哈希失败原因
 */
func Hash(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

/**
 * Verify 校验明文密码是否匹配 bcrypt 哈希。
 *
 * @param hash bcrypt 哈希
 * @param plain 明文密码
 * @return bool 密码是否匹配
 */
func Verify(hash string, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
