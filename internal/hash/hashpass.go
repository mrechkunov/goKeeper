package hash

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword генерирует bcrypt-хэш для пароля
func HashPassword(password string) (string, error) {
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
