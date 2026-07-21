package db

import (
	"context"

	"github.com/mrechkunov/goKeeper.git/internal/config"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/repository"
)

// AddPassword insert data in storage
func AddPassword(ctx context.Context, data model.Passwords) error {
	passStorage := repository.NewPasswordsStorage(config.DBconn)
	return passStorage.InsertPassword(ctx, data)
}

// GetPassword return data from storage selected by login & metadata
func GetPassword(ctx context.Context, login, metadata string) (data model.Passwords, err error) {
	passStorage := repository.NewPasswordsStorage(config.DBconn)
	return passStorage.SelectPassword(ctx, login, metadata)
}

// EditPassword
func EditPassword(ctx context.Context, dataIn model.Passwords) error {
	passStorage := repository.NewPasswordsStorage(config.DBconn)
	return passStorage.UpdatePassword(ctx, dataIn)
}

// DeletePassword delete row with password by login and metadata
func DeletePassword(ctx context.Context, data model.Passwords) error {
	passStorage := repository.NewPasswordsStorage(config.DBconn)
	return passStorage.DeletePassword(ctx, data)
}

// DeleteAllUserPasswords delete all records by login
func DeleteAllUserPasswords(ctx context.Context, login string) error {
	passStorage := repository.NewPasswordsStorage(config.DBconn)
	return passStorage.DeleteAllPasswordsByLogin(ctx, login)
}
