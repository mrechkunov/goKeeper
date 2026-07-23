package cryptodata

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/mrechkunov/goKeeper.git/internal/auth"
	"github.com/mrechkunov/goKeeper.git/internal/config"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
)

// CryptoCard card data to base64 string
func CryptoCard(number, valid, cvv string) (out string, err error) {
	if !auth.ValidLuhnCardNumber(number) {
		return out, errors.New("no correct card number")
	}
	data := number + "|" + valid + "|" + cvv
	cryptodata, err := Encrypt([]byte(data), []byte(config.KeyString))
	if err != nil {
		logger.Log.Errorln(err)
		return
	}
	return base64.StdEncoding.EncodeToString(cryptodata), nil
}

// DecryptCard decrypt card data from base64 to string
func DecryptCard(in string) (string, string, string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		logger.Log.Warnln("decoding error:", err)
		return "", "", "", err
	}

	// Передаем глобальный ключ keyString
	decryptdata, err := Decrypt(decodedBytes, []byte(config.KeyString))
	if err != nil {
		logger.Log.Errorln("decryption error:", err)
		return "", "", "", err // Явно возвращаем ошибку криптографии
	}

	decryptdatastring := string(decryptdata)
	res := strings.SplitN(decryptdatastring, "|", 3)

	// Защита от некорректно расшифрованного формата данных
	if len(res) != 3 {
		return "", "", "", fmt.Errorf("invalid decrypted data format")
	}

	// Безопасный возврат индексов без паники out of bounds
	return res[0], res[1], res[2], nil
}
