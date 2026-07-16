package hash

import (
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
