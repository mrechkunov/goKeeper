package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mrechkunov/goKeeper.git/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Ключ контекста
type contextKey string

const userLoginKey contextKey = "userLogin"

// Фейковая функция извлечения токена, если она объявлена в том же пакете.
func extractToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("empty header")
	}
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}
	return authHeader, nil
}

func TestAuthInterceptor(t *testing.T) {
	tests := []struct {
		name          string
		fullMethod    string
		ctx           context.Context
		expectedCode  codes.Code
		checkContext  bool
		expectedLogin string
	}{
		{
			name:         "Успех: Публичный эндпоинт регистрации не требует токен",
			fullMethod:   "/mrechkunov.goKeeper.proto.GoKeeper/RegisterUser",
			ctx:          context.Background(),
			expectedCode: codes.OK,
		},
		{
			name:         "Успех: Публичный эндпоинт авторизации не требует токен",
			fullMethod:   "/mrechkunov.goKeeper.proto.GoKeeper/AuthenticateUser",
			ctx:          context.Background(),
			expectedCode: codes.OK,
		},
		{
			name:         "Ошибка: Отсутствуют метаданные в контексте",
			fullMethod:   "/mrechkunov.goKeeper.proto.GoKeeper/GetSecretData",
			ctx:          context.Background(),
			expectedCode: codes.Unauthenticated,
		},
		{
			name:         "Ошибка: Метаданные есть, но нет заголовка authorization",
			fullMethod:   "/mrechkunov.goKeeper.proto.GoKeeper/GetSecretData",
			ctx:          metadata.NewIncomingContext(context.Background(), metadata.Pairs("custom-header", "value")),
			expectedCode: codes.Unauthenticated,
		},
		{
			name:         "Ошибка: Невалидный/сломанный токен",
			fullMethod:   "/mrechkunov.goKeeper.proto.GoKeeper/GetSecretData",
			ctx:          metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer invalid.token.string")),
			expectedCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка UnaryServerInfo
			info := &grpc.UnaryServerInfo{
				FullMethod: tt.fullMethod,
			}

			// Создаем фейковый обработчик (handler), который проверяет, дошло ли до него управление
			// и записался ли логин в контекст
			handler := func(handlerCtx context.Context, req interface{}) (interface{}, error) {
				if tt.checkContext {
					login, ok := handlerCtx.Value(userLoginKey).(string)
					if !ok || login != tt.expectedLogin {
						t.Errorf("Handler context expected login %q, got %q", tt.expectedLogin, login)
					}
				}
				return "success_response", nil
			}

			// Вызов интерцептора
			_, err := service.AuthInterceptor(tt.ctx, nil, info, handler)

			// Проверка кода ответа gRPC
			if tt.expectedCode == codes.OK {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error code %v, got nil", tt.expectedCode)
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Expected gRPC status error, got: %v", err)
				}
				if st.Code() != tt.expectedCode {
					t.Errorf("Expected gRPC code %v, got %v (message: %s)", tt.expectedCode, st.Code(), st.Message())
				}
			}
		})
	}
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		wantToken   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "Успешно: Стандартный заголовок",
			header:    "Bearer token123",
			wantToken: "token123",
			wantErr:   false,
		},
		{
			name:      "Успешно: Множественные пробелы между Bearer и токеном",
			header:    "Bearer    token123",
			wantToken: "token123",
			wantErr:   false,
		},
		{
			name:      "Успешно: Пробелы в начале и в конце строки",
			header:    "   Bearer token123   ",
			wantToken: "token123",
			wantErr:   false,
		},
		{
			name:      "Успешно: Регистронезависимое слово Bearer",
			header:    "bEaReR token123",
			wantToken: "token123",
			wantErr:   false,
		},
		{
			name:        "Ошибка: Отсутствует префикс Bearer",
			header:      "token123",
			wantToken:   "",
			wantErr:     true,
			errContains: "неверный формат заголовка авторизации",
		},
		{
			name:        "Ошибка: Токен состоит из нескольких слов",
			header:      "Bearer token part two",
			wantToken:   "",
			wantErr:     true,
			errContains: "неверный формат заголовка авторизации",
		},
		{
			name:        "Ошибка: Пустая строка",
			header:      "",
			wantToken:   "",
			wantErr:     true,
			errContains: "неверный формат заголовка авторизации",
		},
		{
			name:        "Ошибка: Другая схема авторизации (Basic)",
			header:      "Basic dXNlcjpwYXNz",
			wantToken:   "",
			wantErr:     true,
			errContains: "неверный формат заголовка авторизации",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := service.ExtractToken(tt.header)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Errorf("extractToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 2. Если ошибка ожидалась, проверяем её текст
			if tt.wantErr && err != nil {
				if err.Error() != tt.errContains {
					t.Errorf("extractToken() error string = %q, expected %q", err.Error(), tt.errContains)
				}
			}

			// 3. Проверяем сам токен
			if gotToken != tt.wantToken {
				t.Errorf("extractToken() gotToken = %q, want %q", gotToken, tt.wantToken)
			}
		})
	}
}
