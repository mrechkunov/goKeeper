package hash

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword генерирует bcrypt-хэш для пароля
func HashPassword(password string) (string, error) {
	// Второй аргумент — cost (сложность). Значение 12 является оптимальным
	// балансом между безопасностью и производительностью.
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

// CheckPasswordHash сравнивает введенный пароль с сохраненным хэшем
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// func main() {
// 	password := "mySuperSecret123"

// 	// Хэширование
// 	hash, _ := HashPassword(password)
// 	fmt.Println("Сохраняемый хэш:", hash)

// 	// Проверка
// 	isMatch := CheckPasswordHash("mySuperSecret123", hash)
// 	fmt.Printf("Пароль верный: %v\n", isMatch) // true
// }

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
		return nil, err
	}

	return plaintext, nil
}

// func main() {
// 	// 1. Секретный ключ (должен быть 16, 24 или 32 байта для AES-128/192/256)
// 	// Для реальных приложений используйте менеджер секретов или генератор KDF!
// 	keyString := "supersecretkeywhichis32byteslong!!"
// 	key := []byte(keyString)

// 	plaintext := []byte("Секретное сообщение для шифрования")

// 	// 2. Шифрование
// 	ciphertext, err := encrypt(plaintext, key)
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Printf("Зашифрованные данные (hex): %x\n", ciphertext)

// 	// 3. Расшифрование
// 	origText, err := decrypt(ciphertext, key)
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Printf("Расшифрованные данные: %s\n", string(origText))
// }
