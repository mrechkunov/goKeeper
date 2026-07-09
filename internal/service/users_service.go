package service

import (
	"context"

	"github.com/mrechkunov/goKeeper.git/internal/config"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Insert NEW user in DB if login is not exist
func InsertUser(ctx context.Context, user model.Users) error {
	storage := repository.NewUsersStorage(config.DBconn)
	defer storage.Close()
	if storage.IsExist(ctx, user.Login) {
		err := status.Error(codes.AlreadyExists, "User already exist")
		return err
	}
	storage.CreateUser(ctx, user)
	return nil
}

// TODO: autorization user (token check) extend

// TODO: authentification user (login + pass check, responce token)
// TODO: edit user (change password for autorizated user)
// TODO: delete user (delete user and all data in torages)

// func GetUserByLogin(ctx context.Context, login string) model.Users {
// 	storageUsers := repository.NewUsersStorage(config.DBconn)
// 	return storageUsers.GetUserByLogin(ctx, token)
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
