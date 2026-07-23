package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/repository"
)

func TestStoragePasswords_isExist(t *testing.T) {
	// Создаем mock-соединение с БД
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Инициализируем хранилище с моком вместо реальной БД
	sp := repository.NewPasswordsStorage(db)

	// Тестовые данные
	testData := model.Passwords{
		UserLogin: "test_user",
		MetaData:  "some_metadata",
	}

	// Экранируем SQL-запрос для регулярного выражения sqlmock
	expectedSQL := `SELECT p_login, p_pair, p_metadata FROM passwords WHERE p_metadata = \$1 AND p_login = \$2;`

	tests := []struct {
		name    string
		mockAct func()
		want    bool
	}{
		{
			name: "Запись существует в БД",
			mockAct: func() {
				// Создаем фейковую строку ответа
				rows := sqlmock.NewRows([]string{"p_login", "p_pair", "p_metadata"}).
					AddRow("test_user", "encrypted_pair", "some_metadata")

				// Ожидаем вызов запроса с конкретными аргументами и возвращаем строку
				mock.ExpectQuery(expectedSQL).
					WithArgs(testData.MetaData, testData.UserLogin).
					WillReturnRows(rows)
			},
			want: true,
		},
		{
			name: "Запись отсутствует в БД (sql.ErrNoRows)",
			mockAct: func() {
				// Ожидаем вызов запроса, возвращаем ошибку "нет строк"
				mock.ExpectQuery(expectedSQL).
					WithArgs(testData.MetaData, testData.UserLogin).
					WillReturnError(sql.ErrNoRows)
			},
			want: false,
		},
		{
			name: "Критическая ошибка БД",
			mockAct: func() {
				// Симулируем другую ошибку (например, разрыв соединения)
				mock.ExpectQuery(expectedSQL).
					WithArgs(testData.MetaData, testData.UserLogin).
					WillReturnError(errors.New("connection timeout"))
			},

			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем поведение мока для текущего кейса
			tt.mockAct()

			// Вызываем тестируемый метод
			got := sp.IsExist(context.Background(), testData)

			if got != tt.want {
				t.Errorf("isExist() = %v, want %v", got, tt.want)
			}

			// Проверяем, что все ожидаемые запросы к моку были выполнены
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStoragePasswords_InsertPassword(t *testing.T) {
	// Инициализируем sqlmock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sp := repository.NewPasswordsStorage(db)

	testData := model.Passwords{
		UserLogin: "test_user",
		Pair:      "encrypted_password_pair",
		MetaData:  "site_metadata",
	}

	// Экранируем регулярные выражения для SQL-запросов
	existSQL := `SELECT p_login, p_pair, p_metadata FROM passwords WHERE p_metadata = \$1 AND p_login = \$2;`
	insertSQL := `INSERT INTO passwords \(p_login, p_pair, p_metadata\) VALUES \(\$1, \$2, \$3\)`

	tests := []struct {
		name    string
		mockAct func()
		wantErr bool
		errText string
	}{
		{
			name: "Успешная вставка (данных еще нет в БД)",
			mockAct: func() {
				// 1. Симулируем, что IsExist возвращает false (нет строк)
				mock.ExpectQuery(existSQL).
					WithArgs(testData.MetaData, testData.UserLogin).
					WillReturnError(sql.ErrNoRows)

				// 2. Симулируем успешное выполнение INSERT
				mock.ExpectExec(insertSQL).
					WithArgs(testData.UserLogin, testData.Pair, testData.MetaData).
					WillReturnResult(sqlmock.NewResult(1, 1)) // LastInsertId = 1, RowsAffected = 1
			},
			wantErr: false,
		},
		{
			name: "Ошибка: данные уже существуют",
			mockAct: func() {
				// Симулируем, что IsExist возвращает true (находит строку)
				rows := sqlmock.NewRows([]string{"p_login", "p_pair", "p_metadata"}).
					AddRow("test_user", "old_pair", "site_metadata")

				mock.ExpectQuery(existSQL).
					WithArgs(testData.MetaData, testData.UserLogin).
					WillReturnRows(rows)

				// INSERT не должен вызываться, так как функция сделает ранний возврат
			},
			wantErr: true,
			errText: "data is already exists in DB",
		},
		{
			name: "Ошибка при выполнении INSERT",
			mockAct: func() {
				// 1. IsExist возвращает false (данных нет)
				mock.ExpectQuery(existSQL).
					WithArgs(testData.MetaData, testData.UserLogin).
					WillReturnError(sql.ErrNoRows)

				// 2. Симулируем сбой самой базы данных на этапе вставки (например, ошибка уникальности или связи)
				mock.ExpectExec(insertSQL).
					WithArgs(testData.UserLogin, testData.Pair, testData.MetaData).
					WillReturnError(errors.New("db insert failure"))
			},
			wantErr: true,
			errText: "db insert failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем моки под текущий сценарий
			tt.mockAct()

			// Вызываем тестируемый метод
			err := sp.InsertPassword(context.Background(), testData)

			// Проверяем наличие ошибки
			if (err != nil) != tt.wantErr {
				t.Errorf("InsertPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Если ошибка ожидалась, проверяем её текст
			if tt.wantErr && err != nil && err.Error() != tt.errText {
				t.Errorf("InsertPassword() error text = %q, want %q", err.Error(), tt.errText)
			}

			// Проверяем, что все запланированные ожидания sqlmock сработали
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStoragePasswords_SelectPassword(t *testing.T) {
	// Инициализируем sqlmock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sp := repository.NewPasswordsStorage(db)

	// Тестовые входные параметры
	testLogin := "active_user"
	testMetadata := "github_account"

	// Ожидаемый SQL-запрос (экранируем спецсимволы для regex)
	expectedSQL := `SELECT p_login, p_pair, p_metadata FROM passwords WHERE p_login = \$1 AND p_metadata = \$2;`

	tests := []struct {
		name        string
		mockAct     func()
		wantResp    model.Passwords
		wantErr     error
		wantErrType bool
	}{
		{
			name: "Успешный выбор данных",
			mockAct: func() {
				// Создаем фейковую строку ответа СУБД
				rows := sqlmock.NewRows([]string{"p_login", "p_pair", "p_metadata"}).
					AddRow(testLogin, "encrypted_pass_123", testMetadata)

				mock.ExpectQuery(expectedSQL).
					WithArgs(testLogin, testMetadata).
					WillReturnRows(rows)
			},
			wantResp: model.Passwords{
				UserLogin: testLogin,
				Pair:      "encrypted_pass_123",
				MetaData:  "github_account",
			},
			wantErr: nil,
		},
		{
			name: "Запись не найдена (sql.ErrNoRows)",
			mockAct: func() {
				mock.ExpectQuery(expectedSQL).
					WithArgs(testLogin, testMetadata).
					WillReturnError(sql.ErrNoRows)
			},
			wantResp: model.Passwords{}, // Пустая структура
			wantErr:  sql.ErrNoRows,
		},
		{
			name: "Внутренняя ошибка базы данных (например, Connection Timeout)",
			mockAct: func() {
				mock.ExpectQuery(expectedSQL).
					WithArgs(testLogin, testMetadata).
					WillReturnError(errors.New("db failure"))
			},
			wantResp: model.Passwords{},
			wantErr:  errors.New("db failure"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем поведение мока для текущего шага
			tt.mockAct()

			// Вызываем тестируемый метод
			resp, err := sp.SelectPassword(context.Background(), testLogin, testMetadata)

			// Проверяем ошибку
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("SelectPassword() expected error: %v, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr.Error() {
					t.Errorf("SelectPassword() error = %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("SelectPassword() unexpected error: %v", err)
			}

			// Проверяем возвращаемые данные
			if resp.UserLogin != tt.wantResp.UserLogin ||
				resp.Pair != tt.wantResp.Pair ||
				resp.MetaData != tt.wantResp.MetaData {
				t.Errorf("SelectPassword() resp = %+v, want %+v", resp, tt.wantResp)
			}

			// Проверяем, что все ожидания sqlmock были выполнены
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStoragePasswords_UpdatePassword_Optimized(t *testing.T) {
	// Инициализируем sqlmock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sp := repository.NewPasswordsStorage(db)

	// Тестовые данные
	testData := model.Passwords{
		UserLogin: "existing_user",
		Pair:      "new_encrypted_pair",
		MetaData:  "service_metadata",
	}

	// Экранируем регулярное выражение для SQL-запроса UPDATE
	updateSQL := `UPDATE passwords SET p_pair = \$1 WHERE p_login = \$2 AND p_metadata = \$3;`

	tests := []struct {
		name    string
		mockAct func()
		wantErr bool
		errText string
	}{
		{
			name: "Успешное обновление (запись существует)",
			mockAct: func() {
				// Возвращаем результат, где RowsAffected = 1
				mock.ExpectExec(updateSQL).
					WithArgs(testData.Pair, testData.UserLogin, testData.MetaData).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "Ошибка: данные не найдены (RowsAffected = 0)",
			mockAct: func() {
				// Запрос выполнился успешно, но ни одна строка не изменилась (RowsAffected = 0)
				mock.ExpectExec(updateSQL).
					WithArgs(testData.Pair, testData.UserLogin, testData.MetaData).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
			errText: "data is not exists in DB",
		},
		{
			name: "Ошибка уровня СУБД (например, синтаксис или разрыв соединения)",
			mockAct: func() {
				// Симулируем жесткую ошибку базы данных
				mock.ExpectExec(updateSQL).
					WithArgs(testData.Pair, testData.UserLogin, testData.MetaData).
					WillReturnError(errors.New("db connection lost"))
			},
			wantErr: true,
			errText: "db connection lost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем поведение мока для текущего сценария
			tt.mockAct()

			// Вызываем оптимизированный метод
			err := sp.UpdatePassword(context.Background(), testData)

			// Проверяем наличие ошибки
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Проверяем текст ошибки, если она ожидалась
			if tt.wantErr && err != nil && err.Error() != tt.errText {
				t.Errorf("UpdatePassword() error text = %q, want %q", err.Error(), tt.errText)
			}

			// Проверяем, что мок зафиксировал все ожидаемые вызовы
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStoragePasswords_DeletePassword_Optimized(t *testing.T) {
	// Инициализируем sqlmock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sp := repository.NewPasswordsStorage(db)

	// Тестовые данные
	testData := model.Passwords{
		UserLogin: "user_to_delete",
		MetaData:  "legacy_api_key",
	}

	// Экранируем регулярное выражение для SQL-запроса DELETE
	deleteSQL := `DELETE FROM passwords WHERE p_login = \$1 AND p_metadata = \$2;`

	tests := []struct {
		name    string
		mockAct func()
		wantErr bool
		errText string
	}{
		{
			name: "Успешное удаление (запись существовала)",
			mockAct: func() {
				// Запрос затронул 1 строку (успешное удаление)
				mock.ExpectExec(deleteSQL).
					WithArgs(testData.UserLogin, testData.MetaData).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "Ошибка: запись для удаления не найдена (RowsAffected = 0)",
			mockAct: func() {
				// Запрос выполнился, но удалять было нечего
				mock.ExpectExec(deleteSQL).
					WithArgs(testData.UserLogin, testData.MetaData).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
			errText: "data is not exists in DB",
		},
		{
			name: "Ошибка уровня базы данных (например, блокировка таблицы)",
			mockAct: func() {
				// Симулируем ошибку выполнения SQL-запроса
				mock.ExpectExec(deleteSQL).
					WithArgs(testData.UserLogin, testData.MetaData).
					WillReturnError(errors.New("db deadlock detected"))
			},
			wantErr: true,
			errText: "db deadlock detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Активируем сценарий мока
			tt.mockAct()

			// Вызываем оптимизированный метод удаления
			err := sp.DeletePassword(context.Background(), testData)

			// Проверяем наличие ошибки
			if (err != nil) != tt.wantErr {
				t.Errorf("DeletePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Проверяем текст ошибки
			if tt.wantErr && err != nil && err.Error() != tt.errText {
				t.Errorf("DeletePassword() error text = %q, want %q", err.Error(), tt.errText)
			}

			// Проверяем, что все запланированные ожидания sqlmock сработали
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStoragePasswords_DeleteAllPasswordsByLogin(t *testing.T) {
	// Инициализируем sqlmock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sp := repository.NewPasswordsStorage(db)

	testLogin := "user_to_wipe"

	// Экранируем регулярное выражение для SQL-запроса DELETE
	deleteSQL := `DELETE FROM passwords WHERE p_login = \$1;`

	tests := []struct {
		name    string
		mockAct func()
		wantErr bool
		errText string
	}{
		{
			name: "Успешное удаление нескольких записей пользователя",
			mockAct: func() {
				// Симулируем удаление, например, 5 паролей пользователя
				mock.ExpectExec(deleteSQL).
					WithArgs(testLogin).
					WillReturnResult(sqlmock.NewResult(0, 5)) // RowsAffected = 5
			},
			wantErr: false,
		},
		{
			name: "Успешное выполнение, когда у пользователя нет сохраненных паролей",
			mockAct: func() {
				// Запрос выполнен, удалено 0 строк. Ошибки быть не должно.
				mock.ExpectExec(deleteSQL).
					WithArgs(testLogin).
					WillReturnResult(sqlmock.NewResult(0, 0)) // RowsAffected = 0
			},
			wantErr: false,
		},
		{
			name: "Ошибка уровня базы данных (например, таймаут или блокировка)",
			mockAct: func() {
				// Симулируем внутреннюю ошибку СУБД
				mock.ExpectExec(deleteSQL).
					WithArgs(testLogin).
					WillReturnError(errors.New("db transaction timeout"))
			},
			wantErr: true,
			errText: "db transaction timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем поведение мока для текущего сценария
			tt.mockAct()

			// Вызываем тестируемый метод
			err := sp.DeleteAllPasswordsByLogin(context.Background(), testLogin)

			// Проверяем наличие ошибки
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteAllPasswordsByLogin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Проверяем текст ошибки, если она ожидалась
			if tt.wantErr && err != nil && err.Error() != tt.errText {
				t.Errorf("DeleteAllPasswordsByLogin() error text = %q, want %q", err.Error(), tt.errText)
			}

			// Проверяем, что все запланированные ожидания sqlmock сработали
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
