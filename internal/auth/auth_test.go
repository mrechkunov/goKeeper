package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mrechkunov/goKeeper.git/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Unit-тест генерации
func TestGenerateToken(t *testing.T) {
	// Подготовка данных
	userID := "userLogin"
	secretKey := []byte("secret key")

	// Действие: вызываем функцию
	token, err := auth.GenerateToken(userID)

	// Проверки (asserts)
	require.NoError(t, err, "Генерация токена не должна возвращать ошибку")
	assert.NotEmpty(t, token, "Токен не должен быть пустым")

	// Доп. проверка: парсим токен, чтобы убедиться в его корректности
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	require.NoError(t, err, "Токен должен успешно парситься")
	assert.True(t, parsedToken.Valid, "Токен должен быть валидным")
}

func TestValidateToken(t *testing.T) {
	correctSecret := []byte("secret key")
	wrongSecret := []byte("wrong_secret_key")

	// Вспомогательная функция для генерации токенов с нужными параметрами
	createTestToken := func(secret []byte, exp time.Time) string {
		claims := jwt.MapClaims{
			"username": "test_user",
			"exp":      exp.Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString(secret)
		return tokenString
	}

	// Генерируем токены для разных сценариев
	validToken := createTestToken(correctSecret, time.Now().Add(time.Hour))
	expiredToken := createTestToken(correctSecret, time.Now().Add(-time.Hour))
	wrongSignedToken := createTestToken(wrongSecret, time.Now().Add(time.Hour))

	// Структура тест-кейса
	tests := []struct {
		name        string
		tokenString string
		wantErr     bool
	}{
		{
			name:        "Успешная валидация корректного токена",
			tokenString: validToken,
			wantErr:     false,
		},
		{
			name:        "Ошибка: Токен просрочен",
			tokenString: expiredToken,
			wantErr:     true,
		},
		{
			name:        "Ошибка: Токен подписан другим ключом",
			tokenString: wrongSignedToken,
			wantErr:     true,
		},
		{
			name:        "Ошибка: Передана некорректная строка вместо токена",
			tokenString: "not.a.valid.jwt.string",
			wantErr:     true,
		},
		{
			name:        "Ошибка: Пустая строка",
			tokenString: "",
			wantErr:     true,
		},
	}

	// Запуск тестов в цикле
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.ValidateToken(tt.tokenString)

			// Проверяем наличие ошибки в соответствии с ожиданием
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetLoginByToken(t *testing.T) {
	correctSecret := []byte("secret key")
	wrongSecret := []byte("wrong_secret_key")
	testUsername := "secure_user"

	// Вспомогательная функция для генерации токенов с кастомными claims
	createTestToken := func(secret []byte, claims jwt.MapClaims) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString(secret)
		return tokenString
	}

	// 1. Стандартный валидный токен
	validToken := createTestToken(correctSecret, jwt.MapClaims{
		"username": testUsername,
		"exp":      time.Now().Add(time.Hour).Unix(),
	})

	// 2. Просроченный токен
	expiredToken := createTestToken(correctSecret, jwt.MapClaims{
		"username": testUsername,
		"exp":      time.Now().Add(-time.Hour).Unix(),
	})

	// 3. Токен с неверной подписью
	wrongSignedToken := createTestToken(wrongSecret, jwt.MapClaims{
		"username": testUsername,
		"exp":      time.Now().Add(time.Hour).Unix(),
	})

	// 4. Токен без поля username в claims
	tokenWithoutUsername := createTestToken(correctSecret, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	// 5. Токен, где username — это число, а не строка (вызовет ошибку приведения типов)
	tokenWithInvalidUsernameType := createTestToken(correctSecret, jwt.MapClaims{
		"username": 12345, // int вместо string
		"exp":      time.Now().Add(time.Hour).Unix(),
	})

	// Структура тест-кейса
	tests := []struct {
		name          string
		tokenString   string
		expectedLogin string
		wantErr       bool
	}{
		{
			name:          "Успешное получение логина",
			tokenString:   validToken,
			expectedLogin: testUsername,
			wantErr:       false,
		},
		{
			name:          "Ошибка: Токен просрочен",
			tokenString:   expiredToken,
			expectedLogin: "",
			wantErr:       true,
		},
		{
			name:          "Ошибка: Неверная подпись токена",
			tokenString:   wrongSignedToken,
			expectedLogin: "",
			wantErr:       true,
		},
		{
			name:          "Ошибка: Передана пустая строка",
			tokenString:   "",
			expectedLogin: "",
			wantErr:       true,
		},
		{
			name:          "Ошибка: В токене отсутствует claim 'username'",
			tokenString:   tokenWithoutUsername,
			expectedLogin: "",
			wantErr:       true,
		},
		{
			name:          "Ошибка: Поле 'username' имеет неверный тип (не string)",
			tokenString:   tokenWithInvalidUsernameType,
			expectedLogin: "",
			wantErr:       true,
		},
	}

	// Запуск тестов
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Перед запуском кейса с неверным типом username временно перехватим панику,
			// если функция упадет из-за некорректного приведения типов.
			defer func() {
				if r := recover(); r != nil {
					if tt.name == "Ошибка: Поле 'username' имеет неверный тип (не string)" {
						t.Errorf("Функция запаниковала на неверном типе данных (нужна безопасная проверка типа): %v", r)
					} else {
						t.Fatalf("Неожиданная паника: %v", r)
					}
				}
			}()

			login, err := auth.GetLoginByToken(tt.tokenString)

			// Проверка на наличие ошибки
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLoginByToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Проверка возвращаемого логина
			if login != tt.expectedLogin {
				t.Errorf("GetLoginByToken() login = %q, expected %q", login, tt.expectedLogin)
			}
		})
	}
}

