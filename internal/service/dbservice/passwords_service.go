package dbservice

import (
	"context"

	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/repository"
)

// AddPassword insert data in storage
var AddPassword = func(ctx context.Context, data model.Passwords) error {
	return repository.R.PasswordStorage.InsertPassword(ctx, data)
}

// GetPassword return data from storage selected by login & metadata
var GetPassword = func(ctx context.Context, login, metadata string) (data model.Passwords, err error) {
	return repository.R.PasswordStorage.SelectPassword(ctx, login, metadata)
}

// EditPassword
var EditPassword = func(ctx context.Context, dataIn model.Passwords) error {
	return repository.R.PasswordStorage.UpdatePassword(ctx, dataIn)
}

// DeletePassword delete row with password by login and metadata
var DeletePassword = func(ctx context.Context, data model.Passwords) error {
	return repository.R.PasswordStorage.DeletePassword(ctx, data)
}

// DeleteAllUserPasswords delete all records by login
var DeleteAllUserPasswords = func(ctx context.Context, login string) error {
	return repository.R.PasswordStorage.DeleteAllPasswordsByLogin(ctx, login)
}
