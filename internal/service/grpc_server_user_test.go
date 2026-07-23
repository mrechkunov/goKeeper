package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mrechkunov/goKeeper.git/internal/auth"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/service/dbservice"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestGoKeeperServer_GetPassHash(t *testing.T) {
	server := &GoKeeperServer{}
	ctx := context.Background()

	// Сохраняем оригинальную функцию dbservice, чтобы восстановить её после теста
	oldGetUser := dbservice.GetUser
	defer func() { dbservice.GetUser = oldGetUser }()

	// Тестовые данные
	testLogin := "mrechkunov_user"
	testHash := "$2a$14$fakebcryptandhashstringforgokeepertest"
	unknown_user := "unknown_user"
	tests := []struct {
		name        string
		req         *pb.User
		mockAct     func()
		wantLogin   string
		wantHash    string
		wantErr     bool
		expectedErr string
	}{
		{
			name: "Успешно: Хэш пользователя найден и возвращен",
			req:  pb.User_builder{Login: &testLogin}.Build(),
			mockAct: func() {
				// Настраиваем мок на возврат успешных данных
				dbservice.GetUser = func(ctx context.Context, login string) (model.Users, error) {
					if login == testLogin {
						return model.Users{
							Login:        testLogin,
							PasswordHash: testHash,
						}, nil
					}
					return model.Users{}, errors.New("not found")
				}
			},
			wantLogin: testLogin,
			wantHash:  testHash,
			wantErr:   false,
		},
		{
			name: "Ошибка: Пользователь не существует в базе данных",
			req:  pb.User_builder{Login: &unknown_user}.Build(),
			mockAct: func() {
				// Настраиваем мок на возврат ошибки базы данных
				dbservice.GetUser = func(ctx context.Context, login string) (model.Users, error) {
					return model.Users{}, errors.New("user not found in storage")
				}
			},
			wantErr:     true,
			expectedErr: "user not found in storage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Инициализируем поведение dbservice для текущего кейса
			tt.mockAct()

			// Вызываем тестируемый метод gRPC-сервера
			resp, err := server.GetPassHash(ctx, tt.req)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetPassHash() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем её текст
			if tt.wantErr && err != nil {
				if err.Error() != tt.expectedErr {
					t.Errorf("GetPassHash() error text = %q, want %q", err.Error(), tt.expectedErr)
				}
				return
			}

			// 3. Если всё успешно, проверяем корректность сборки Protobuf-ответа через builder
			if resp.GetLogin() != tt.wantLogin {
				t.Errorf("GetPassHash() resp.Login = %q, want %q", resp.GetLogin(), tt.wantLogin)
			}

			if resp.GetPasswordHash() != tt.wantHash {
				t.Errorf("GetPassHash() resp.PasswordHash = %q, want %q", resp.GetPasswordHash(), tt.wantHash)
			}
		})
	}
}

// Реализуем интерфейс grpc.ServerTransportStream
type mockServerTransportStream struct {
	capturedMD metadata.MD
}

func (m *mockServerTransportStream) Method() string {
	return "/mrechkunov.goKeeper.proto.GoKeeper/RegisterUser"
}

func (m *mockServerTransportStream) SetHeader(md metadata.MD) error {
	for k, v := range md {
		m.capturedMD[k] = append(m.capturedMD[k], v...)
	}
	return nil
}

func (m *mockServerTransportStream) SendHeader(md metadata.MD) error {
	return nil
}

func (m *mockServerTransportStream) SetTrailer(md metadata.MD) error {
	return nil
}

func TestGoKeeperServer_RegisterUser(t *testing.T) {
	server := &GoKeeperServer{}

	// Сохраняем оригинал функции AddUser для восстановления после тестов
	oldAddUser := dbservice.AddUser
	defer func() { dbservice.AddUser = oldAddUser }()

	testLogin := "mrechkunov_grpc"
	testPassword := "my_password_123"
	duplicate_user := "duplicate_user"
	tests := []struct {
		name         string
		req          *pb.User
		mockAct      func()
		wantErr      bool
		expectedErr  string
		checkHeaders bool
	}{
		{
			name: "Успешно: Пользователь зарегистрирован и токен выдан",
			req: pb.User_builder{
				Login:        &testLogin,
				PasswordHash: &testPassword,
			}.Build(),
			mockAct: func() {
				dbservice.AddUser = func(ctx context.Context, user model.Users) error {
					return nil // Успешная запись в БД
				}
			},
			wantErr:      false,
			checkHeaders: true,
		},
		{
			name: "Ошибка: Дубликат логина в базе данных",
			req: pb.User_builder{
				Login:        &duplicate_user,
				PasswordHash: &testPassword,
			}.Build(),
			mockAct: func() {
				dbservice.AddUser = func(ctx context.Context, user model.Users) error {
					return errors.New("user already exists")
				}
			},
			wantErr:     true,
			expectedErr: "user already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockAct()

			//Создаем  транспортный поток
			stream := &mockServerTransportStream{capturedMD: metadata.Pairs()}
			ctx := grpc.NewContextWithServerTransportStream(context.Background(), stream)

			_, err := server.RegisterUser(ctx, tt.req)

			// Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Если ошибка ожидалась, проверяем её текст
			if tt.wantErr && err != nil {
				if err.Error() != tt.expectedErr {
					t.Errorf("RegisterUser() error text = %q, want %q", err.Error(), tt.expectedErr)
				}
				return
			}

			// Для успешного кейса проверяем токен в заголовках
			if tt.checkHeaders {
				tokens := stream.capturedMD["authorization"]
				if len(tokens) == 0 {
					t.Fatal("RegisterUser() не прикрепил токен авторизации в grpc метаданные")
				}

				if tokens[0] == "" {
					t.Fatal("Переданный токен авторизации оказался пустым")
				}
			}
		})
	}
}

