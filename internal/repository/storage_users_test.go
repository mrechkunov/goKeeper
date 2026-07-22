package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/repository"
)

func TestStorageUsers_SelectUser(t *testing.T) {
	// Создаем мок БД
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("не удалось создать мок базы данных: %s", err)
	}
	defer db.Close()
	// Ожидаемые данные из базы
	expectedLogin := "test_user"
	expectedPasswordHash := "test_pass"

	// Готовим ожидание для SQL-запроса
	rows := sqlmock.NewRows([]string{"u_login", "u_password"}).
		AddRow(expectedLogin, expectedPasswordHash)

	mock.ExpectQuery(`SELECT u_login, u_password FROM users WHERE u_login = \$1`).
		WithArgs(expectedLogin).
		WillReturnRows(rows)

	// Вызываем тестируемую функцию, передавая мок БД
	storage := repository.NewUsersStorage(db)
	user, err := storage.SelectUser(context.Background(), expectedLogin)

	// Проверяем результаты
	if err != nil {
		t.Errorf("ошибка не ожидалась, получено: %s", err)
	}
	if user.Login != expectedLogin {
		t.Errorf("ожидалось имя %q, получено %q", expectedLogin, user.Login)
	}

	// Проверяем, все ли ожидания были выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("были невыполненные ожидания: %s", err)
	}

}

func TestStorageUsers_UpdateUser(t *testing.T) {
	// Создаем мок БД
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("не удалось создать мок базы данных: %s", err)
	}
	defer db.Close()

	// Передаваемые данные в запрос к mock БД
	Login := "test_user"
	Password := "test_new_pass"

	mock.ExpectExec(`UPDATE users SET u_password = \$1 WHERE u_login = \$2;`).
		WithArgs(Password, Login).WillReturnResult(sqlmock.NewResult(1, 1)) // 1 affected row

	// данные передаваемые в функцию
	testUserToUpdate := model.Users{
		Login:        "test_user",
		PasswordHash: "test_new_pass",
	}
	// Вызываем тестируемую функцию, передавая мок БД
	storage := repository.NewUsersStorage(db)
	err = storage.UpdateUser(context.Background(), testUserToUpdate)

	// Проверяем результаты
	if err != nil {
		t.Errorf("ошибка не ожидалась, получено: %s", err)
	}
	// Проверяем, все ли ожидания были выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("были невыполненные ожидания: %s", err)
	}
}

func TestStorageUsers_InsertUser(t *testing.T) {
	// Создаем мок БД
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("не удалось создать мок базы данных: %s", err)
	}
	defer db.Close()

	// Подготавливаем тестовые данные
	user := model.Users{
		Login:        "test_user",
		PasswordHash: "test_pass",
	}

	// Задаем ожидания (Expectations) для mock-базы
	// Ожидаем вызов ExecContext с нашим SQL-запросом и аргументами
	query := `INSERT INTO users \(u_login, u_password\) VALUES \(\$1, \$2\)`
	mock.ExpectExec(query).
		WithArgs(user.Login, user.PasswordHash).
		WillReturnResult(sqlmock.NewResult(1, 1)) // Симулируем успешную вставку (LastInsertId=1, AffectedRows=1)

	// Вызываем тестируемую функцию
	storage := repository.NewUsersStorage(db)
	err = storage.InsertUser(context.Background(), user)

	// Проверяем результаты
	if err != nil {
		t.Errorf("ожидалось отсутствие ошибки, но получено: %s", err)
	}

	// Проверяем, что все заданные ожидания были выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ожидания не выполнены: %s", err)
	}
}

func TestStorageUsers_DeleteUser(t *testing.T) {
	// Создаем мок БД
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("не удалось создать мок базы данных: %s", err)
	}
	defer db.Close()
	// Подготавливаем тестовые данные
	user := model.Users{
		Login:        "test_user",
		PasswordHash: "test_pass",
	}
	// Ожидаем SQL-запрос с конкретным параметром (например, ID = 1)
	// и указываем, что он должен затронуть 1 строку
	mock.ExpectExec(`DELETE FROM users WHERE u_login = \$1;`).
		WithArgs(user.Login).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Вызываем тестируемую функцию
	storage := repository.NewUsersStorage(db)
	err = storage.DeleteUser(context.Background(), user)
	// Проверяем результаты
	if err != nil {
		t.Errorf("ожидалось отсутствие ошибки, но получено: %s", err)
	}
	// Проверяем, что все заданные ожидания были выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ожидания не выполнены: %s", err)
	}
}

func TestStorageUsers_IsExist(t *testing.T) {
	// Создаем мок БД
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("не удалось создать мок базы данных: %s", err)
	}
	defer db.Close()
	// Подготавливаем тестовые данные
	user := "test_user"
	// Сценарий 1: Пользователь существует (база возвращает true)
	t.Run("user exists", func(t *testing.T) {
		// Ожидаем запрос с конкретным аргументом
		// Готовим ожидание для SQL-запроса
		rows := sqlmock.NewRows([]string{"u_login", "u_password"}).
			AddRow(user, "test_pass")
		mock.ExpectQuery(`SELECT u_login, u_password FROM users WHERE u_login = \$1;`).WithArgs(user).WillReturnRows(rows)
		// Вызываем функцию
		storage := repository.NewUsersStorage(db)
		exists, err := storage.IsExist(context.Background(), user)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !exists {
			t.Errorf("expected user to exist, but got false")
		}
	})

	// Сценарий 2: Пользователь не существует (база возвращает ошибка sql.NoRows)
	t.Run("user does not exist", func(t *testing.T) {
		mock.ExpectQuery(`SELECT u_login, u_password FROM users WHERE u_login = \$1;`).WithArgs(user).WillReturnError(sql.ErrNoRows)
		// Вызываем функцию
		storage := repository.NewUsersStorage(db)
		exists, err := storage.IsExist(context.Background(), user)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if exists {
			t.Errorf("expected user to not exist, but got true")
		}
	})
	// Проверяем, что все наши ожидания (ExpectQuery) оправдались
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
