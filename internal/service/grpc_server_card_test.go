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

func TestGoKeeperServer_SaveCard(t *testing.T) {
	server := &service.GoKeeperServer{}
	ctx := context.Background()

	// Сохраняем оригинал функции AddCard для восстановления после тестов
	oldAddCard := dbservice.AddCard
	defer func() { dbservice.AddCard = oldAddCard }()

	testLogin := "mrechkunov_holder"
	testCipher := "encrypted_base64_card_payload_gcm"
	testMetadata := "travel_mastercard"

	tests := []struct {
		name         string
		req          *pb.CardData
		mockAct      func()
		wantErr      bool
		expectedCode codes.Code
		expectedMsg  string
	}{
		{
			name: "Успешно: Данные карты сохранены в репозиторий",
			req: pb.CardData_builder{
				Login:      &testLogin,
				Cipherdata: &testCipher,
				Metadata:   &testMetadata,
			}.Build(),
			mockAct: func() {
				dbservice.AddCard = func(ctx context.Context, data model.Cards) error {
					if data.UserLogin != testLogin || data.CipherData != testCipher || data.MetaData != testMetadata {
						return errors.New("mock db: properties mismatch")
					}
					return nil // Симулируем успех
				}
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Карта с такими метаданными уже существует (AlreadyExists)",
			req: pb.CardData_builder{
				Login:      &testLogin,
				Cipherdata: &testCipher,
				Metadata:   &testMetadata,
			}.Build(),
			mockAct: func() {
				dbservice.AddCard = func(ctx context.Context, data model.Cards) error {
					return errors.New("duplicate key unique constraint violation")
				}
			},
			wantErr:      true,
			expectedCode: codes.AlreadyExists,
			expectedMsg:  "server error card not saved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockAct()

			// Вызываем тестируемый gRPC-метод
			resp, err := server.SaveCard(ctx, tt.req)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("SaveCard() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ожидался сбой, проверяем gRPC-код и текст ошибки
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус gRPC, получили обычную ошибку: %v", err)
				}
				if st.Code() != tt.expectedCode {
					t.Errorf("SaveCard() код ответа = %v, ожидали = %v", st.Code(), tt.expectedCode)
				}
				if st.Message() != tt.expectedMsg {
					t.Errorf("SaveCard() сообщение ошибки = %q, ожидали = %q", st.Message(), tt.expectedMsg)
				}
				return
			}

			// 3. При успешном сценарии проверяем, что возвращается проинициализированный EmptyMessage
			if !tt.wantErr {
				if resp == nil {
					t.Fatal("SaveCard() вернул nil вместо &pb.EmptyMessage{}")
				}
			}
		})
	}
}

func TestGoKeeperServer_GetCard(t *testing.T) {
	server := &service.GoKeeperServer{}
	ctx := context.Background()

	// Сохраняем оригинал функции GetCard для восстановления после тестов
	oldGetCard := dbservice.GetCard
	defer func() { dbservice.GetCard = oldGetCard }()

	testLogin := "mrechkunov_client"
	testMetadata := "personal_visa"
	testCipher := "encrypted_card_fields_base64_string"
	missingMeta := "missing_meta"
	tests := []struct {
		name         string
		req          *pb.CardData
		mockAct      func()
		wantLogin    string
		wantCipher   string
		wantMetadata string
		wantErr      bool
		expectedErr  string
	}{
		{
			name: "Успешно: Данные карты найдены в базе данных",
			req: pb.CardData_builder{
				Login:    &testLogin,
				Metadata: &testMetadata,
			}.Build(),
			mockAct: func() {
				dbservice.GetCard = func(ctx context.Context, login, metadata string) (model.Cards, error) {
					if login == testLogin && metadata == testMetadata {
						return model.Cards{
							UserLogin:  testLogin,
							CipherData: testCipher,
							MetaData:   testMetadata,
						}, nil
					}
					return model.Cards{}, errors.New("not found")
				}
			},
			wantLogin:    testLogin,
			wantCipher:   testCipher,
			wantMetadata: testMetadata,
			wantErr:      false,
		},
		{
			name: "Ошибка: Запись отсутствует в базе данных",
			req: pb.CardData_builder{
				Login:    &testLogin,
				Metadata: &missingMeta,
			}.Build(),
			mockAct: func() {
				dbservice.GetCard = func(ctx context.Context, login, metadata string) (model.Cards, error) {
					return model.Cards{}, errors.New("card record not found")
				}
			},
			wantErr:     true,
			expectedErr: "card record not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockAct()

			// Вызываем метод получения карты
			resp, err := server.GetCard(ctx, tt.req)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCard() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем её текст
			if tt.wantErr && err != nil {
				if err.Error() != tt.expectedErr {
					t.Errorf("GetCard() error text = %q, want %q", err.Error(), tt.expectedErr)
				}
				return
			}

			// 3. Если всё успешно, проверяем корректность сборки Protobuf-ответа через builder
			if resp.GetLogin() != tt.wantLogin {
				t.Errorf("GetCard() resp.Login = %q, want %q", resp.GetLogin(), tt.wantLogin)
			}
			if resp.GetCipherdata() != tt.wantCipher {
				t.Errorf("GetCard() resp.Cipherdata = %q, want %q", resp.GetCipherdata(), tt.wantCipher)
			}
			if resp.GetMetadata() != tt.wantMetadata {
				t.Errorf("GetCard() resp.Metadata = %q, want %q", resp.GetMetadata(), tt.wantMetadata)
			}
		})
	}
}

