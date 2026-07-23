package cryptodata

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/mrechkunov/goKeeper.git/internal/config"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
)

// Encrypt bytes AES with keyString
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Генерируем случайный Nonce (число одноразового использования)
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Шифруем данные и прикрепляем nonce к началу массива
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt bytes AES-GCM with key
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("зашифрованный текст слишком короткий")
	}

	// Извлекаем nonce из начала зашифрованного сообщения
	nonce := ciphertext[:nonceSize]
	ciphertextBytes := ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return nil, err // Возвращает ошибку, если ключ не совпал (cipher: message authentication failed)
	}
	return plaintext, nil
}

// CryptoPair login password to base64 string
func CryptoPair(login, pass string) (string, error) {
	pair := login + "|" + pass
	cryptopair, err := Encrypt([]byte(pair), []byte(config.KeyString))
	if err != nil {
		logger.Log.Errorln("encryption error:", err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cryptopair), nil
}

// DecryptPair login password from base64 string
func DecryptPair(in string) (string, string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		logger.Log.Warnln("decoding error:", err)
		return "", "", err
	}

	decryptpair, err := Decrypt(decodedBytes, []byte(config.KeyString))
	if err != nil {
		logger.Log.Errorln("decryption error:", err)
		return "", "", err
	}

	decryptpairstring := string(decryptpair)
	res := strings.SplitN(decryptpairstring, "|", 2)

	// Защита от криптографического мусора (неверного ключа)
	if len(res) != 2 {
		return "", "", fmt.Errorf("invalid decrypted pair format")
	}

	return res[0], res[1], nil
}
