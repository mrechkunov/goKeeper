package repository

import (
	"database/sql"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

// добавить данные если таких нет C
// вернуть данные если корректный uuid и metadata R
// изменить данные если передан корректный uuid и metadata U
// удалить данные если корректный uuid и metadata D
