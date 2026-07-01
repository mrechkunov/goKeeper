package repository

import (
	"database/sql"

	"github.com/mrechkunov/goKeeper.git/internal/config"
	"github.com/mrechkunov/goKeeper.git/internal/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewConnect return new connection to DB
func NewConnect() (*sql.DB, error) {
	db, err := sql.Open("pgx", config.SrvConfig.DBConnStr)
	if err != nil {
		logger.Log.Errorln(err)
	}
	return db, err
}