func TestGoKeeperServer_EditCard(t *testing.T) {
	server := &service.GoKeeperServer{}
	ctx := context.Background()

	// Сохраняем оригинал функции EditCard для восстановления после тестов
	oldEditCard := dbservice.EditCard
	defer func() { dbservice.EditCard = oldEditCard }()

	testLogin := "mrechkunov_user"
	testCipher := "new_encrypted_base64_card_payload"
	testMetadata := "salary_visa"
	nonExistentCardMeta := "non_existent_card_meta"
	tests := []struct {
		name        string
		req         *pb.CardData
		mockAct     func()
		wantErr     bool
		expectedErr string
	}{
		{
			name: "Успешно: Данные карты обновлены в базе данных",
			req: pb.CardData_builder{
				Login:      &testLogin,
				Cipherdata: &testCipher,
				Metadata:   &testMetadata,
			}.Build(),
			mockAct: func() {
				dbservice.EditCard = func(ctx context.Context, data model.Cards) error {
					if data.UserLogin != testLogin || data.CipherData != testCipher || data.MetaData != testMetadata {
						return errors.New("mock db: properties mismatch")
					}
					return nil // Симулируем успешный апдейт
				}
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Изменяемая карта не найдена в базе данных",
			req: pb.CardData_builder{
				Login:      &testLogin,
				Cipherdata: &testCipher,
				Metadata:   &nonExistentCardMeta,
			}.Build(),
			mockAct: func() {
				dbservice.EditCard = func(ctx context.Context, data model.Cards) error {
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

			// Вызываем метод изменения данных карты
			resp, err := server.EditCard(ctx, tt.req)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("EditCard() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем её текст
			if tt.wantErr && err != nil {
				if err.Error() != tt.expectedErr {
					t.Errorf("EditCard() error text = %q, want %q", err.Error(), tt.expectedErr)
				}
				return
			}

			// 3. При успешном сценарии проверяем, что возвращается проинициализированный EmptyMessage
			if !tt.wantErr {
				if resp == nil {
					t.Fatal("EditCard() вернул nil вместо &pb.EmptyMessage{}")
				}
			}
		})
	}
}

func TestGoKeeperServer_DeleteCard(t *testing.T) {
	server := &service.GoKeeperServer{}
	ctx := context.Background()

	// Сохраняем оригинал функции DeleteCard для восстановления после тестов
	oldDeleteCard := dbservice.DeleteCard
	defer func() { dbservice.DeleteCard = oldDeleteCard }()

	testLogin := "mrechkunov_user"
	testCipher := "encrypted_base64_card_payload_to_delete"
	testMetadata := "old_expired_card"
	nonExistentCard := "non_existent_card"
	tests := []struct {
		name        string
		req         *pb.CardData
		mockAct     func()
		wantErr     bool
		expectedErr string
	}{
		{
			name: "Успешно: Данные карты удалены из базы данных",
			req: pb.CardData_builder{
				Login:      &testLogin,
				Cipherdata: &testCipher,
				Metadata:   &testMetadata,
			}.Build(),
			mockAct: func() {
				dbservice.DeleteCard = func(ctx context.Context, data model.Cards) error {
					if data.UserLogin != testLogin || data.CipherData != testCipher || data.MetaData != testMetadata {
						return errors.New("mock db: properties mismatch")
					}
					return nil // Симулируем успешное удаление
				}
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Удаляемая карта не найдена в базе данных",
			req: pb.CardData_builder{
				Login:      &testLogin,
				Cipherdata: &testCipher,
				Metadata:   &nonExistentCard,
			}.Build(),
			mockAct: func() {
				dbservice.DeleteCard = func(ctx context.Context, data model.Cards) error {
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

			// Вызываем метод удаления карты
			resp, err := server.DeleteCard(ctx, tt.req)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("DeleteCard() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем её текст
			if tt.wantErr && err != nil {
				if err.Error() != tt.expectedErr {
					t.Errorf("DeleteCard() error text = %q, want %q", err.Error(), tt.expectedErr)
				}
				return
			}

			// 3. При успешном сценарии проверяем, что возвращается проинициализированный EmptyMessage
			if !tt.wantErr {
				if resp == nil {
					t.Fatal("DeleteCard() вернул nil вместо &pb.EmptyMessage{}")
				}
			}
		})
	}
}
