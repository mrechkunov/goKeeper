package dbservice

import (
	"context"
	"fmt"

	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/repository"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Add user in DB if login is not exist
var AddUser = func(ctx context.Context, user model.Users) error {
	exist, err := repository.R.UserStorage.IsExist(ctx, user.Login)
	if err != nil || exist {
		err := status.Error(codes.AlreadyExists, "User already exist")
		return err
	} else {
		repository.R.UserStorage.InsertUser(ctx, user)
		return nil
	}
}

// GetUser return user by login
var GetUser = func(ctx context.Context, login string) (user model.Users, err error) {
	user, err = repository.R.UserStorage.SelectUser(ctx, login)
	if err != nil {
		return user, err
	}
	return user, nil
}

// EditUser change password for autorizated user
var EditUser = func(ctx context.Context, user model.Users) (err error) {
	return repository.R.UserStorage.UpdateUser(ctx, user)
}

// DeleteUser delete user and all data in storages
var DeleteUser = func(ctx context.Context, user model.Users) error {
	// создаем errgroup с контекстом.
	// если одна из горутин вернет ошибку, контекст g.WithContext автоматически отменится.
	g, ctx := errgroup.WithContext(ctx)

	// Удаление паролей
	g.Go(func() error {

		// новая переменная ошибки
		if err := repository.R.PasswordStorage.DeleteAllPasswordsByLogin(ctx, user.Login); err != nil {
			logger.Log.Warnln("error while deleting user passwords data:", err)
			return fmt.Errorf("delete passwords: %w", err)
		}
		return nil
	})

	// Удаление карт
	g.Go(func() error {

		// новая переменная ошибки
		if err := repository.R.CardStorage.DeleteAllCardsByLogin(ctx, user.Login); err != nil {
			logger.Log.Warnln("error while deleting user card data:", err)
			return fmt.Errorf("delete cards: %w", err)
		}
		return nil
	})

	// Удаление файлов
	g.Go(func() error {
		// новая переменная ошибки
		if err := DeleteAllUserFiles(ctx, user.Login); err != nil {
			logger.Log.Warnln("error while deleting user files:", err)
			return fmt.Errorf("delete files: %w", err)
		}
		return nil
	})

	// Ожидаем завершения всех горутин. Возвращается первая возникшая ошибка.
	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to clear user data: %w", err)
	}

	// Удаление самого пользователя из БД (выполняется только если все горутины завершились без ошибок)
	if err := repository.R.UserStorage.DeleteUser(ctx, user); err != nil {
		return fmt.Errorf("failed to delete user from storage: %w", err)
	}

	return nil
}
