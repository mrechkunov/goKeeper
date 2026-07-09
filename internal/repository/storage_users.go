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

type StorageUsers struct {
	DBconnection *sql.DB
}

// создаем новый сторадж для работы с таблицей пользователей
func NewUsersStorage(DBconn *sql.DB) StorageUsers {
	return StorageUsers{DBconnection: DBconn}
}

// is user exist in db ?
func (su *StorageUsers) IsExist(ctx context.Context, login string) (bool, error) {
	var user model.Users
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	sqlStatement := `SELECT u_login, u_password, u_token FROM users WHERE u_login = $1;`
	err := su.DBconnection.QueryRowContext(ctxWithTimeout, sqlStatement, login).Scan(&user.Login, &user.Password, &user.Token)
	if err != nil && err != sql.ErrNoRows {
		logger.Log.Errorln("Error while try isExist user: ", err)
		return false, err
	}
	if err == sql.ErrNoRows {
		logger.Log.Infoln("user with login: ", login, "is not exist in DB")
		return false, nil
	}
	return true, nil
}

// добавить пользователя C
func (su *StorageUsers) CreateUser(ctx context.Context, user model.Users) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	sqlStatement := `INSERT INTO users (u_login, u_password, u_token) 
				VALUES ($1, $2, $3)`
	_, err := su.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, user.Login, user.Password, user.Token)
	if err != nil {
		logger.Log.Errorln("error while insert user to DB", err)
		return err
	}
	return nil
}

// авторизовать пользователя (проверить что он есть и вернуть token) R
func (su *StorageUsers) ReadUser(ctx context.Context, login string) (user model.Users, err error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	exist, err := su.IsExist(ctxWithTimeout, login)
	if err != nil {
		logger.Log.Errorln("Error while read user", err)
	}
	if !exist {
		return user, err
	}
	sqlStatement := `SELECT u_login, u_password, u_token FROM users WHERE u_login = $1;`
	err = su.DBconnection.QueryRowContext(ctxWithTimeout, sqlStatement, login).Scan(&user.Login, &user.Password, &user.Token)
	return user, nil
}

// изменить данные пользователя (только если token верный) U
func (su *StorageUsers) UpdateUser(ctx context.Context, user model.Users) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	sqlStatement := `UPDATE users 
				SET u_password = $1, 
				u_token = $2
				WHERE u_login = $3;`
	_, err := su.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, user.Password, user.Token, user.Login)
	if err != nil {
		logger.Log.Errorln("error while update user data in DB", err)
		return err
	}
	return nil
}

// удалить пользователя и все его данные D
func (su *StorageUsers) DeleteUser(ctx context.Context, user model.Users) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	sqlStatement := `DELETE FROM users
				WHERE u_login = $1;`
	_, err := su.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, user.Login)
	if err != nil {
		logger.Log.Errorln("error while delete user in DB", err)
		return err
	}

	//TODO: make sync group and run in gorutines
	PassStorage := NewPasswordsStorage(su.DBconnection)
	PassStorage.DeleteDataByToken(ctxWithTimeout, user.Token)
	PassStorage.Close()

	//TODO: delete in cards table
	//TODO: delete in binary file table

	return nil
}
