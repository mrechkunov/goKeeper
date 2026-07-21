package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/mrechkunov/goKeeper.git/internal/auth"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	cliservice "github.com/mrechkunov/goKeeper.git/internal/service/client"
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
		Login:        "michael",
		PasswordHash: "pass",
	}
	token, err := cliservice.AuthenticateUser(context.Background(), client, user)
	if err != nil {
		logger.Log.Warnln(err)
	}
	tokenMgr.UpdateToken(token)
	fmt.Println("token", token)
	fmt.Println("--------------Authenticate user done----------")
	fmt.Println("--------------Safe pass----------")

	login, err := auth.GetLoginByToken(token)
	if err != nil {
		logger.Log.Warnln("error while get login by token", err)
	}
	passwordToSave := model.Passwords{
		UserLogin:      login,
		MetaData:       "eweeeeererwew.ru/test",
		LoginToSave:    "testwwe",
		PasswordToSave: "tests3434ds435",
	}
	err = cliservice.SavePass(context.Background(), client, passwordToSave)
	if err != nil {
		fmt.Println("save error:", err)
	}
	fmt.Println("--------------Save Pass Done----------")
	fmt.Println("--------------Get Pass----------")
	dataToGet := model.Passwords{
		UserLogin: passwordToSave.UserLogin,
		MetaData:  passwordToSave.MetaData,
	}
	data, err := cliservice.GetPass(context.Background(), client, dataToGet)
	if err != nil {
		fmt.Println("get error:", err)
	}
	fmt.Println("ulogin: ", data.UserLogin)
	fmt.Println("metadata: ", data.MetaData)
	fmt.Println("login: ", data.LoginToSave)
	fmt.Println("pass: ", data.PasswordToSave)
	fmt.Println("--------------Get Pass Done----------")
	fmt.Println("--------------Edit Pass ----------")
	passwordToedit := model.Passwords{
		UserLogin:      login,
		MetaData:       "eweeeeererwew.ru/test",
		LoginToSave:    "testwwe",
		PasswordToSave: "EditedPass",
	}
	err = cliservice.EditPass(context.Background(), client, passwordToedit)
	if err != nil {
		fmt.Println("edit error:", err)
	}
	dataToGetAfterEdit := model.Passwords{
		UserLogin: passwordToedit.UserLogin,
		MetaData:  passwordToedit.MetaData,
	}
	data, err = cliservice.GetPass(context.Background(), client, dataToGetAfterEdit)
	if err != nil {
		fmt.Println("get error:", err)
	}
	fmt.Println("ulogin: ", data.UserLogin)
	fmt.Println("metadata: ", data.MetaData)
	fmt.Println("login: ", data.LoginToSave)
	fmt.Println("pass: ", data.PasswordToSave)
}
