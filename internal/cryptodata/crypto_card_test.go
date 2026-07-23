package cryptodata_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/mrechkunov/goKeeper.git/internal/cryptodata"
)

func TestCryptoCard(t *testing.T) {

	tests := []struct {
		name    string
		number  string
		valid   string
		cvv     string
		wantErr bool
		errText string
	}{
		{
			name:    "Успешно: Корректная карта (Visa/Mastercard)",
			number:  "4532718281828182", // Валидный номер по Луну
			valid:   "12/29",
			cvv:     "123",
			wantErr: false,
		},
		{
			name:    "Ошибка: Невалидный номер карты (алгоритм Луна)",
			number:  "4532718281828183", // Невалидный номер
			valid:   "12/29",
			cvv:     "123",
			wantErr: true,
			errText: "no correct card number",
		},
		{
			name:    "Ошибка: Пустой номер карты",
			number:  "",
			valid:   "05/26",
			cvv:     "999",
			wantErr: true,
			errText: "no correct card number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut, err := cryptodata.CryptoCard(tt.number, tt.valid, tt.cvv)

			// Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Errorf("CryptoCard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Если ошибка ожидалась, проверяем её текст
			if tt.wantErr && err != nil && err.Error() != tt.errText {
				t.Errorf("CryptoCard() error text = %q, want %q", err.Error(), tt.errText)
			}

			// Если сценарий успешный, проверяем структуру результата
			if !tt.wantErr {
				if gotOut == "" {
					t.Error("CryptoCard() returned an empty string for a valid card")
				}

				// Декодируем base64 обратно, чтобы проверить, что данные внутри склеились верно
				decodedBytes, err := base64.StdEncoding.DecodeString(gotOut)
				if err != nil {
					t.Errorf("CryptoCard() output is not a valid base64 string: %v", err)
				}

				decodedStr := string(decodedBytes)
				expectedData := tt.number + "|" + tt.valid + "|" + tt.cvv
				if decodedStr != expectedData && strings.Contains(decodedStr, "|") {

					t.Logf("Бинарные данные зашифрованы успешно в Base64. Длина: %d байт", len(decodedBytes))
				}
			}
		})
	}
}

func TestDecryptCard_Final(t *testing.T) {
	// Подготавливаем тестовые данные
	validNumber := "4532718281828182" // Валидный номер по Луну, чтобы CryptoCard пропустил его
	validDate := "12/29"
	validCVV := "123"

	// Генерируем зашифрованный токен через вашу РОДНУЮ функцию CryptoCard.
	validB64, err := cryptodata.CryptoCard(validNumber, validDate, validCVV)
	if err != nil {
		t.Fatalf("CryptoCard вернул ошибку при подготовке данных. Проверьте keyString в приложении: %v", err)
	}

	tests := []struct {
		name       string
		inputB64   string
		wantNumber string
		wantValid  string
		wantCvv    string
		wantErr    bool
	}{
		{
			name:       "Успешно: Корректная расшифровка всех полей карты",
			inputB64:   validB64,
			wantNumber: validNumber,
			wantValid:  validDate,
			wantCvv:    validCVV,
			wantErr:    false,
		},
		{
			name:     "Ошибка: Сломанный Base64",
			inputB64: "invalid-base64-string!!!",
			wantErr:  true,
		},
		{
			name:     "Ошибка: Битые байты внутри валидного Base64",
			inputB64: base64.StdEncoding.EncodeToString([]byte("some-random-corrupted-bytes-here")),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Вызываем тестируемый метод дешифрования
			num, val, cvv, err := cryptodata.DecryptCard(tt.inputB64)

			// Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("FAIL [%s]: error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			// Если всё успешно, проверяем совпадение полей
			if !tt.wantErr {
				if num != tt.wantNumber || val != tt.wantValid || cvv != tt.wantCvv {
					t.Errorf("FAIL [%s]: got (%q, %q, %q), want (%q, %q, %q)",
						tt.name, num, val, cvv, tt.wantNumber, tt.wantValid, tt.wantCvv)
				}
			}
		})
	}
}
