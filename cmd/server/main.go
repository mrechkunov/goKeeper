package main

import (
	"fmt"
	"net"
	"os"

	"github.com/mrechkunov/goKeeper.git/internal/config"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/service"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	config.Init()
	logger.Log.Infoln("Reading config")
	defer logger.Log.Sync() // закрываем логгер при выходе из main
	// Storage := repository.NewUsersStorage(config.DBconn)
	// defer Storage.Close()

	// Нужно определить порт для сервера из конфига
	listen, err := net.Listen("tcp", config.SrvConfig.GRPCServerAddress)
	if err != nil {
		logger.Log.Warnln("ошибка при инициализации listener", "error", err)
		os.Exit(1)
	}
	// Загрузка TLS-сертификата и ключа сервера
	creds, err := credentials.NewServerTLSFromFile("./internal/config/cert/server.crt", "./internal/config/cert/server.key")
	if err != nil {
		logger.Log.Warnln("Ошибка загрузки TLS сертификата:", err)
	}
	// Создаем gRPC сервер без зарегистрированной службы
	s := grpc.NewServer(grpc.Creds(creds))
	// Регистрируем сервис
	pb.RegisterGoKeeperServer(s, &service.GoKeeperServer{})
	fmt.Println("сервер gRPC начал работу")
	// Получение запроса gRpc
	go func() {
		if err := s.Serve(listen); err != nil {
			logger.Log.Warnln("ошибка при работе сервера", "error", err)
			os.Exit(1)
		}
	}()

}
