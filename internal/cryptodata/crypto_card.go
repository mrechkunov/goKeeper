package cryptodata

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/mrechkunov/goKeeper.git/internal/auth"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
)

// CryptoCard card data to base64 string
func CryptoCard(number, valid, cvv string) (out string, err error) {
	if !auth.ValidLuhnCardNumber(number) {
		return out, errors.New("no correct card number")
	}
	data := number + "|" + valid + "|" + cvv
	cryptodata, err := Encrypt([]byte(data), []byte(keyString))
	if err != nil {
		logger.Log.Errorln(err)
		return
	}
	return base64.StdEncoding.EncodeToString(cryptodata), nil
}

// DecryptCard decrypt card data from base64 to string
func DecryptCard(in string) (number, valid, cvv string, err error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		logger.Log.Warnln("decoding error:", err)
		return
	}
	decryptdata, err := Decrypt(decodedBytes, []byte(keyString))
	if err != nil {
		logger.Log.Errorln(err)
		return
	}
	decryptdatastring := string(decryptdata)
	res := strings.SplitN(decryptdatastring, "|", 3)
	return res[0], res[1], res[3], nil
}
