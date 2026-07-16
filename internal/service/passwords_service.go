package service

import (
	"context"

	"github.com/mrechkunov/goKeeper.git/internal/config"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/repository"
)

// AddData insert data in storage
func AddData(ctx context.Context, data model.Passwords) error {
	passStorage := repository.NewPasswordsStorage(config.DBconn)
	return passStorage.InsertData(ctx, data)
}

// TODO: read record (uuid + metadata)

// TODO: edit record (model.password)
// TODO: delete record (uuid + metadata)
// TODO: delete records (uuid)

// func GetUserByLogin(ctx context.Context, login string) model.Users {
// 	storageUsers := repository.NewUsersStorage(config.DBconn)
// 	return storageUsers.GetUserByLogin(ctx, uuid)
// }

// func GetUserByLogin(ctx context.Context, login string) model.Users {
// 	storageUsers := repository.NewUsersStorage(config.DBconn)
// 	return storageUsers.GetUserByLogin(ctx, login)
// }

// func UpdateUser(ctx context.Context, user *model.Users) error {
// 	storageUsers := repository.NewUsersStorage(config.DBconn)
// 	return storageUsers.UpdateUser(ctx, *user)
// }

// func InsertNewUser(ctx context.Context, user *model.Users) error {
// 	storageUsers := repository.NewUsersStorage(config.DBconn)
// 	if storageUsers.InsertUser(ctx, *user) != nil {
// 		return storageUsers.InsertUser(ctx, *user)
// 	}
// 	storageBalance := repository.NewBalanceStorage(config.DBconn)
// 	if storageBalance.AddUserBalance(ctx, user.Login) != nil {
// 		return storageBalance.AddUserBalance(ctx, user.Login)
// 	}
// 	return nil
// }
