package main

import (
	"github.com/mrechkunov/goKeeper.git/internal/config"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/repository"
)

func main() {
	config.Init()
	logger.Log.Infoln("Reading config")
	Storage := repository.NewUsersStorage(config.DBconn)
	defer Storage.Close()
	defer logger.Log.Sync() // закрываем логгер при выходе из main

}
