package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/service"
	"github.com/mrechkunov/goKeeper.git/internal/service/dbservice"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGoKeeperServer_SavePassword(t *testing.T) {
	server := &service.GoKeeperServer{}
	ctx := context.Background()

	// Сохраняем оригинал функции AddPassword для восстановления после тестов
	oldAddPassword := dbservice.AddPassword
	defer func() { dbservice.AddPassword = oldAddPassword }()

	testLogin := "mrechkunov_user"
	testPair := "encrypted_base64_pair"
	testMetadata := "github_account"

	tests := []struct {
		name         string
		req          *pb.PasswordData
		mockAct      func()
		wantErr      bool
		expectedCode codes.Code
		expectedMsg  string
	}{
		{
			name: "Успешно: Пароль сохранен в базе данных",
			req: pb.PasswordData_builder{
				Login:    &testLogin,
				Pair:     &testPair,
				Metadata: &testMetadata,
			}.Build(),
			mockAct: func() {
				dbservice.AddPassword = func(ctx context.Context, data model.Passwords) error {
					if data.UserLogin != testLogin || data.Pair != testPair || data.MetaData != testMetadata {
						return errors.New("data mismatch in mock storage")
					}
					return nil // Симулируем успешную запись
				}
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Сбой СУБД (запись уже существует / конфликт)",
			req: pb.PasswordData_builder{
				Login:    &testLogin,
				Pair:     &testPair,
				Metadata: &testMetadata,
			}.Build(),
			mockAct: func() {
				dbservice.AddPassword = func(ctx context.Context, data model.Passwords) error {
					return errors.New("unique violation constraint") // Имитируем ошибку дубликата в БД
				}
			},
			wantErr:      true,
			expectedCode: codes.AlreadyExists,
			expectedMsg:  "server error pass not saved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockAct()

			// Вызываем метод сохранения пароля
			resp, err := server.SavePassword(ctx, tt.req)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("SavePassword() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем её gRPC статус-код и сообщение
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус-ошибка gRPC, получили обычную ошибку: %v", err)
				}
				if st.Code() != tt.expectedCode {
					t.Errorf("SavePassword() код ответа = %v, ожидали = %v", st.Code(), tt.expectedCode)
				}
				if st.Message() != tt.expectedMsg {
					t.Errorf("SavePassword() сообщение ошибки = %q, ожидали = %q", st.Message(), tt.expectedMsg)
				}
				return
			}

			// 3. При успешном сценарии проверяем, что возвращается проинициализированный EmptyMessage
			if !tt.wantErr {
				if resp == nil {
					t.Fatal("SavePassword() вернул nil вместо &pb.EmptyMessage{}")
				}
			}
		})
	}
}

func TestGoKeeperServer_GetPassword(t *testing.T) {
	server := &service.GoKeeperServer{}
	ctx := context.Background()

	// Сохраняем оригинал функции GetPassword для восстановления после тестов
	oldGetPassword := dbservice.GetPassword
	defer func() { dbservice.GetPassword = oldGetPassword }()

	testLogin := "mrechkunov_client"
	testMetadata := "yandex_mail"
	testPair := "encrypted_login_and_password_pair"
	unknown_meta := "unknown_meta"
	tests := []struct {
		name         string
		req          *pb.PasswordData
		mockAct      func()
		wantLogin    string
		wantPair     string
		wantMetadata string
		wantErr      bool
		expectedErr  string
	}{
		{
			name: "Успешно: Пароль найден в базе данных",
			req: pb.PasswordData_builder{
				Login:    &testLogin,
				Metadata: &testMetadata,
			}.Build(),
			mockAct: func() {
				dbservice.GetPassword = func(ctx context.Context, login, metadata string) (model.Passwords, error) {
					if login == testLogin && metadata == testMetadata {
						return model.Passwords{
							UserLogin: testLogin,
							Pair:      testPair,
							MetaData:  testMetadata,
						}, nil
					}
					return model.Passwords{}, errors.New("not found")
				}
			},
			wantLogin:    testLogin,
			wantPair:     testPair,
			wantMetadata: testMetadata,
			wantErr:      false,
		},
		{
			name: "Ошибка: Запись отсутствует в базе данных",
			req: pb.PasswordData_builder{
				Login:    &testLogin,
				Metadata: &unknown_meta,
			}.Build(),
			mockAct: func() {
				dbservice.GetPassword = func(ctx context.Context, login, metadata string) (model.Passwords, error) {
					return model.Passwords{}, errors.New("password record not found")
				}
			},
			wantErr:     true,
			expectedErr: "password record not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockAct()

			// Вызываем метод получения пароля
			resp, err := server.GetPassword(ctx, tt.req)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetPassword() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем её текст
			if tt.wantErr && err != nil {
				if err.Error() != tt.expectedErr {
					t.Errorf("GetPassword() error text = %q, want %q", err.Error(), tt.expectedErr)
				}
				return
			}

			// 3. Если всё успешно, проверяем корректность сборки Protobuf-ответа через builder
			if resp.GetLogin() != tt.wantLogin {
				t.Errorf("GetPassword() resp.Login = %q, want %q", resp.GetLogin(), tt.wantLogin)
			}
			if resp.GetPair() != tt.wantPair {
				t.Errorf("GetPassword() resp.Pair = %q, want %q", resp.GetPair(), tt.wantPair)
			}
			if resp.GetMetadata() != tt.wantMetadata {
				t.Errorf("GetPassword() resp.Metadata = %q, want %q", resp.GetMetadata(), tt.wantMetadata)
			}
		})
	}
}

