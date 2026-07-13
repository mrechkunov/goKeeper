package main

import (
	"context"
	"log"

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
	//Настраиваем TLS-конфигурацию
	creds, err := credentials.NewClientTLSFromFile(certFile, serverName)
	if err != nil {
		log.Fatalf("Не удалось загрузить сертификаты: %v", err)
	}

	//Устанавливаем соединение с сервером
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Не удалось подключиться к gRPC серверу: %v", err)
	}
	defer conn.Close()
	// 3. Создаем клиента для сервиса
	client := pb.NewGoKeeperClient(conn)
	// 4. Делаем запрос к серверу

	var user pb.User
	user.SetLogin("ivan")
	user.SetPasswordHash("")
	resp, err := client.RegisterUser(context.Background(), &user)
	if err != nil {
		logger.Log.Errorln("error while register user: ", err)
	}
	logger.Log.Infoln("server resp:", resp)

}
