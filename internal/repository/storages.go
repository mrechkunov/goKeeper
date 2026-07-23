package repository

import "github.com/mrechkunov/goKeeper.git/internal/config"

type Repo struct {
	CardStorage     StorageCards
	UserStorage     StorageUsers
	PasswordStorage StoragePasswords
	FileStorage     StorageFile
}

var R = Repo{
	CardStorage:     *NewCardsStorage(config.DBconn),
	UserStorage:     *NewUsersStorage(config.DBconn),
	PasswordStorage: *NewPasswordsStorage(config.DBconn),
	FileStorage:     *NewFileStorage(config.DBconn),
}
