package main

import (
	"context"
	"log"
	"time"

	"github.com/mrechkunov/goKeeper.git/internal/logger"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	address    = "localhost:50010"
	certFile   = "./internal/config/cert/server.crt" // Путь к сертификату CA, который выдал сертификат серверу
	serverName = "localhost"                         // Должен совпадать с Common Name (CN) в сертификате сервера
)

func main() {
	// 1. Настраиваем TLS-конфигурацию
	creds, err := credentials.NewClientTLSFromFile(certFile, serverName)
	if err != nil {
		log.Fatalf("Не удалось загрузить сертификаты: %v", err)
	}

	// 2. Устанавливаем соединение с сервером
	// Используем grpc.NewClient вместо устаревшего grpc.Dial
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Не удалось подключиться к gRPC серверу: %v", err)
	}
	defer conn.Close()

	// 3. Создаем клиента для сервиса
	client := pb.NewGoKeeperClient(conn)
	// 4. Делаем запрос к серверу
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user pb.User
	user.SetLogin("ivan")
	user.SetPasswordHash("test")
	resp, err := client.RegisterUser(ctx, &user)
	if err != nil {
		logger.Log.Errorln("error while register user: ", err)
	}
	logger.Log.Infoln("server resp:", resp)
}

// Вызываем метод сервиса (пример: SayHello)
//req := &pb.GoKeeper_RegisterUser_FullMethodName()
// 	resp, err := client.RegisterUser(ctx, req)
// 	if err != nil {
// 		log.Fatalf("Ошибка выполнения вызова SayHello: %v", err)
// 	}

// 	log.Printf("Ответ от сервера: %s", resp.GetMessage())
// }