func TestGoKeeperServer_AuthenticateUser(t *testing.T) {
	server := &GoKeeperServer{}

	// Сохраняем оригинал функции GetUser для восстановления после тестов
	oldGetUser := dbservice.GetUser
	defer func() { dbservice.GetUser = oldGetUser }()

	testLogin := "mrechkunov_auth"
	testPassword := "my_secret_pass"
	wrong_password := "wrong_password"
	unknown_user := "unknown_user"

	// Генерируем валидный bcrypt-хэш для СУБД
	hashedPassword, err := auth.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("Не удалось подготовить хэш пароля для теста: %v", err)
	}

	tests := []struct {
		name         string
		req          *pb.User
		mockAct      func()
		wantErr      bool
		expectedCode codes.Code
		checkHeaders bool
	}{
		{
			name: "Успешно: Корректная пара логин/пароль, токен выдан",
			req: pb.User_builder{
				Login:        &testLogin,
				PasswordHash: &testPassword,
			}.Build(),
			mockAct: func() {
				dbservice.GetUser = func(ctx context.Context, login string) (model.Users, error) {
					if login == testLogin {
						return model.Users{Login: testLogin, PasswordHash: hashedPassword}, nil
					}
					return model.Users{}, errors.New("not found")
				}
			},
			wantErr:      false,
			checkHeaders: true,
		},
		{
			name: "Ошибка: Неверный пароль пользователя (Unauthenticated)",
			req: pb.User_builder{
				Login:        &testLogin,
				PasswordHash: &wrong_password,
			}.Build(),
			mockAct: func() {
				dbservice.GetUser = func(ctx context.Context, login string) (model.Users, error) {
					return model.Users{Login: testLogin, PasswordHash: hashedPassword}, nil
				}
			},
			wantErr:      true,
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Ошибка: Пользователь не найден в базе данных",
			req: pb.User_builder{
				Login:        &unknown_user,
				PasswordHash: &testPassword,
			}.Build(),
			mockAct: func() {
				dbservice.GetUser = func(ctx context.Context, login string) (model.Users, error) {
					return model.Users{}, errors.New("user not found")
				}
			},
			wantErr:      true,
			expectedCode: codes.Unknown, // Ошибка dbservice возвращается "как есть"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockAct()

			// Перехватываем транспортный поток заголовков
			stream := &mockServerTransportStream{capturedMD: metadata.Pairs()}
			ctx := grpc.NewContextWithServerTransportStream(context.Background(), stream)

			_, err := server.AuthenticateUser(ctx, tt.req)

			// 1. Проверяем наличие ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("AuthenticateUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем её gRPC status-код
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if !ok && tt.expectedCode != codes.Unknown {
					t.Fatalf("Ожидался статус-ошибка gRPC, получили обычную ошибку: %v", err)
				}
				if ok && st.Code() != tt.expectedCode {
					t.Errorf("AuthenticateUser() код ответа = %v, ожидали = %v", st.Code(), tt.expectedCode)
				}
				return
			}

			// 3. Для успешного входа проверяем наличие выданного токена
			if tt.checkHeaders {
				tokens := stream.capturedMD["authorization"]
				if len(tokens) == 0 {
					t.Fatal("AuthenticateUser() не записал токен в gRPC метаданные")
				}
				if tokens[0] == "" {
					t.Fatal("Выданный токен авторизации оказался пустой строкой")
				}
			}
		})
	}
}

