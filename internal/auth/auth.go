package auth

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"golang.org/x/crypto/bcrypt"
)

const secretKey = "secret key"

// generate and sign token
func GenerateToken(uLogin string) (string, error) {
	// Создаем claims (данные токена)
	claims := jwt.MapClaims{
		"username": uLogin,
		"exp":      time.Now().Add(time.Hour * 2).Unix(), // Срок действия 2 часа
		"iat":      time.Now().Unix(),
	}
	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Подписываем токен
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		logger.Log.Errorln("error while sign token", err)
		return "", err
	}
	return tokenString, nil
}

// ValidateToken validate token signature
func ValidateToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil || !token.Valid {
		logger.Log.Infoln(tokenString, err)
		return err
	}
	return nil
}

// GetLoginByToken returns login by token if token is valid
func GetLoginByToken(tokenString string) (login string, err error) {
	err = ValidateToken(tokenString)
	if err != nil {
		return
	}
	// Парсим токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		logger.Log.Warnln("Error while parsing token", err)
		return
	}

	// Извлекаем claims и получаем логин
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		login = claims["username"].(string)
	} else {
		err = errors.New("Error while claims get: token is expired")
	}
	return login, err
}

// HashPassword Функция для хеширования пароля
func HashPassword(password string) (string, error) {
	// Cost (число итераций) определяет время вычисления.
	// Значение 14 является хорошим балансом между безопасностью и скоростью.
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash Функция для проверки, совпадает ли введенный пароль с хешем
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// проверяет номер карты по алгоритму Луна
func ValidLuhnCardNumber(num *int64) bool {
	number := strconv.FormatInt(*num, 10)
	// убираем все пробелы в строке
	number = strings.ReplaceAll(number, " ", "")
	// проверяем что больше 2-х цифр
	if len(number) <= 1 {
		return false
	}
	sum := 0
	// проходим слева направо
	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false // если не цифра
		}
		// Удваиваем каждую вторую цифру начиная с самой правой -1
		if (len(number)-1-i)%2 == 1 {
			digit *= 2
			if digit > 9 {
				// если удвоение двухзначное вычитаем 9
				digit -= 9
			}
		}
		sum += digit
	}
	// номер заказа валиден если сумма делится без остатка на 10
	return sum%10 == 0
}
