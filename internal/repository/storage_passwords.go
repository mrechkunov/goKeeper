package repository

import (
	"context"
	"database/sql"
	"errors"
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

// проверить есть ли данные в БД по uuid + metadata
func (sp *StoragePasswords) isExist(ctx context.Context, data model.Passwords) bool {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	var resp model.Passwords
	sqlStatement := `SELECT p_uuid, p_pair, p_metadata FROM passwords WHERE p_metadata = $1 AND p_uuid = $2;`
	err := sp.DBconnection.QueryRowContext(ctxWithTimeout, sqlStatement, data.Metadata, data.Uuid).Scan(&resp.Uuid, &resp.Pair, &resp.Metadata)
	if err == sql.ErrNoRows {
		logger.Log.Infoln("passwords is not exist in DB")
		return false
	}
	return true
}

// добавить данные если таких нет C
func (sp *StoragePasswords) InsertData(ctx context.Context, data model.Passwords) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if sp.isExist(ctxWithTimeout, data) {
		err := errors.New("data is already exists in DB")
		logger.Log.Infoln("data is already exists in DB")
		return err
	}
	sqlStatement := `INSERT INTO passwords (p_uuid, p_pair, p_metadata) 
				VALUES ($1, $2, $3)`
	_, err := sp.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, data.Uuid, data.Pair, data.Metadata)
	if err != nil {
		logger.Log.Errorln("error while insert passwords to DB", err)
		return err
	}
	return nil
}

// вернуть данные если корректный uuid и metadata R
func (sp *StoragePasswords) GetData(ctx context.Context, uuid string, metadata string) (resp model.Passwords, err error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	sqlStatement := `SELECT p_uuid, p_pair, p_metadata FROM passwords WHERE p_uuid = $1 AND p_metadata = $2;`
	err = sp.DBconnection.QueryRowContext(ctxWithTimeout, sqlStatement, uuid, metadata).Scan(&resp.Uuid, &resp.Pair, &resp.Metadata)
	if err == sql.ErrNoRows {
		logger.Log.Infoln("passwords is not exist in DB")
		return resp, err
	}
	return resp, nil
}

// изменить данные если передан корректный uuid и metadata U
func (sp *StoragePasswords) UpdateData(ctx context.Context, data model.Passwords) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if !sp.isExist(ctxWithTimeout, data) {
		err := errors.New("data is not exists in DB")
		logger.Log.Infoln("data is not exists in DB")
		return err
	}
	sqlStatement := `UPDATE passwords 
					SET p_pair = $1,
					WHERE p_uuid = $2 AND p_metadata = $3;`
	_, err := sp.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, data.Pair, data.Uuid, data.Metadata)
	if err != nil {
		logger.Log.Errorln("error while update passwords to DB", err)
		return err
	}
	return nil
}

// удалить данные если корректный uuid и metadata D
func (sp *StoragePasswords) DeleteData(ctx context.Context, data model.Passwords) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if !sp.isExist(ctxWithTimeout, data) {
		err := errors.New("data is not exists in DB")
		logger.Log.Infoln("data is not exists in DB")
		return err
	}
	sqlStatement := `DELETE FROM passwords 
					WHERE p_uuid = $1 AND p_metadata = $2;`
	_, err := sp.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, data.Uuid, data.Metadata)
	if err != nil {
		logger.Log.Errorln("error while delete passwords to DB", err)
		return err
	}
	return nil
}

// удалить все сохраненные пароли пользователя
func (sp *StoragePasswords) DeleteDataByUuid(ctx context.Context, uuid string) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	sqlStatement := `DELETE FROM passwords 
					WHERE p_uuid = $1;`
	_, err := sp.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, uuid)
	if err != nil {
		logger.Log.Errorln("error while delete passwords to DB for user: ", uuid, err)
		return err
	}
	return nil
}
