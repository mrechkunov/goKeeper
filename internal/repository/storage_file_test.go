package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/repository"
	"github.com/stretchr/testify/assert"
)

// --- Тесты для метода InsertFile ---

func TestStorageFile_InsertFile_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sf := repository.NewFileStorage(db)
	file := model.File{UserLogin: "user1", FilePath: "/path/to/file", MetaData: "meta"}

	// Ожидаем корректный INSERT запрос
	mock.ExpectExec(`INSERT INTO files \(f_login, f_file_path, f_metadata\)`).
		WithArgs(file.UserLogin, file.FilePath, file.MetaData).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = sf.InsertFile(context.Background(), file)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStorageFile_InsertFile_AlreadyExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sf := repository.NewFileStorage(db)
	file := model.File{UserLogin: "user1", FilePath: "/path/to/file", MetaData: "meta"}

	// Симулируем ошибку нарушения уникальности Postgres (23505)
	pqErr := &pq.Error{Code: "23505", Message: "unique_violation"}
	mock.ExpectExec(`INSERT INTO files`).
		WithArgs(file.UserLogin, file.FilePath, file.MetaData).
		WillReturnError(pqErr)

	err = sf.InsertFile(context.Background(), file)

	// Проверяем, что репозиторий вернул нашу типизированную бизнес-ошибку
	assert.ErrorIs(t, err, repository.ErrDataAlreadyExists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Тесты для метода SelectFile ---

func TestStorageFile_SelectFile_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sf := repository.NewFileStorage(db)
	login, meta := "user1", "meta"

	// Создаем фейковую строку ответа базы данных
	rows := sqlmock.NewRows([]string{"f_login", "f_file_path", "f_metadata"}).
		AddRow("user1", "/path/to/file", "meta")

	mock.ExpectQuery(`SELECT f_login, f_file_path, f_metadata FROM files`).
		WithArgs(login, meta).
		WillReturnRows(rows)

	res, err := sf.SelectFile(context.Background(), login, meta)

	assert.NoError(t, err)
	assert.Equal(t, "user1", res.UserLogin)
	assert.Equal(t, "/path/to/file", res.FilePath)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStorageFile_SelectFile_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sf := repository.NewFileStorage(db)

	// Симулируем отсутствие строк в БД
	mock.ExpectQuery(`SELECT f_login`).
		WithArgs("user1", "meta").
		WillReturnError(sql.ErrNoRows)

	_, err = sf.SelectFile(context.Background(), "user1", "meta")

	assert.ErrorIs(t, err, repository.ErrDataNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Тесты для метода UpdateFile ---

func TestStorageFile_UpdateFile_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sf := repository.NewFileStorage(db)
	file := model.File{UserLogin: "user1", FilePath: "/new/path", MetaData: "meta"}

	// Ожидаем успешный UPDATE, затронувший 1 строку
	mock.ExpectExec(`UPDATE files SET f_file_path = \$1`).
		WithArgs(file.FilePath, file.UserLogin, file.MetaData).
		WillReturnResult(sqlmock.NewResult(1, 1)) // 1 строка обновлена

	err = sf.UpdateFile(context.Background(), file)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStorageFile_UpdateFile_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sf := repository.NewFileStorage(db)
	file := model.File{UserLogin: "missing_user", FilePath: "/path", MetaData: "meta"}

	// Симулируем ситуацию, когда запись для обновления не найдена (RowsAffected = 0)
	mock.ExpectExec(`UPDATE files`).
		WithArgs(file.FilePath, file.UserLogin, file.MetaData).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк затронуто

	err = sf.UpdateFile(context.Background(), file)

	assert.ErrorIs(t, err, repository.ErrDataNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Тесты для метода DeleteAllFilesByLogin ---

func TestStorageFile_DeleteAllFilesByLogin_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sf := repository.NewFileStorage(db)
	login := "user1"

	// 1. Мокаем SELECT путей файлов (внутри GetFilesPathByLogin)
	selectRows := sqlmock.NewRows([]string{"f_file_path"}).
		AddRow("/path/1.txt").
		AddRow("/path/2.txt")
	mock.ExpectQuery(`SELECT f_file_path FROM files WHERE f_login = \$1`).
		WithArgs(login).
		WillReturnRows(selectRows)

	// 2. Мокаем DELETE запрос
	mock.ExpectExec(`DELETE FROM files WHERE f_login = \$1`).
		WithArgs(login).
		WillReturnResult(sqlmock.NewResult(0, 2))

	paths, err := sf.DeleteAllFilesByLogin(context.Background(), login)

	assert.NoError(t, err)
	assert.Len(t, paths, 2)
	assert.Equal(t, []string{"/path/1.txt", "/path/2.txt"}, paths)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStorageFile_GetFilesPathByLogin_DbIterationError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sf := repository.NewFileStorage(db)

	// Симулируем критическую ошибку в процессе итерации rows.Next() (например, обрыв сети)
	rows := sqlmock.NewRows([]string{"f_file_path"}).
		AddRow("/path/1.txt").
		RowError(0, errors.New("postgres connection lost")) // Ошибка посреди чтения

	mock.ExpectQuery(`SELECT f_file_path FROM files`).
		WithArgs("user1").
		WillReturnRows(rows)

	paths, err := sf.GetFilesPathByLogin(context.Background(), "user1")

	// Метод должен поймать ошибку и не возвращать неполный список путей
	assert.Nil(t, paths)
	assert.ErrorContains(t, err, "rows iteration error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 1. Успешный сценарий: запись найдена и успешно удалена из базы данных
func TestStorageFile_DeleteFile_Success(t *testing.T) {
	// Инициализируем sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Создаем экземпляр отрефакторенного репозитория
	sf := repository.NewFileStorage(db)

	// Тестовые данные
	fileData := model.File{
		UserLogin: "alex_pro",
		MetaData:  "user_avatar_2026",
	}

	// Ожидаем вызов DELETE с конкретными аргументами.
	// Метод WillReturnResult симулирует успешное удаление 1 строки (RowsAffected = 1)
	mock.ExpectExec(`DELETE FROM files WHERE f_login = \$1 AND f_metadata = \$2;`).
		WithArgs(fileData.UserLogin, fileData.MetaData).
		WillReturnResult(sqlmock.NewResult(0, 1)) // LastInsertId: 0, RowsAffected: 1

	// Вызов тестируемого метода
	err = sf.DeleteFile(context.Background(), fileData)

	// Проверки (Asserts)
	assert.NoError(t, err)                        // Ошибок быть не должно
	assert.NoError(t, mock.ExpectationsWereMet()) // Все ожидания мока выполнились
}

// 2. Сценарий "Запись не найдена": запрос выполнился, но RowsAffected равен 0
func TestStorageFile_DeleteFile_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sf := repository.NewFileStorage(db)

	fileData := model.File{
		UserLogin: "unknown_user",
		MetaData:  "missing_document",
	}

	// Симулируем ситуацию, когда в БД нет такой записи (RowsAffected = 0)
	mock.ExpectExec(`DELETE FROM files WHERE f_login = \$1 AND f_metadata = \$2;`).
		WithArgs(fileData.UserLogin, fileData.MetaData).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк затронуто

	err = sf.DeleteFile(context.Background(), fileData)

	// Проверяем, что метод вернул правильную доменную ошибку продакшена
	assert.ErrorIs(t, err, repository.ErrDataNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 3. Сценарий ошибки драйвера: база данных вернула системную ошибку при выполнении запроса
func TestStorageFile_DeleteFile_DbError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sf := repository.NewFileStorage(db)

	fileData := model.File{
		UserLogin: "john_doe",
		MetaData:  "private_key",
	}

	// Симулируем падение базы данных (например, ошибку синтаксиса или deadlock)
	dbError := errors.New("postgres: connection refused or deadlocked")
	mock.ExpectExec(`DELETE FROM files`).
		WithArgs(fileData.UserLogin, fileData.MetaData).
		WillReturnError(dbError)

	err = sf.DeleteFile(context.Background(), fileData)

	// Проверяем, что ошибка обернута и содержит исходный текст проблемы СУБД
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exec delete file")
	assert.Contains(t, err.Error(), dbError.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}
