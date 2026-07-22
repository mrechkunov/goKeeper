package auth_test

import (
	"testing"

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

// func TestGenerateToken(t *testing.T) {
// 	tests := []struct {
// 		name string // description of this test case
// 		// Named input parameters for target function.
// 		uLogin  string
// 		want    string
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, gotErr := auth.GenerateToken(tt.uLogin)
// 			if gotErr != nil {
// 				if !tt.wantErr {
// 					t.Errorf("GenerateToken() failed: %v", gotErr)
// 				}
// 				return
// 			}
// 			if tt.wantErr {
// 				t.Fatal("GenerateToken() succeeded unexpectedly")
// 			}
// 			// TODO: update the condition below to compare got with tt.want.
// 			if true {
// 				t.Errorf("GenerateToken() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
