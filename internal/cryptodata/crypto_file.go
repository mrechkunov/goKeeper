package cryptodata

import (
	"github.com/mrechkunov/goKeeper.git/internal/logger"
)

// CryptoFile file bytes crypt with key
func CryptoFile(data []byte) (cryptodata []byte, err error) {

	cryptodata, err = Encrypt([]byte(data), []byte(keyString))
	if err != nil {
		logger.Log.Errorln(err)
		return cryptodata, err
	}
	return cryptodata, nil
}

// DecryptFile decrypt file bytes with key
func DecryptFile(cryptodata []byte) (data []byte, err error) {
	decryptdata, err := Decrypt(cryptodata, []byte(keyString))
	if err != nil {
		logger.Log.Errorln(err)
		return decryptdata, err
	}
	return decryptdata, nil
}
