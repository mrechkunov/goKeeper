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
func NewPasswordsStorage(DBconn *sql.DB) *StoragePasswords {
	return &StoragePasswords{DBconnection: DBconn}
}

// проверить есть ли данные в БД по login + metadata
func (sp *StoragePasswords) IsExist(ctx context.Context, data model.Passwords) bool {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	var resp model.Passwords
	sqlStatement := `SELECT p_login, p_pair, p_metadata FROM passwords WHERE p_metadata = $1 AND p_login = $2;`
	err := sp.DBconnection.QueryRowContext(ctxWithTimeout, sqlStatement, data.MetaData, data.UserLogin).Scan(&resp.UserLogin, &resp.Pair, &resp.MetaData)
	if err == sql.ErrNoRows {
		return false
	}
	return true
}

// InsertPassword добавить данные пароля, если таких нет
func (sp *StoragePasswords) InsertPassword(ctx context.Context, data model.Passwords) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	if sp.IsExist(ctxWithTimeout, data) {
		err := errors.New("data is already exists in DB")
		return err
	}
	sqlStatement := `INSERT INTO passwords (p_login, p_pair, p_metadata) 
				VALUES ($1, $2, $3)`
	_, err := sp.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, data.UserLogin, data.Pair, data.MetaData)
	if err != nil {
		return err
	}
	return nil
}

// SelectPassword вернуть данные пароля, если корректный login и metadata R
func (sp *StoragePasswords) SelectPassword(ctx context.Context, login string, metadata string) (resp model.Passwords, err error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	sqlStatement := `SELECT p_login, p_pair, p_metadata FROM passwords WHERE p_login = $1 AND p_metadata = $2;`
	err = sp.DBconnection.QueryRowContext(ctxWithTimeout, sqlStatement, login, metadata).Scan(&resp.UserLogin, &resp.Pair, &resp.MetaData)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Log.Infoln("passwords is not exist in DB")
		} else {
			logger.Log.Infoln("error while scanning row from DB", err)
		}
		return resp, err
	}

	return resp, nil
}

// UpdatePassword изменить данные, если передан корректный login и metadata
func (sp *StoragePasswords) UpdatePassword(ctx context.Context, data model.Passwords) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	sqlStatement := `UPDATE passwords 
					SET p_pair = $1 
					WHERE p_login = $2 AND p_metadata = $3;`

	result, err := sp.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, data.Pair, data.UserLogin, data.MetaData)
	if err != nil {
		return err
	}

	// Проверяем, сколько строк обновилось в базе данных
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Errorln("error while getting rows affected", err)
		return err
	}

	// Если 0, значит записи с таким login и metadata не существовало
	if rowsAffected == 0 {
		err := errors.New("data is not exists in DB")
		return err
	}

	return nil
}

// DeletePassword удалить данные пароля, если корректный login и metadata
func (sp *StoragePasswords) DeletePassword(ctx context.Context, data model.Passwords) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	sqlStatement := `DELETE FROM passwords 
					WHERE p_login = $1 AND p_metadata = $2;`

	result, err := sp.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, data.UserLogin, data.MetaData)
	if err != nil {
		return err
	}

	// Проверяем, сколько строк было удалено
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Errorln("error while getting rows affected in delete", err)
		return err
	}

	// Если 0 строк удалено, значит, такой записи изначально не было
	if rowsAffected == 0 {
		err := errors.New("data is not exists in DB")
		logger.Log.Infoln("data is not exists in DB for deletion")
		return err
	}

	return nil
}

// DeleteAllPasswordsByLogin удалить все сохраненные пароли пользователя
func (sp *StoragePasswords) DeleteAllPasswordsByLogin(ctx context.Context, login string) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	sqlStatement := `DELETE FROM passwords 
					WHERE p_login = $1;`
	_, err := sp.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, login)
	if err != nil {
		return err
	}
	return nil
}
