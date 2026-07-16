package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/mrechkunov/goKeeper.git/internal/cryptodata"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/service"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	address    = "localhost:50010"
	certFile   = "./internal/config/cert/server.crt" // Путь к сертификату CA, который выдал сертификат серверу
	serverName = "localhost"                         // Должен совпадать с Common Name (CN) в сертификате сервера
)

// TokenManager отвечает за безопасное обновление и хранение токена
type TokenManager struct {
	mu          sync.RWMutex
	accessToken string
}

// UpdateToken обновляет токен (например, пришел новый от OAuth2 провайдера)
func (tm *TokenManager) UpdateToken(newToken string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.accessToken = newToken
}

// GetToken возвращает текущий токен
func (tm *TokenManager) GetToken() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.accessToken
}

// PerRPCCreds реализует интерфейс credentials.PerRPCCredentials
type PerRPCCreds struct {
	tokenManager *TokenManager
}

// GetRequestMetadata вызывается перед каждым gRPC-вызовом
func (c *PerRPCCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	// Берем актуальный токен из менеджера
	token := c.tokenManager.GetToken()

	return map[string]string{
		"authorization": "Bearer " + token,
	}, nil
}

// RequireTransportSecurity указывает, требуется ли TLS для передачи учетных данных
func (c *PerRPCCreds) RequireTransportSecurity() bool {
	// В продакшене всегда должно быть true (gRPC отклонит отправку токена без TLS)
	return true
}

func main() {
	//Настраиваем TLS-конфигурацию
	creds, err := credentials.NewClientTLSFromFile(certFile, serverName)
	if err != nil {
		log.Fatalf("Не удалось загрузить сертификаты: %v", err)
	}
	// 1. Инициализируем менеджер токенов
	tokenMgr := &TokenManager{}
	tokenMgr.UpdateToken("initial_token_value")

	// 2. Создаем credentials.PerRPCCredentials
	perRPC := &PerRPCCreds{
		tokenManager: tokenMgr,
	}

	//Устанавливаем соединение с сервером
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(perRPC))
	if err != nil {
		log.Fatalf("Не удалось подключиться к gRPC серверу: %v", err)
	}
	defer conn.Close()
	//Создаем клиента для сервиса
	client := pb.NewGoKeeperClient(conn)

	fmt.Println("--------------Authenticate user----------")
	user := model.Users{
		Login:        "ivan",
		PasswordHash: "pass",
	}
	token, err := service.AuthenticateUser(context.Background(), client, user)
	if err != nil {
		logger.Log.Warnln(err)
	}
	tokenMgr.UpdateToken(token)
	fmt.Println("token", token)
	login := user.Login

	loginToSave := "testwewe"
	passToSave := "test435"
	metaToSave := "yandesdsdwewew3x.ru/test"
	pair, err := cryptodata.CryptoPair(loginToSave, passToSave)
	if err != nil {
		logger.Log.Infoln(err)
	}
	passwordToSave := model.Passwords{
		Login:    login,
		Pair:     pair,
		Metadata: metaToSave,
	}

	err = service.SavePass(context.Background(), client, passwordToSave)
	if err != nil {
		fmt.Println("save error:", err)
	}

	dataToGet := model.Passwords{
		Login:    passwordToSave.Login,
		Metadata: passwordToSave.Metadata,
	}
	data, err := service.GetPass(context.Background(), client, dataToGet)
	if err != nil {
		fmt.Println("get error:", err)
	}
	//fmt.Println(data)

	getLogin, getPass, err := cryptodata.DecryptPair(data.Pair)
	fmt.Println("ulogin: ", data.Login)
	fmt.Println("metadata: ", data.Metadata)
	fmt.Println("login: ", getLogin)
	fmt.Println("pass: ", getPass)
}
