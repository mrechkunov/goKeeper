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
	usersStorage := repository.NewUsersStorage(config.DBconn)
	exist, err := usersStorage.IsExist(ctx, user.Login)
	if err != nil || exist {
		err := status.Error(codes.AlreadyExists, "User already exist")
		return err
	} else {
		usersStorage.CreateUser(ctx, user)
		return nil
	}
}

// GetUserByLogin return user by login
func GetUserByLogin(ctx context.Context, login string) (user model.Users, err error) {
	usersStorage := repository.NewUsersStorage(config.DBconn)
	user, err = usersStorage.ReadUser(ctx, login)
	if err != nil {
		return user, err
	}
	return user, nil
}

// EditUser change password for autorizated user
func EditUser(ctx context.Context, user model.Users) (err error) {
	usersStorage := repository.NewUsersStorage(config.DBconn)
	return usersStorage.UpdateUser(ctx, user)
}

// TODO: delete user (delete user and all data in storages)
func UserDelete(ctx context.Context, user model.Users) (err error) {
	usersStorage := repository.NewUsersStorage(config.DBconn)
	return usersStorage.DeleteUser(ctx, user)
}
