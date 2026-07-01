package main

import (
	"github.com/mrechkunov/goKeeper.git/internal/config"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/repository"
)

func main() {
	config.Init()
	Storage := repository.StorageInit()
	defer Storage.Close()
	defer logger.Log.Sync() // закрываем логгер при выходе из main
	logger.Log.Infoln("Reading config")
}
