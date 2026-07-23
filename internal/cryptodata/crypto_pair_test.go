package cryptodata_test

import (
	"encoding/base64"
	"testing"

	"github.com/mrechkunov/goKeeper.git/internal/cryptodata"
)

func TestCryptoAndDecryptPair(t *testing.T) {
	// Тестовые данные аккаунта
	testLogin := "my_awesome_user"
	testPassword := "super_secret_P@ssw0rd!"

	// 1. Тестируем успешное шифрование пары в Base64
	validB64, err := cryptodata.CryptoPair(testLogin, testPassword)
	if err != nil {
		t.Fatalf("CryptoPair вернул неожиданную ошибку: %v. Проверьте keyString в приложении.", err)
	}

	if validB64 == "" {
		t.Fatal("CryptoPair вернул пустую строку для валидных данных")
	}

	// 2. Тестируем успешное дешифрование полученного токена
	gotLogin, gotPass, err := cryptodata.DecryptPair(validB64)
	if err != nil {
		t.Fatalf("DecryptPair вернул ошибку при расшифровке валидных данных: %v", err)
	}

	if gotLogin != testLogin {
		t.Errorf("DecryptPair() login = %q, want %q", gotLogin, testLogin)
	}

	if gotPass != testPassword {
		t.Errorf("DecryptPair() password = %q, want %q", gotPass, testPassword)
	}

	// 3. Тестируем негативные сценарии
	t.Run("Негативные кейсы", func(t *testing.T) {
		tests := []struct {
			name     string
			inputB64 string
			wantErr  bool
		}{
			{
				name:     "Ошибка: Сломанный Base64",
				inputB64: "not-a-valid-base64-string!!!",
				wantErr:  true,
			},
			{
				name:     "Ошибка: Битые зашифрованные байты внутри Base64",
				inputB64: base64.StdEncoding.EncodeToString([]byte("random-corrupted-crypt-bytes-here")),
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, _, err := cryptodata.DecryptPair(tt.inputB64)
				if (err != nil) != tt.wantErr {
					t.Errorf("DecryptPair() [%s] error = %v, wantErr %v", tt.name, err, tt.wantErr)
				}
			})
		}
	})
}
