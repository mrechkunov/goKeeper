package config

import (
	"database/sql"
	"encoding/json"
	"flag"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
)

type ConfigSrvData struct {
	MigrationsPath    string `json:"migrations_path"`
	DBConnStr         string `json:"database_dsn"`
	GRPCServerAddress string `json:"grpc_server_address"`
}

var configFileAddress string
var SrvConfig ConfigSrvData
var DBconn *sql.DB

func Init() {
	cf := flag.String("c", "", "config file address")
	flag.Parse()
	configFileAddress = *cf
	// если есть адрес конфига, парсим сначала его и присваеваем все значения структуре конфигурации
	if configFileAddress != "" {
		// Считайтываем файл целиком
		data, err := os.ReadFile(configFileAddress)
		if err != nil {
			logger.Log.Warnln("Ошибка чтения файла:", err)
		}
		// Распарсим JSON в структуру
		err = json.Unmarshal(data, &SrvConfig)
		if err != nil {
			logger.Log.Warnln("Ошибка парсинга JSON:", err)
		}
	}
	// create connect to DB and run Up all migrations
	var err error
	DBconn, err = NewConnect()
	if err != nil {
		logger.Log.Errorln("error while connecting to DB (configure service)", err)
	}
	migrations(DBconn)
}

func NewConnect() (*sql.DB, error) {
	db, err := sql.Open("pgx", SrvConfig.DBConnStr)
	if err != nil {
		logger.Log.Errorln(err)
	}
	return db, err
}

func migrations(DBconn *sql.DB) {
	m, err := migrate.New(
		SrvConfig.MigrationsPath,
		SrvConfig.DBConnStr)
	if err != nil {
		logger.Log.Errorln("error initializing migrate:", err)
	}
	// Apply all available migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Log.Errorln("error applying migrations:", err)
	}
	logger.Log.Infoln("database migrations applied successfully!")
	err = DBconn.Ping()
	if err != nil {
		logger.Log.Warnln("error while ping DB after migratioans applied", err)
	}
}
