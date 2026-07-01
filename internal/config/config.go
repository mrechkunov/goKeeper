package config

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/mrechkunov/goKeeper.git/internal/logger"
)

type DBconfig struct {
	MigrationsPath string `json:"migrations_path"`
	DBConnStr      string `json:"database_dsn"`
}

type GRPCSrvConf struct {
	GRPCServerAddress string `json:"grpc_server_address"`
}

func Init() {
	cf := flag.String("c", "", "config file address")

	// если есть адрес конфига, парсим сначала его и присваеваем все значения структуре конфигурации
	if ConfigFileAddress != "" {
		// Считайтываем файл целиком
		data, err := os.ReadFile(ConfigFileAddress)
		if err != nil {
			logger.Log.Warnln("Ошибка чтения файла:", err)
		}
		// Распарсим JSON в структуру
		err = json.Unmarshal(data, &ConfigFileData)
		if err != nil {
			logger.Log.Warnln("Ошибка парсинга JSON:", err)
		}
	}
}
