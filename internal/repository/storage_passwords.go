package repository

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
)

type StoragePasswords struct {
	DBconnection *sql.DB
}

// создаем новый сторадж для работы с таблицей паролей
func NewPasswordsStorage(DBconn *sql.DB) StoragePasswords {
	return StoragePasswords{DBconnection: DBconn}
}

// Close DB connection
func (sp *StoragePasswords) Close() error {
	return sp.DBconnection.Close()
}

// проверить есть ли данные в БД по token + mrtadata
func (sp *StoragePasswords) isExist(ctx context.Context, data model.Passwords) (bool, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	var resp model.Passwords
	sqlStatement := `SELECT p_token, p_pair, p_metadata FROM passwords WHERE p_metadata = $1 AND p_token = $2;`
	err := sp.DBconnection.QueryRowContext(ctxWithTimeout, sqlStatement, data.Metadata, data.Token).Scan(&resp.Token, &resp.Pair, &resp.Metadata)
	if err == sql.ErrNoRows {
		logger.Log.Infoln("passwords is not exist in DB")
		return false, nil
	}
	return true, err
}

// добавить данные если таких нет C
func (sp *StoragePasswords) InsertData(ctx context.Context, data model.Passwords) error {

	return nil
}

// вернуть данные если корректный uuid и metadata R
// изменить данные если передан корректный uuid и metadata U
// удалить данные если корректный uuid и metadata D
