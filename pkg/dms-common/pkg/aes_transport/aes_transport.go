package aes_transport

import (
	"fmt"

	pkgAes "github.com/actiontech/dms/pkg/dms-common/pkg/aes"
)

// DecryptSecretPassword decrypts Base64(AES-256-CBC(password)) with the
// compile-time SecretKey. Always uses NewEncryptor(SecretKey); never follows
// ResetAesSecretKey / std runtime key.
func DecryptSecretPassword(secretPassword string) (string, error) {
	if secretPassword == "" {
		return "", fmt.Errorf("口令密文不完整")
	}
	enc := pkgAes.NewEncryptor(pkgAes.SecretKey)
	plain, err := enc.AesDecrypt(secretPassword)
	if err != nil {
		return "", fmt.Errorf("口令解密失败")
	}
	return plain, nil
}

// EncryptForTest encrypts plaintext with the fixed SecretKey (self-test / curl helpers).
func EncryptForTest(plaintext string) (string, error) {
	enc := pkgAes.NewEncryptor(pkgAes.SecretKey)
	return enc.AesEncrypt(plaintext)
}
