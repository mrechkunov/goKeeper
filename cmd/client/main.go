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
	// passHash, err := auth.HashPassword("password")
	// if err != nil {
	// 	logger.Log.Infoln("error while hash password")
	// }
	fmt.Println("--------------register user----------")
	var user pb.User
	user.SetLogin("ivan")
	user.SetPasswordHash("password")
	var tokenAuth string
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

	fmt.Println("--------------Authenticate user----------")
	respAuth, err := client.AuthenticateUser(context.Background(), &user, grpc.Header(&header))
	if err != nil {
		logger.Log.Errorln("error while Authenticate user: ", err)
	}
	logger.Log.Infoln("server resp:", respAuth)
	if vals := header.Get("authorization"); len(vals) > 0 {
		tokenAuth = vals[0]
		fmt.Println("token", tokenAuth)
	}
	fmt.Println("login:", respAuth.GetLogin())
	fmt.Println("pass:", respAuth.GetPasswordHash())
	fmt.Println("token:", tokenAuth)

	fmt.Println("--------------Edit user----------")
	md := metadata.Pairs(
		"authorization", tokenAuth,
	)
	user.SetLogin("ivan")
	user.SetPasswordHash("password123")
	ctxWithAuth := metadata.NewOutgoingContext(context.Background(), md)
	respEdit, err := client.EditUser(ctxWithAuth, &user, grpc.Header(&header))
	if err != nil {
		logger.Log.Errorln("error while edit user: ", err)
	}
	logger.Log.Infoln("server resp:", respEdit)
	if vals := header.Get("authorization"); len(vals) > 0 {
		tokenAuth = vals[0]
		fmt.Println("token", tokenAuth)
	}
	fmt.Println("login:", respEdit.GetLogin())
	fmt.Println("pass:", respEdit.GetPasswordHash())
	fmt.Println("--------------Delete user----------")
	user.SetLogin("ivan")
	ctxWithAuth = metadata.NewOutgoingContext(context.Background(), md)
	respDelete, err := client.DeleteUser(ctxWithAuth, &user, grpc.Header(&header))
	if err != nil {
		logger.Log.Errorln("error while delete user: ", err)
	}
	logger.Log.Infoln("server resp:", respDelete)
	if vals := header.Get("authorization"); len(vals) > 0 {
		tokenAuth = vals[0]
		fmt.Println("token", tokenAuth)
	}
}
