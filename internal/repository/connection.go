package repository

import (
	"database/sql"
	"goKeeper/internal/config"
	"goKeeper/internal/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewConnect return new connection to DB
func NewConnect() (*sql.DB, error) {
	db, err := sql.Open("pgx", config.DBconfig.DBConnStr)
	if err != nil {
		logger.Log.Errorln(err)
	}
	return db, err
}
