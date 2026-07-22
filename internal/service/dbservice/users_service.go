package dbservice

import (
	"context"
	"sync"

	"github.com/mrechkunov/goKeeper.git/internal/config"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Add user in DB if login is not exist
func AddUser(ctx context.Context, user model.Users) error {
	usersStorage := repository.NewUsersStorage(config.DBconn)
	exist, err := usersStorage.IsExist(ctx, user.Login)
	if err != nil || exist {
		err := status.Error(codes.AlreadyExists, "User already exist")
		return err
	} else {
		usersStorage.InsertUser(ctx, user)
		return nil
	}
}

// GetUser return user by login
func GetUser(ctx context.Context, login string) (user model.Users, err error) {
	usersStorage := repository.NewUsersStorage(config.DBconn)
	user, err = usersStorage.SelectUser(ctx, login)
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

// DeleteUser delete user and all data in storages
func DeleteUser(ctx context.Context, user model.Users) (err error) {
	var wg sync.WaitGroup
	wg.Add(3)
	// удалим все данные пользователя ранее созданные
	go func() {
		defer wg.Done()
		PassStorage := repository.NewPasswordsStorage(config.DBconn)
		PassStorage.DeleteAllPasswordsByLogin(ctx, user.Login)
	}()
	go func() {
		defer wg.Done()
		CardStorage := repository.NewCardsStorage(config.DBconn)
		CardStorage.DeleteAllCardsByLogin(ctx, user.Login)
	}()
	go func() {
		defer wg.Done()
		// удалим записи из БД
		DeleteAllUserFiles(ctx, user.Login)
	}()
	wg.Wait()
	usersStorage := repository.NewUsersStorage(config.DBconn)
	return usersStorage.DeleteUser(ctx, user)
}
