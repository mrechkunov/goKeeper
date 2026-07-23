package cryptodata_test

import (
	"bytes"
	"testing"

	"github.com/mrechkunov/goKeeper.git/internal/cryptodata"
)

func TestCryptoAndDecryptFile(t *testing.T) {
	// Исходные тестовые данные файла (текст, картинка или бинарник)
	originalFileData := []byte("Привет! Это секретное содержимое файла конфигурации goKeeper.")

	// 1. Тестируем успешное шифрование файла
	encryptedBytes, err := cryptodata.CryptoFile(originalFileData)
	if err != nil {
		t.Fatalf("CryptoFile вернул неожиданную ошибку: %v. Проверьте валидность keyString.", err)
	}

	if len(encryptedBytes) == 0 {
		t.Fatal("CryptoFile вернул пустой срез байт для непустого файла")
	}

	if bytes.Equal(encryptedBytes, originalFileData) {
		t.Fatal("CryptoFile вернул незашифрованные данные (исходный текст совпадает с результатом)")
	}

	// 2. Тестируем успешное дешифрование файла
	decryptedBytes, err := cryptodata.DecryptFile(encryptedBytes)
	if err != nil {
		t.Fatalf("DecryptFile вернул неожиданную ошибку при расшифровке: %v", err)
	}

	// Проверяем, что файл после расшифровки байт-в-байт совпадает с оригиналом
	if !bytes.Equal(decryptedBytes, originalFileData) {
		t.Errorf("Данные файла повреждены после цикла шифрования. Получили %q, ожидали %q", string(decryptedBytes), string(originalFileData))
	}

	// 3. Тестируем негативный сценарий: Попытка расшифровать поврежденный файл
	t.Run("Ошибка при повреждении байт файла", func(t *testing.T) {
		if len(encryptedBytes) <= 5 {
			t.Skip("Зашифрованный файл слишком короткий для проведения теста повреждения")
		}

		// Создаем копию зашифрованных данных и умышленно ломаем один байт в середине файла
		corruptedBytes := make([]byte, len(encryptedBytes))
		copy(corruptedBytes, encryptedBytes)
		corruptedBytes[len(corruptedBytes)/2] ^= 0xFF // Инвертируем байт

		// AES-GCM должен обнаружить нарушение целостности (tag mismatch)
		_, err := cryptodata.DecryptFile(corruptedBytes)
		if err == nil {
			t.Error("DecryptFile успешно расшифровал поврежденные байты, ожидалась ошибка проверки целостности (cipher: message authentication failed)")
		}
	})
}
