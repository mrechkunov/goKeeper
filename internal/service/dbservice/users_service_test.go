package dbdbservice_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mrechkunov/goKeeper.git/internal/config"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/service/db"
	"github.com/stretchr/testify/assert"
)

func TestDeleteUser(t *testing.T) {
	// Создаем мок БД
	dbmock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("не удалось создать мок базы данных: %s", err)
	}
	defer dbmock.Close()
	// меняем коннект к бд
	config.DBconn = dbmock

	ctx := context.Background()
	user := model.Users{Login: "test_user"}

	// Настраиваем ожидания для SQL-запросов, которые уходят из репозиториев.
	mock.ExpectExec(`DELETE FROM passwords WHERE login = \$1;`).
		WithArgs(user.Login).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`DELETE FROM cards WHERE login = \$1`).
		WithArgs(user.Login).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`DELETE FROM files WHERE login = \$1`).
		WithArgs(user.Login).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Запрос на удаление самого юзера выполнится в конце
	mock.ExpectExec(`DELETE FROM users WHERE login = \$1`).
		WithArgs(user.Login).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Вызов тестируемой функции
	err = db.DeleteUser(ctx, user)

	// Проверки
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet()) // Проверяем, что все sql-запросы выполнились
}
