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

// NewUsersStorage new storage for work with users table in db
func NewUsersStorage(DBconn *sql.DB) StorageUsers {
	return StorageUsers{DBconnection: DBconn}
}

// IsExist return true if user with login is exist in db
func (su *StorageUsers) IsExist(ctx context.Context, login string) (bool, error) {
	var user model.Users
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	sqlStatement := `SELECT u_login, u_password FROM users WHERE u_login = $1;`
	err := su.DBconnection.QueryRowContext(ctxWithTimeout, sqlStatement, login).Scan(&user.Login, &user.PasswordHash)
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

// CreateUser добавить пользователя C
func (su *StorageUsers) CreateUser(ctx context.Context, user model.Users) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	sqlStatement := `INSERT INTO users (u_login, u_password) 
				VALUES ($1, $2)`
	_, err := su.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, user.Login, user.PasswordHash)
	if err != nil {
		logger.Log.Errorln("error while insert user to DB", err)
		return err
	}
	return nil
}

// ReadUser Вернуть пользователя по логину (проверить что он есть и вернуть user) R
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
	sqlStatement := `SELECT u_login, u_password FROM users WHERE u_login = $1;`
	err = su.DBconnection.QueryRowContext(ctxWithTimeout, sqlStatement, login).Scan(&user.Login, &user.PasswordHash)
	return user, nil
}

// UpdateUser изменить данные пользователя (только если login верный) U
func (su *StorageUsers) UpdateUser(ctx context.Context, user model.Users) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	sqlStatement := `UPDATE users 
				SET u_password = $1
				WHERE u_login = $2;`
	_, err := su.DBconnection.ExecContext(ctxWithTimeout, sqlStatement, user.PasswordHash, user.Login)
	if err != nil {
		logger.Log.Errorln("error while update user data in DB", err)
		return err
	}
	return nil
}

// DeleteUser удалить пользователя и все его данные D
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
	PassStorage.DeleteDataByLogin(ctxWithTimeout, user.Login)

	//TODO: delete in cards table
	//TODO: delete in binary file table

	return nil
}
