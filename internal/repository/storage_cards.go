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
func (sc *StorageCards) IsExist(ctx context.Context, data model.Cards) (bool, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	var resp model.Cards
	sqlStatement := `SELECT c_login, c_ciperdata, c_metadata FROM cards WHERE c_metadata = $1 AND c_login = $2;`
	err := sc.DBconnection.QueryRowContext(ctxWithTimeout, sqlStatement, data.MetaData, data.UserLogin).Scan(&resp.UserLogin, &resp.CipherData, &resp.MetaData)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Log.Infoln("card is not exist in DB")
			return false, nil // Ошибки нет, данных просто не существует
		}
		logger.Log.Errorln("database error in isExist for cards", err)
		return false, err // Возвращаем реальную ошибку СУБД наружу
	}
	return true, nil
}

// InsertCard добавить данные карты в БД, если таких нет

func (sc *StorageCards) InsertCard(ctx context.Context, data model.Cards) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	// Используем ON CONFLICT DO NOTHING для атомарной вставки
	sqlStatement := `INSERT INTO cards (c_login, c_cipherdata, c_metadata) 
					VALUES ($1, $2, $3)
					ON CONFLICT (c_login, c_metadata) DO NOTHING;`

	result, err := sc.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, data.UserLogin, data.CipherData, data.MetaData)
	if err != nil {
		logger.Log.Errorln("error while insert card in DB", err)
		return err
	}

	// Проверяем, вставилась ли строка
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Errorln("error while getting rows affected in insert", err)
		return err
	}

	// Если 0, значит сработал конфликт уникальности (запись уже была)
	if rowsAffected == 0 {
		err := errors.New("data is already exists in DB")
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

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Log.Infoln("card is not exist in DB")
		} else {
			logger.Log.Errorln("error while scanning card row from DB", err)
		}
		return resp, err // Возвращаем ошибку (включая ErrNoRows) наружу
	}

	return resp, nil
}

// UpdateCard изменить данные карты, если передан корректный login и metadata
func (sc *StorageCards) UpdateCard(ctx context.Context, data model.Cards) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	sqlStatement := `UPDATE cards 
					SET c_cipherdata = $1 
					WHERE c_login = $2 AND c_metadata = $3;`

	result, err := sc.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, data.CipherData, data.UserLogin, data.MetaData)
	if err != nil {
		return err
	}

	// Проверяем, сколько строк было обновлено
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Errorln("error while getting rows affected in update card", err)
		return err
	}

	// Если 0 строк, значит карты с таким login и metadata не существовало
	if rowsAffected == 0 {
		err := errors.New("data is not exists in DB")
		logger.Log.Infoln("data is not exists in DB")
		return err
	}

	return nil
}

// DeleteCard удалить данные карты, если корректный login и metadata
func (sc *StorageCards) DeleteCard(ctx context.Context, data model.Cards) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	sqlStatement := `DELETE FROM cards 
					WHERE c_login = $1 AND c_metadata = $2;`

	result, err := sc.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, data.UserLogin, data.MetaData)
	if err != nil {
		return err
	}

	// Проверяем, сколько строк было удалено
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Errorln("error while getting rows affected in delete card", err)
		return err
	}

	// Если 0 строк, значит карты с таким login и metadata не существовало
	if rowsAffected == 0 {
		err := errors.New("data is not exists in DB")
		logger.Log.Infoln("data is not exists in DB")
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
		return err
	}
	return nil
}