func TestValidLuhnCardNumber(t *testing.T) {
	// Структура тест-кейса
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		// --- Валидные номера (должны возвращать true) ---
		{
			name:   "Валидный 16-значный номер (Visa/Mastercard)",
			number: "4532718281828182",
			want:   true,
		},
		{
			name:   "Валидный номер с пробелами",
			number: "4532 7182 8182 8182",
			want:   true,
		},
		{
			name:   "Минимальный валидный номер (длина 2)",
			number: "18", // (8*1) + 1 = 9? Нет, справа налево: 8 (индекс 0 от конца, без удвоения) + 1*2 (удвоение) = 10. 10%10 == 0
			want:   true,
		},
		{
			name:   "Валидный номер нечетной длины (длина 15)",
			number: "378282246310005", // Типичный Amex
			want:   true,
		},

		// --- Невалидные номера (должны возвращать false) ---
		{
			name:   "Невалидный номер (ошибка в контрольной цифре)",
			number: "4532718281828183",
			want:   false,
		},
		{
			name:   "Слишком короткий номер (длина 1)",
			number: "0",
			want:   false,
		},
		{
			name:   "Пустая строка",
			number: "",
			want:   false,
		},
		{
			name:   "Строка только из пробелов",
			number: "   ",
			want:   false,
		},

		// --- Ошибки ввода / Некорректные символы (должны возвращать false) ---
		{
			name:   "Содержит буквы",
			number: "453271828182818a",
			want:   false,
		},
		{
			name:   "Содержит спецсимволы (дефисы)",
			number: "4532-7182-8182-8182",
			want:   false,
		},
		{
			name:   "Символы юникода/не-ASCII цифры",
			number: "453271828182818٢", // Арабская двойка
			want:   false,
		},
	}

	// Запуск тестов в цикле
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := auth.ValidLuhnCardNumber(tt.number)
			if got != tt.want {
				t.Errorf("ValidLuhnCardNumber(%q) = %v; want %v", tt.number, got, tt.want)
			}
		})
	}
}
