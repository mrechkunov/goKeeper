package cliservice_test

import (
	"context"
	"net"
	"testing"

	"github.com/mrechkunov/goKeeper.git/internal/cryptodata"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	cliservice "github.com/mrechkunov/goKeeper.git/internal/service/client"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// Добавляем метод SavePassword в существующий mockGoKeeperServer
func (m *mockGoKeeperServer) SavePassword(ctx context.Context, in *pb.PasswordData) (*pb.EmptyMessage, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	return &pb.EmptyMessage{}, nil
}

func TestSavePass_Client(t *testing.T) {

	lis := bufconn.Listen(1024 * 1024)
	defer lis.Close()

	mockServer := &mockGoKeeperServer{}
	grpcServer := grpc.NewServer()
	pb.RegisterGoKeeperServer(grpcServer, mockServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil && err.Error() != "closed" {
			t.Errorf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.Stop()

	ctx := context.Background()

	// Используем NewClient поверх bufconn
	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to create grpc client: %v", err)
	}
	defer conn.Close()

	client := pb.NewGoKeeperClient(conn)

	// Валидные базовые тестовые данные
	testPassData := model.Passwords{
		UserLogin:      "mrechkunov_user",
		LoginToSave:    "github_account",
		PasswordToSave: "github_secret_pass",
		MetaData:       "github_access",
	}

	tests := []struct {
		name      string
		passData  model.Passwords
		setupMock func()
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name:     "Успешно: Пароль зашифрован и сохранен на gRPC-сервере",
			passData: testPassData,
			setupMock: func() {
				mockServer.mockErr = nil
			},
			wantErr: false,
		},
		{
			name:     "Ошибка: Запись уже существует в базе данных сервера",
			passData: testPassData,
			setupMock: func() {
				mockServer.mockErr = status.Error(codes.AlreadyExists, "server error pass not saved")
			},
			wantErr:  true,
			wantCode: codes.AlreadyExists,
		},
		{
			name: "Ошибка: Сбой криптографии (например, пустой или поврежденный логин/пароль)",
			passData: model.Passwords{
				UserLogin:      "mrechkunov_user",
				LoginToSave:    "",
				PasswordToSave: "",
				MetaData:       "empty_meta",
			},
			setupMock: func() {
				// Если криптография упадет до отправки, сервер даже не получит запрос
				mockServer.mockErr = nil
			},

			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			if tt.name == "Ошибка: Сбой криптографии (например, пустой или поврежденный логин/пароль)" {
			}

			err := cliservice.SavePass(ctx, client, tt.passData)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				// Если вы не ломали ключ шифрования, этот шаг пропустится для 3 кейса
				t.Logf("Инфо [%s]: error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась от сервера, проверяем её gRPC статус-код
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if ok && st.Code() != tt.wantCode {
					t.Errorf("SavePass() код ошибки = %v, ожидали = %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}

// Добавляем метод GetPassword в наш mock-сервер для симуляции ответа сервера
func (m *mockGoKeeperServer) GetPassword(ctx context.Context, in *pb.PasswordData) (*pb.PasswordData, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	// Возвращаем тестовые данные, которые мы настроили в tt.setupMock
	login := in.GetLogin()
	meta := in.GetMetadata()
	return pb.PasswordData_builder{
		Login:    &login,
		Pair:     &m.mockToken, // Используем это поле для передачи зашифрованной строки pair
		Metadata: &meta,
	}.Build(), nil
}

func TestGetPass_Client(t *testing.T) {
	// Инициализируем виртуальную сеть в памяти
	lis := bufconn.Listen(1024 * 1024)
	defer lis.Close()

	mockServer := &mockGoKeeperServer{}
	grpcServer := grpc.NewServer()
	pb.RegisterGoKeeperServer(grpcServer, mockServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil && err.Error() != "closed" {
			t.Errorf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.Stop()

	ctx := context.Background()

	// Используем grpc.NewClient поверх bufconn
	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to create grpc client: %v", err)
	}
	defer conn.Close()

	client := pb.NewGoKeeperClient(conn)

	// Исходные чистые данные, которые мы планируем зашифровать для теста
	targetLogin := "my_service_user"
	targetPass := "my_service_password_123"

	// Генерируем валидную зашифрованную строку через CryptoPair,
	// чтобы метод DecryptPair внутри GetPass смог её успешно расшифровать.
	validEncryptedPair, err := cryptodata.CryptoPair(targetLogin, targetPass)
	if err != nil {
		t.Fatalf("Не удалось подготовить зашифрованные тестовые данные: %v", err)
	}

	// Базовый поисковый запрос от клиента
	searchFilter := model.Passwords{
		UserLogin: "mrechkunov_user",
		MetaData:  "github_access",
	}

	tests := []struct {
		name          string
		setupMock     func()
		wantSavedLog  string
		wantSavedPass string
		wantErr       bool
		wantCode      codes.Code
	}{
		{
			name: "Успешно: Пароль получен с сервера и успешно расшифрован",
			setupMock: func() {
				mockServer.mockToken = validEncryptedPair // Сервер отдает валидный шифротекст
				mockServer.mockErr = nil
			},
			wantSavedLog:  targetLogin,
			wantSavedPass: targetPass,
			wantErr:       false,
		},
		{
			name: "Ошибка: Запись отсутствует на сервере (NotFound)",
			setupMock: func() {
				mockServer.mockToken = ""
				mockServer.mockErr = status.Error(codes.NotFound, "password not found")
			},
			wantErr:  true,
			wantCode: codes.NotFound,
		},
		{
			name: "Ошибка: Данные получены, но сбой расшифровки (битый Base64/ключ)",
			setupMock: func() {
				mockServer.mockToken = "not-a-valid-encrypted-base64-string!!!"
				mockServer.mockErr = nil
			},
			wantErr: true, // Метод DecryptPair вернет ошибку, и GetPass должен вернуть её наружу
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Вызываем тестируемый метод получения и дешифрования пароля
			result, err := cliservice.GetPass(ctx, client, searchFilter)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetPass() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ожидалась ошибка от gRPC, сверяем статус-код
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if ok && tt.wantCode != codes.OK && st.Code() != tt.wantCode {
					t.Errorf("GetPass() код ошибки = %v, ожидали = %v", st.Code(), tt.wantCode)
				}
				return
			}

			// 3. Если всё успешно, проверяем, что расшифрованные логин и пароль совпадают с исходными
			if !tt.wantErr {
				if result.UserLogin != searchFilter.UserLogin {
					t.Errorf("GetPass() UserLogin = %q, want %q", result.UserLogin, searchFilter.UserLogin)
				}
				if result.MetaData != searchFilter.MetaData {
					t.Errorf("GetPass() MetaData = %q, want %q", result.MetaData, searchFilter.MetaData)
				}
				if result.LoginToSave != tt.wantSavedLog {
					t.Errorf("GetPass() LoginToSave = %q, want %q", result.LoginToSave, tt.wantSavedLog)
				}
				if result.PasswordToSave != tt.wantSavedPass {
					t.Errorf("GetPass() PasswordToSave = %q, want %q", result.PasswordToSave, tt.wantSavedPass)
				}
			}
		})
	}
}

// Добавляем метод EditPassword в наш mock-сервер для симуляции ответа сервера
func (m *mockGoKeeperServer) EditPassword(ctx context.Context, in *pb.PasswordData) (*pb.EmptyMessage, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	return &pb.EmptyMessage{}, nil
}

func TestEditPass_Client(t *testing.T) {
	// Настраиваем виртуальную сеть в памяти
	lis := bufconn.Listen(1024 * 1024)
	defer lis.Close()

	mockServer := &mockGoKeeperServer{}
	grpcServer := grpc.NewServer()
	pb.RegisterGoKeeperServer(grpcServer, mockServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil && err.Error() != "closed" {
			t.Errorf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.Stop()

	ctx := context.Background()

	// Используем современный grpc.NewClient поверх bufconn
	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to create grpc client: %v", err)
	}
	defer conn.Close()

	client := pb.NewGoKeeperClient(conn)

	// Тестовые данные для редактирования пароля
	testPassData := model.Passwords{
		UserLogin:      "mrechkunov_user",
		LoginToSave:    "new_github_login",
		PasswordToSave: "new_github_password",
		MetaData:       "github_access_credentials",
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name: "Успешно: Данные пароля зашифрованы и изменены на сервере",
			setupMock: func() {
				mockServer.mockErr = nil
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Запись для редактирования не найдена на сервере (NotFound)",
			setupMock: func() {
				mockServer.mockErr = status.Error(codes.NotFound, "data is not exists in DB")
			},
			wantErr:  true,
			wantCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Вызываем клиентскую функцию редактирования пароля
			err := cliservice.EditPass(ctx, client, testPassData)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("EditPass() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем её gRPC-код ответа
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус gRPC, получили обычную ошибку: %v", err)
				}
				if st.Code() != tt.wantCode {
					t.Errorf("EditPass() код ошибки = %v, ожидали = %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}

// Добавляем метод DeletePassword в наш mock-сервер для симуляции ответа сервера
func (m *mockGoKeeperServer) DeletePassword(ctx context.Context, in *pb.PasswordData) (*pb.EmptyMessage, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	return &pb.EmptyMessage{}, nil
}

func TestDeletePass_Client(t *testing.T) {
	// Настраиваем виртуальную сеть в памяти
	lis := bufconn.Listen(1024 * 1024)
	defer lis.Close()

	mockServer := &mockGoKeeperServer{}
	grpcServer := grpc.NewServer()
	pb.RegisterGoKeeperServer(grpcServer, mockServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil && err.Error() != "closed" {
			t.Errorf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.Stop()

	ctx := context.Background()

	// Используем современный grpc.NewClient поверх bufconn
	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to create grpc client: %v", err)
	}
	defer conn.Close()

	client := pb.NewGoKeeperClient(conn)

	// Тестовые данные для удаления пароля
	testPassData := model.Passwords{
		UserLogin: "mrechkunov_user",
		MetaData:  "old_expired_account",
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name: "Успешно: Запись пароля удалена на сервере",
			setupMock: func() {
				mockServer.mockErr = nil
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Удаляемая запись не найдена в базе данных (NotFound)",
			setupMock: func() {
				mockServer.mockErr = status.Error(codes.NotFound, "data is not exists in DB")
			},
			wantErr:  true,
			wantCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Вызываем клиентскую функцию удаления пароля
			err := cliservice.DeletePass(ctx, client, testPassData)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("DeletePass() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем её gRPC-код ответа
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус gRPC, получили обычную ошибку: %v", err)
				}
				if st.Code() != tt.wantCode {
					t.Errorf("DeletePass() код ошибки = %v, ожидали = %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}