func TestGoKeeperServer_EditUser(t *testing.T) {
	server := &GoKeeperServer{}

	// Сохраняем оригинал функции EditUser для восстановления после тестов
	oldEditUser := dbservice.EditUser
	defer func() { dbservice.EditUser = oldEditUser }()

	testLogin := "mrechkunov_edit"
	testNewPassword := "brand_new_password_2026"

	tests := []struct {
		name         string
		ctx          context.Context
		req          *pb.User
		mockAct      func()
		wantErr      bool
		expectedCode codes.Code
	}{
		{
			name: "Успешно: Пароль изменен авторизованным пользователем",
			ctx:  context.WithValue(context.Background(), userLoginKey, testLogin),
			req:  pb.User_builder{PasswordHash: &testNewPassword}.Build(),
			mockAct: func() {
				dbservice.EditUser = func(ctx context.Context, user model.Users) error {
					if user.Login != testLogin {
						return errors.New("wrong login in storage step")
					}
					return nil // Симулируем успешный апдейт в БД
				}
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Пользователь не авторизован (нет логина в контексте)",
			ctx:  context.Background(), // Пустой контекст без userLoginKey
			req:  pb.User_builder{PasswordHash: &testNewPassword}.Build(),
			mockAct: func() {
				dbservice.EditUser = func(ctx context.Context, user model.Users) error {
					return nil
				}
			},
			wantErr:      true,
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Ошибка: Сбой СУБД при обновлении записи",
			ctx:  context.WithValue(context.Background(), userLoginKey, testLogin),
			req:  pb.User_builder{PasswordHash: &testNewPassword}.Build(),
			mockAct: func() {
				dbservice.EditUser = func(ctx context.Context, user model.Users) error {
					return errors.New("db connection failure")
				}
			},
			wantErr:      true,
			expectedCode: codes.Unknown, // Ошибка СУБД прокидывается как есть
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockAct()

			resp, err := server.EditUser(tt.ctx, tt.req)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("EditUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ожидалась gRPC-ошибка, проверяем статус-код
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if ok && st.Code() != tt.expectedCode {
					t.Errorf("EditUser() код ответа = %v, ожидали = %v", st.Code(), tt.expectedCode)
				}
				return
			}

			// 3. Если всё успешно, проверяем корректность возвращаемого Protobuf-объекта
			if resp.GetLogin() != testLogin {
				t.Errorf("EditUser() resp.Login = %q, want %q", resp.GetLogin(), testLogin)
			}

			if !strings.HasPrefix(resp.GetPasswordHash(), "$2") {
				t.Errorf("Пароль в ответе не захеширован: %q", resp.GetPasswordHash())
			}
		})
	}
}

func TestGoKeeperServer_DeleteUser(t *testing.T) {
	server := &GoKeeperServer{}
	ctx := context.Background()

	// Сейвим оригинальный метод базы данных для восстановления после тестов
	oldDeleteUser := dbservice.DeleteUser
	defer func() { dbservice.DeleteUser = oldDeleteUser }()

	testLogin := "mrechkunov_delete"
	testHash := "$2a$14$fakehashstringfordeleteusergokeeper"

	tests := []struct {
		name         string
		req          *pb.User
		mockAct      func()
		wantErr      bool
		expectedCode codes.Code
	}{
		{
			name: "Успешно: Пользователь полностью удален из базы",
			req: pb.User_builder{
				Login:        &testLogin,
				PasswordHash: &testHash,
			}.Build(),
			mockAct: func() {
				dbservice.DeleteUser = func(ctx context.Context, user model.Users) error {
					if user.Login != testLogin || user.PasswordHash != testHash {
						return errors.New("passed data mismatch")
					}
					return nil // Симулируем успех
				}
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Сбой СУБД на удалении (Internal)",
			req: pb.User_builder{
				Login:        &testLogin,
				PasswordHash: &testHash,
			}.Build(),
			mockAct: func() {
				dbservice.DeleteUser = func(ctx context.Context, user model.Users) error {
					return errors.New("hard drive io error") // Имитируем падение диска
				}
			},
			wantErr:      true,
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockAct()

			resp, err := server.DeleteUser(ctx, tt.req)

			// 1. Чекаем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если упало по плану, валидируем gRPC статус код и месседж
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус-ошибка gRPC, получили обычную ошибку: %v", err)
				}
				if st.Code() != tt.expectedCode {
					t.Errorf("DeleteUser() код ответа = %v, ожидали = %v", st.Code(), tt.expectedCode)
				}

				expectedMsg := "server error user not deleted"
				if st.Message() != expectedMsg {
					t.Errorf("DeleteUser() сообщение ошибки = %q, ожидали = %q", st.Message(), expectedMsg)
				}
			}

			// 3. Если всё ок, убеждаемся, что вернулся проинициализированный EmptyMessage
			if !tt.wantErr {
				if resp == nil {
					t.Fatal("DeleteUser() вернул nil вместо &pb.EmptyMessage{}")
				}
			}
		})
	}
}
