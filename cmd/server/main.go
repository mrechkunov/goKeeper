package main

import (
	"goKeeper/internal/config"
	"goKeeper/internal/logger"
)

func main() {
	config.Init()
	Storage := repository.StorageInit()
	defer Storage.Close()
	defer logger.Log.Sync() // закрываем логгер при выходе из main
	logger.Log.Infoln("Reading config")
}
