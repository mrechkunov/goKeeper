package cliservice_test

import (
	"context"
	"net"
	"testing"

	"github.com/mrechkunov/goKeeper.git/internal/model"
	cliservice "github.com/mrechkunov/goKeeper.git/internal/service/client"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type mockGoKeeperServer struct {
	pb.UnimplementedGoKeeperServer
	mockToken string
	mockErr   error
}

func (m *mockGoKeeperServer) RegisterUser(ctx context.Context, in *pb.User) (*pb.EmptyMessage, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	headerMD := metadata.Pairs("authorization", m.mockToken)
	grpc.SetHeader(ctx, headerMD)
	return &pb.EmptyMessage{}, nil
}

func TestRegisterUser_Client(t *testing.T) {
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

	testUser := model.Users{
		Login:        "mrechkunov_client",
		PasswordHash: "super_secret_pass",
	}
	expectedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.testToken"

	tests := []struct {
		name      string
		setupMock func()
		wantToken string
		wantErr   bool
	}{
		{
			name: "Успешно: Пользователь зарегистрирован, токен извлечен",
			setupMock: func() {
				mockServer.mockToken = expectedToken
				mockServer.mockErr = nil
			},
			wantToken: expectedToken,
			wantErr:   false,
		},
		{
			name: "Ошибка: Сервер вернул AlreadyExists",
			setupMock: func() {
				mockServer.mockToken = ""
				mockServer.mockErr = status.Error(codes.AlreadyExists, "User already exist")
			},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			gotToken, err := cliservice.RegisterUser(ctx, client, testUser)

			if (err != nil) != tt.wantErr {
				t.Fatalf("RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && gotToken != tt.wantToken {
				t.Errorf("RegisterUser() gotToken = %q, want %q", gotToken, tt.wantToken)
			}
		})
	}
}

func (m *mockGoKeeperServer) AuthenticateUser(ctx context.Context, in *pb.User) (*pb.EmptyMessage, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	headerMD := metadata.Pairs("authorization", m.mockToken)
	grpc.SetHeader(ctx, headerMD)
	return &pb.EmptyMessage{}, nil
}

func TestAuthenticateUser_Client(t *testing.T) {
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

	// ИСПРАВЛЕНИЕ: Используем grpc.NewClient вместо grpc.DialContext
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

	testUser := model.Users{
		Login:        "active_user",
		PasswordHash: "valid_password_hash",
	}
	expectedToken := "eyJhbGciOiJIUzI1NiJ9.correct_auth_token_string"

	tests := []struct {
		name      string
		setupMock func()
		wantToken string
		wantErr   bool
	}{
		{
			name: "Успешно: Пользователь аутентифицирован, токен получен",
			setupMock: func() {
				mockServer.mockToken = expectedToken
				mockServer.mockErr = nil
			},
			wantToken: expectedToken,
			wantErr:   false,
		},
		{
			name: "Ошибка: Неверная пара логин/пароль (Unauthenticated)",
			setupMock: func() {
				mockServer.mockToken = ""
				mockServer.mockErr = status.Error(codes.Unauthenticated, "wrong pair login/password")
			},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			gotToken, err := cliservice.AuthenticateUser(ctx, client, testUser)

			if (err != nil) != tt.wantErr {
				t.Fatalf("AuthenticateUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && gotToken != tt.wantToken {
				t.Errorf("AuthenticateUser() gotToken = %q, want %q", gotToken, tt.wantToken)
			}
		})
	}
}

// Добавляем метод EditUser в наш mock-сервер для симуляции ответа
func (m *mockGoKeeperServer) EditUser(ctx context.Context, in *pb.User) (*pb.User, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	// Возвращаем измененного пользователя, как это заложено в контракте proto
	return in, nil
}

func TestChangePass_Client(t *testing.T) {
	// Инициализируем буфер сети в памяти
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

	// Используем современный grpc.NewClient для ленивого подключения поверх bufconn
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

	testUser := model.Users{
		Login:        "mrechkunov_user",
		PasswordHash: "brand_new_secure_password_2026",
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name: "Успешно: Пароль пользователя изменен на сервере",
			setupMock: func() {
				mockServer.mockErr = nil
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Токен просрочен или невалиден (Unauthenticated)",
			setupMock: func() {
				mockServer.mockErr = status.Error(codes.Unauthenticated, "token is not valid")
			},
			wantErr:  true,
			wantCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Вызываем тестируемую клиентскую функцию смены пароля
			err := cliservice.ChangePass(ctx, client, testUser)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("ChangePass() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ожидалась ошибка, проверяем gRPC-код ответа
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус gRPC, получили обычную ошибку: %v", err)
				}
				if st.Code() != tt.wantCode {
					t.Errorf("ChangePass() код ошибки = %v, ожидали = %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}

// Добавляем метод DeleteUser в наш mock-сервер для симуляции ответа сервера
func (m *mockGoKeeperServer) DeleteUser(ctx context.Context, in *pb.User) (*pb.EmptyMessage, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	return &pb.EmptyMessage{}, nil
}

func TestDeleteUser_Client(t *testing.T) {
	// Инициализируем буфер сети в оперативной памяти
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

	// Используем современный grpc.NewClient для подключения поверх bufconn
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

	testUser := model.Users{
		Login:        "mrechkunov_delete_target",
		PasswordHash: "password_hash_12345",
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name: "Успешно: Пользователь полностью удален с сервера",
			setupMock: func() {
				mockServer.mockErr = nil
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Внутренний сбой СУБД на стороне сервера (Internal)",
			setupMock: func() {
				mockServer.mockErr = status.Error(codes.Internal, "server error user not deleted")
			},
			wantErr:  true,
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Вызываем тестируемую клиентскую функцию удаления
			err := cliservice.DeleteUser(ctx, client, testUser)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем gRPC-код ответа
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус gRPC, получили обычную ошибку: %v", err)
				}
				if st.Code() != tt.wantCode {
					t.Errorf("DeleteUser() код ошибки = %v, ожидали = %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}
