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

func TestStorageCards_isExist(t *testing.T) {
	// Инициализируем sqlmock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sc := repository.NewCardsStorage(db)

	// Тестовые данные для карты
	testData := model.Cards{
		UserLogin:  "card_holder_user",
		CipherData: "encrypted_card_fields",
		MetaData:   "personal_visa",
	}

	// Экранируем регулярное выражение для SQL-запроса SELECT
	expectedSQL := `SELECT c_login, c_ciperdata, c_metadata FROM cards WHERE c_metadata = \$1 AND c_login = \$2;`

	tests := []struct {
		name        string
		mockAct     func()
		wantExist   bool
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Карта успешно найдена в БД",
			mockAct: func() {
				rows := sqlmock.NewRows([]string{"c_login", "c_chiperdata", "c_metadata"}).
					AddRow(testData.UserLogin, testData.CipherData, testData.MetaData)

				mock.ExpectQuery(expectedSQL).
					WithArgs(testData.MetaData, testData.UserLogin).
					WillReturnRows(rows)
			},
			wantExist:   true,
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "Карта отсутствует в БД (sql.ErrNoRows)",
			mockAct: func() {
				mock.ExpectQuery(expectedSQL).
					WithArgs(testData.MetaData, testData.UserLogin).
					WillReturnError(sql.ErrNoRows)
			},
			wantExist:   false,
			wantErr:     false, // Для вызывающего кода это штатная ситуация, а не ошибка
			expectedErr: nil,
		},
		{
			name: "Ошибка БД (сетевой сбой или падение Postgres)",
			mockAct: func() {
				mock.ExpectQuery(expectedSQL).
					WithArgs(testData.MetaData, testData.UserLogin).
					WillReturnError(errors.New("connection refused"))
			},
			wantExist:   false,
			wantErr:     true,
			expectedErr: errors.New("connection refused"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем поведение мока под текущий кейс
			tt.mockAct()

			// Вызываем исправленный метод с новой сигнатурой
			gotExist, err := sc.IsExist(context.Background(), testData)

			// 1. Проверяем флаг существования данных
			if gotExist != tt.wantExist {
				t.Errorf("isExist() gotExist = %v, wantExist %v", gotExist, tt.wantExist)
			}

			// 2. Проверяем наличие ошибки в соответствии с ожиданиями
			if (err != nil) != tt.wantErr {
				t.Errorf("isExist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 3. Если ошибка ожидалась, проверяем её текст
			if tt.wantErr && err != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("isExist() error text = %q, want %q", err.Error(), tt.expectedErr.Error())
			}

			// Проверяем, что все запланированные ожидания sqlmock сработали
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStorageCards_InsertCard(t *testing.T) {
	// Инициализируем sqlmock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sc := repository.NewCardsStorage(db)

	// Тестовые данные для карты
	testData := model.Cards{
		UserLogin:  "cardholder_login",
		CipherData: "encrypted_card_info",
		MetaData:   "shopping_card",
	}

	// Экранируем SQL-запрос для регулярного выражения
	insertSQL := `INSERT INTO cards \(c_login, c_cipherdata, c_metadata\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(c_login, c_metadata\) DO NOTHING;`

	tests := []struct {
		name    string
		mockAct func()
		wantErr bool
		errText string
	}{
		{
			name: "Успешное добавление карты (данных нет в БД)",
			mockAct: func() {
				// Возвращаем результат, где RowsAffected = 1
				mock.ExpectExec(insertSQL).
					WithArgs(testData.UserLogin, testData.CipherData, testData.MetaData).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "Ошибка: карта уже существует в БД (RowsAffected = 0)",
			mockAct: func() {
				// Запрос выполнился без ошибок СУБД, но RowsAffected = 0 (сработал ON CONFLICT)
				mock.ExpectExec(insertSQL).
					WithArgs(testData.UserLogin, testData.CipherData, testData.MetaData).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
			errText: "data is already exists in DB",
		},
		{
			name: "Ошибка уровня самой базы данных",
			mockAct: func() {
				// Симулируем падение соединения с базой данных
				mock.ExpectExec(insertSQL).
					WithArgs(testData.UserLogin, testData.CipherData, testData.MetaData).
					WillReturnError(errors.New("fatal: db connection closed"))
			},
			wantErr: true,
			errText: "fatal: db connection closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем мок под текущий сценарий
			tt.mockAct()

			// Вызываем оптимизированный метод
			err := sc.InsertCard(context.Background(), testData)

			// Проверяем наличие ошибки
			if (err != nil) != tt.wantErr {
				t.Errorf("InsertCard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Проверяем текст ошибки
			if tt.wantErr && err != nil && err.Error() != tt.errText {
				t.Errorf("InsertCard() error text = %q, want %q", err.Error(), tt.errText)
			}

			// Убеждаемся, что все ожидания мока выполнились
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStorageCards_SelectCard(t *testing.T) {
	// Инициализируем sqlmock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sc := repository.NewCardsStorage(db)

	// Тестовые входные параметры
	testLogin := "card_holder"
	testMetadata := "main_mastercard"

	// Экранируем регулярное выражение для SQL-запроса SELECT
	expectedSQL := `SELECT c_login, c_cipherdata, c_metadata FROM cards WHERE c_login = \$1 AND c_metadata = \$2;`

	tests := []struct {
		name     string
		mockAct  func()
		wantResp model.Cards
		wantErr  error
	}{
		{
			name: "Успешное получение данных карты",
			mockAct: func() {
				// Создаем фейковую строку ответа СУБД
				rows := sqlmock.NewRows([]string{"c_login", "c_cipherdata", "c_metadata"}).
					AddRow(testLogin, "encrypted_card_bytes_abc", testMetadata)

				mock.ExpectQuery(expectedSQL).
					WithArgs(testLogin, testMetadata).
					WillReturnRows(rows)
			},
			wantResp: model.Cards{
				UserLogin:  testLogin,
				CipherData: "encrypted_card_bytes_abc",
				MetaData:   testMetadata,
			},
			wantErr: nil,
		},
		{
			name: "Карта не найдена в БД (sql.ErrNoRows)",
			mockAct: func() {
				mock.ExpectQuery(expectedSQL).
					WithArgs(testLogin, testMetadata).
					WillReturnError(sql.ErrNoRows)
			},
			wantResp: model.Cards{},
			wantErr:  sql.ErrNoRows,
		},
		{
			name: "Критическая ошибка базы данных (например, драйвера СУБД)",
			mockAct: func() {
				mock.ExpectQuery(expectedSQL).
					WithArgs(testLogin, testMetadata).
					WillReturnError(errors.New("driver: bad connection"))
			},
			wantResp: model.Cards{},
			wantErr:  errors.New("driver: bad connection"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем мок под текущий сценарий
			tt.mockAct()

			// Вызываем исправленный метод
			resp, err := sc.SelectCard(context.Background(), testLogin, testMetadata)

			// 1. Проверяем ошибку
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("SelectCard() expected error: %v, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr.Error() {
					t.Errorf("SelectCard() error = %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("SelectCard() unexpected error: %v", err)
			}

			// 2. Проверяем возвращаемые поля структуры
			if resp.UserLogin != tt.wantResp.UserLogin ||
				resp.CipherData != tt.wantResp.CipherData ||
				resp.MetaData != tt.wantResp.MetaData {
				t.Errorf("SelectCard() resp = %+v, want %+v", resp, tt.wantResp)
			}

			// 3. Проверяем выполнение всех ожиданий мока
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStorageCards_UpdateCard(t *testing.T) {
	// Инициализируем sqlmock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sc := repository.NewCardsStorage(db)

	// Тестовые данные для обновления
	testData := model.Cards{
		UserLogin:  "existing_holder",
		CipherData: "new_encrypted_card_data",
		MetaData:   "salary_card",
	}

	// Экранируем регулярное выражение для SQL-запроса UPDATE
	updateSQL := `UPDATE cards SET c_cipherdata = \$1 WHERE c_login = \$2 AND c_metadata = \$3;`

	tests := []struct {
		name    string
		mockAct func()
		wantErr bool
		errText string
	}{
		{
			name: "Успешное обновление карты (запись существует)",
			mockAct: func() {
				// Симулируем успешный апдейт 1 строки
				mock.ExpectExec(updateSQL).
					WithArgs(testData.CipherData, testData.UserLogin, testData.MetaData).
					WillReturnResult(sqlmock.NewResult(0, 1)) // RowsAffected = 1
			},
			wantErr: false,
		},
		{
			name: "Ошибка: карта не найдена в БД (RowsAffected = 0)",
			mockAct: func() {
				// Запрос выполнился, но обновлять было нечего
				mock.ExpectExec(updateSQL).
					WithArgs(testData.CipherData, testData.UserLogin, testData.MetaData).
					WillReturnResult(sqlmock.NewResult(0, 0)) // RowsAffected = 0
			},
			wantErr: true,
			errText: "data is not exists in DB",
		},
		{
			name: "Ошибка уровня самой базы данных",
			mockAct: func() {
				// Симулируем жесткий сбой СУБД
				mock.ExpectExec(updateSQL).
					WithArgs(testData.CipherData, testData.UserLogin, testData.MetaData).
					WillReturnError(errors.New("postgres: write-ahead log error"))
			},
			wantErr: true,
			errText: "postgres: write-ahead log error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем поведение мока под текущий сценарий
			tt.mockAct()

			// Вызываем оптимизированный метод
			err := sc.UpdateCard(context.Background(), testData)

			// Проверяем наличие ошибки
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateCard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Проверяем текст ошибки
			if tt.wantErr && err != nil && err.Error() != tt.errText {
				t.Errorf("UpdateCard() error text = %q, want %q", err.Error(), tt.errText)
			}

			// Проверяем выполнение всех ожиданий мока
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStorageCards_DeleteCard(t *testing.T) {
	// Инициализируем sqlmock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sc := repository.NewCardsStorage(db)

	// Тестовые данные для удаления
	testData := model.Cards{
		UserLogin: "cardholder_to_delete",
		MetaData:  "expired_visa",
	}

	// Экранируем регулярное выражение для SQL-запроса DELETE
	deleteSQL := `DELETE FROM cards WHERE c_login = \$1 AND c_metadata = \$2;`

	tests := []struct {
		name    string
		mockAct func()
		wantErr bool
		errText string
	}{
		{
			name: "Успешное удаление карты (запись существовала)",
			mockAct: func() {
				// Симулируем успешное удаление 1 строки
				mock.ExpectExec(deleteSQL).
					WithArgs(testData.UserLogin, testData.MetaData).
					WillReturnResult(sqlmock.NewResult(0, 1)) // RowsAffected = 1
			},
			wantErr: false,
		},
		{
			name: "Ошибка: карта для удаления не найдена в БД (RowsAffected = 0)",
			mockAct: func() {
				// Запрос выполнился успешно, но удалять было нечего
				mock.ExpectExec(deleteSQL).
					WithArgs(testData.UserLogin, testData.MetaData).
					WillReturnResult(sqlmock.NewResult(0, 0)) // RowsAffected = 0
			},
			wantErr: true,
			errText: "data is not exists in DB",
		},
		{
			name: "Ошибка уровня базы данных (например, deadlock или сбой диска)",
			mockAct: func() {
				// Симулируем жесткую ошибку СУБД
				mock.ExpectExec(deleteSQL).
					WithArgs(testData.UserLogin, testData.MetaData).
					WillReturnError(errors.New("postgres: table is locked"))
			},
			wantErr: true,
			errText: "postgres: table is locked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем поведение мока под текущий сценарий
			tt.mockAct()

			// Вызываем оптимизированный метод
			err := sc.DeleteCard(context.Background(), testData)

			// Проверяем наличие ошибки
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteCard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Проверяем текст ошибки
			if tt.wantErr && err != nil && err.Error() != tt.errText {
				t.Errorf("DeleteCard() error text = %q, want %q", err.Error(), tt.errText)
			}

			// Проверяем выполнение всех ожиданий мока
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStorageCards_DeleteAllCardsByLogin(t *testing.T) {
	// Инициализируем sqlmock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sc := repository.NewCardsStorage(db)

	testLogin := "user_to_clear"

	// Экранируем регулярное выражение для SQL-запроса DELETE
	deleteSQL := `DELETE FROM cards WHERE c_login = \$1;`

	tests := []struct {
		name    string
		mockAct func()
		wantErr bool
		errText string
	}{
		{
			name: "Успешное удаление нескольких карт пользователя",
			mockAct: func() {
				// Симулируем удаление 3 сохраненных карт
				mock.ExpectExec(deleteSQL).
					WithArgs(testLogin).
					WillReturnResult(sqlmock.NewResult(0, 3)) // RowsAffected = 3
			},
			wantErr: false,
		},
		{
			name: "Успешное выполнение, если у пользователя не было привязанных карт",
			mockAct: func() {
				// Удалено 0 строк — это штатный сценарий, ошибка не возвращается
				mock.ExpectExec(deleteSQL).
					WithArgs(testLogin).
					WillReturnResult(sqlmock.NewResult(0, 0)) // RowsAffected = 0
			},
			wantErr: false,
		},
		{
			name: "Ошибка базы данных (например, разрыв соединения)",
			mockAct: func() {
				// Симулируем внутреннюю ошибку СУБД
				mock.ExpectExec(deleteSQL).
					WithArgs(testLogin).
					WillReturnError(errors.New("postgres: connection refused"))
			},
			wantErr: true,
			errText: "postgres: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем поведение мока под текущий кейс
			tt.mockAct()

			// Вызываем метод массового удаления
			err := sc.DeleteAllCardsByLogin(context.Background(), testLogin)

			// Проверяем наличие ошибки
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteAllCardsByLogin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Проверяем текст ошибки
			if tt.wantErr && err != nil && err.Error() != tt.errText {
				t.Errorf("DeleteAllCardsByLogin() error text = %q, want %q", err.Error(), tt.errText)
			}

			// Убеждаемся, что все ожидания мока выполнились
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
