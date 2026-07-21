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

type StorageCards struct {
	DBconnection *sql.DB
}

// создаем новый сторадж для работы с таблицей карт
func NewCardsStorage(DBconn *sql.DB) StorageCards {
	return StorageCards{DBconnection: DBconn}
}

// проверить есть ли данные в БД по login + metadata
func (sc *StorageCards) isExist(ctx context.Context, data model.Cards) bool {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	var resp model.Cards
	sqlStatement := `SELECT c_login, c_chiperdata, c_metadata FROM cards WHERE c_metadata = $1 AND c_login = $2;`
	err := sc.DBconnection.QueryRowContext(ctxWithTimeout, sqlStatement, data.MetaData, data.UserLogin).Scan(&resp.UserLogin, &resp.CipherData, &resp.MetaData)
	if err == sql.ErrNoRows {
		logger.Log.Infoln("card is not exist in DB")
		return false
	}
	return true
}

// InsertCard добавить данные карты в БД, если таких нет
func (sc *StorageCards) InsertCard(ctx context.Context, data model.Cards) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	if sc.isExist(ctxWithTimeout, data) {
		err := errors.New("data is already exists in DB")
		return err
	}
	sqlStatement := `INSERT INTO cards (c_login, c_cipherdata, c_metadata) 
				VALUES ($1, $2, $3)`
	_, err := sc.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, data.UserLogin, data.CipherData, data.MetaData)
	if err != nil {
		logger.Log.Errorln("error while insert card in DB", err)
		return err
	}
	return nil
}

// SelectCard вернуть данные карты, если корректный login и metadata
func (sс *StorageCards) SelectCard(ctx context.Context, login string, metadata string) (resp model.Cards, err error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	sqlStatement := `SELECT c_login, c_cipherdata, c_metadata FROM cards WHERE c_login = $1 AND c_metadata = $2;`
	err = sс.DBconnection.QueryRowContext(ctxWithTimeout, sqlStatement, login, metadata).Scan(&resp.UserLogin, &resp.CipherData, &resp.MetaData)
	if err == sql.ErrNoRows {
		logger.Log.Infoln("card is not exist in DB")
		return resp, err
	}
	return resp, nil
}

// UpdateCard изменить данные карты, если передан корректный login и metadata
func (sc *StorageCards) UpdateCard(ctx context.Context, data model.Cards) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if !sc.isExist(ctxWithTimeout, data) {
		err := errors.New("data is not exists in DB")
		logger.Log.Infoln("data is not exists in DB")
		return err
	}
	sqlStatement := `UPDATE cards 
					SET c_cipherdata = $1 
					WHERE c_login = $2 AND c_metadata = $3;`
	_, err := sc.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, data.CipherData, data.UserLogin, data.MetaData)
	if err != nil {
		logger.Log.Errorln("error while update card to DB", err)
		return err
	}
	return nil
}

// DeleteCard удалить данные карты, если корректный login и metadata D
func (sc *StorageCards) DeleteCard(ctx context.Context, data model.Cards) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	if !sc.isExist(ctxWithTimeout, data) {
		err := errors.New("data is not exists in DB")
		logger.Log.Infoln("data is not exists in DB")
		return err
	}
	sqlStatement := `DELETE FROM cards 
					WHERE c_login = $1 AND c_metadata = $2;`
	_, err := sc.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, data.UserLogin, data.MetaData)
	if err != nil {
		logger.Log.Errorln("error while delete card from DB", err)
		return err
	}
	return nil
}

// DeleteAllCardsByLogin удалить все сохраненные карты пользователя
func (sc *StorageCards) DeleteAllCardsByLogin(ctx context.Context, login string) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	sqlStatement := `DELETE FROM cards 
					WHERE c_login = $1;`
	_, err := sc.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, login)
	if err != nil {
		logger.Log.Errorln("error while delete cards from DB for user: ", login, err)
		return err
	}
	return nil
}
