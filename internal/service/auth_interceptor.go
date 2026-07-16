package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/mrechkunov/goKeeper.git/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Определяем собственный тип для ключа, чтобы избежать конфликтов
type contextKey string

// Создаем константу ключа
const userLoginKey contextKey = "userLogin"

// AuthInterceptor извлекает токен и проверяет его валидность.
func AuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Список методов, для которых аутентификация не требуется
	if info.FullMethod == "/mrechkunov.goKeeper.proto.GoKeeper/AuthenticateUser" {
		return handler(ctx, req)
	}
	if info.FullMethod == "/mrechkunov.goKeeper.proto.GoKeeper/RegisterUser" {
		return handler(ctx, req)
	}
	// Извлекаем метаданные из контекста
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no metadata in context")
	}
	// Получаем заголовок Authorization
	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no token in metadata")
	}
	authHeader := values[0]
	token, err := extractToken(authHeader)
	// if err != nil {
	// 	return nil, status.Error(codes.Unauthenticated, err.Error())
	// }

	// Валидация токена
	err = auth.ValidateToken(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "token is not valid")
	}
	login, err := auth.GetLoginByToken(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "no login in token")
	}
	//записываем в контекст логин пользователя
	ctxWithLogin := context.WithValue(ctx, userLoginKey, login)
	// Если всё успешно, передаем управление дальше по цепочке
	return handler(ctxWithLogin, req)
}

// extractToken парсит заголовок Bearer Token
func extractToken(header string) (string, error) {
	parts := strings.Split(header, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("неверный формат заголовка авторизации")
	}
	return parts[1], nil
}
