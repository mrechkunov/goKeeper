package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mrechkunov/goKeeper.git/internal/logger"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
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
	//Создаем клиента для сервиса
	client := pb.NewGoKeeperClient(conn)
	//Делаем запрос к серверу

	var user pb.User
	user.SetLogin("ivan")
	user.SetPasswordHash("test")
	var header metadata.MD
	resp, err := client.RegisterUser(context.Background(), &user, grpc.Header(&header))
	if err != nil {
		logger.Log.Errorln("error while register user: ", err)
	}
	logger.Log.Infoln("server resp:", resp)

	if vals := header.Get("authorization"); len(vals) > 0 {
		token := vals[0]

		fmt.Println("token", token)

	}
	if vals := header.Get("userlogin"); len(vals) > 0 {

		login := vals[0]

		fmt.Println("login", login)
	}

}