func TestGoKeeperServer_EditPassword(t *testing.T) {
	server := &service.GoKeeperServer{}
	ctx := context.Background()

	// Сохраняем оригинал функции EditPassword для восстановления после тестов
	oldEditPassword := dbservice.EditPassword
	defer func() { dbservice.EditPassword = oldEditPassword }()

	testLogin := "mrechkunov_user"
	testPair := "new_encrypted_base64_pair"
	testMetadata := "yandex_account"
	non_existent_meta := "non_existent_meta"
	tests := []struct {
		name        string
		req         *pb.PasswordData
		mockAct     func()
		wantErr     bool
		expectedErr string
	}{
		{
			name: "Успешно: Данные пароля обновлены в базе данных",
			req: pb.PasswordData_builder{
				Login:    &testLogin,
				Pair:     &testPair,
				Metadata: &testMetadata,
			}.Build(),
			mockAct: func() {
				dbservice.EditPassword = func(ctx context.Context, data model.Passwords) error {
					if data.UserLogin != testLogin || data.Pair != testPair || data.MetaData != testMetadata {
						return errors.New("data mismatch in mock storage")
					}
					return nil // Симулируем успешный апдейт
				}
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Запись для обновления не найдена в базе данных",
			req: pb.PasswordData_builder{
				Login:    &testLogin,
				Pair:     &testPair,
				Metadata: &non_existent_meta,
			}.Build(),
			mockAct: func() {
				dbservice.EditPassword = func(ctx context.Context, data model.Passwords) error {
					return errors.New("data is not exists in DB")
				}
			},
			wantErr:     true,
			expectedErr: "data is not exists in DB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockAct()

			// Вызываем метод изменения пароля
			resp, err := server.EditPassword(ctx, tt.req)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("EditPassword() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем её текст
			if tt.wantErr && err != nil {
				if err.Error() != tt.expectedErr {
					t.Errorf("EditPassword() error text = %q, want %q", err.Error(), tt.expectedErr)
				}
				return
			}

			// 3. При успешном сценарии проверяем, что возвращается проинициализированный EmptyMessage
			if !tt.wantErr {
				if resp == nil {
					t.Fatal("EditPassword() вернул nil вместо &pb.EmptyMessage{}")
				}
			}
		})
	}
}

func TestGoKeeperServer_DeletePassword(t *testing.T) {
	server := &service.GoKeeperServer{}
	ctx := context.Background()

	// Сохраняем оригинал функции DeletePassword для восстановления после тестов
	oldDeletePassword := dbservice.DeletePassword
	defer func() { dbservice.DeletePassword = oldDeletePassword }()

	testLogin := "mrechkunov_user"
	testPair := "encrypted_base64_pair_to_delete"
	testMetadata := "old_bank_account"
	non_existent_meta := "non_existent_meta"
	tests := []struct {
		name        string
		req         *pb.PasswordData
		mockAct     func()
		wantErr     bool
		expectedErr string
	}{
		{
			name: "Успешно: Пароль удален из базы данных",
			req: pb.PasswordData_builder{
				Login:    &testLogin,
				Pair:     &testPair,
				Metadata: &testMetadata,
			}.Build(),
			mockAct: func() {
				dbservice.DeletePassword = func(ctx context.Context, data model.Passwords) error {
					if data.UserLogin != testLogin || data.Pair != testPair || data.MetaData != testMetadata {
						return errors.New("data mismatch in mock storage")
					}
					return nil // Симулируем успешное удаление
				}
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Запись для удаления не найдена в базе данных",
			req: pb.PasswordData_builder{
				Login:    &testLogin,
				Pair:     &testPair,
				Metadata: &non_existent_meta,
			}.Build(),
			mockAct: func() {
				dbservice.DeletePassword = func(ctx context.Context, data model.Passwords) error {
					return errors.New("data is not exists in DB")
				}
			},
			wantErr:     true,
			expectedErr: "data is not exists in DB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockAct()

			// Вызываем метод удаления пароля
			resp, err := server.DeletePassword(ctx, tt.req)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("DeletePassword() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем её текст
			if tt.wantErr && err != nil {
				if err.Error() != tt.expectedErr {
					t.Errorf("DeletePassword() error text = %q, want %q", err.Error(), tt.expectedErr)
				}
				return
			}

			// 3. При успешном сценарии проверяем, что возвращается проинициализированный EmptyMessage
			if !tt.wantErr {
				if resp == nil {
					t.Fatal("DeletePassword() вернул nil вместо &pb.EmptyMessage{}")
				}
			}
		})
	}
}
