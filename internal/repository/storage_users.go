package repository

import (
	"database/sql"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type StorageUsers struct {
	DBconnection *sql.DB
}

// создаем новый сторадж для работы с таблицей пользователей
func NewUsersStorage(DBconn *sql.DB) StorageUsers {
	return StorageUsers{DBconnection: DBconn}
}

// Close DB connection
func (d *StorageUsers) Close() error {
	return d.DBconnection.Close()
}

// добавить пользователя C
// авторизовать пользователя (проверить что он есть и вернуть uuid) R
// изменить данные пользователя (только если uuid верный) U
// удалить пользователя и все его данные D
