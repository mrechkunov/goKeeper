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

// Добавляем метод SaveCard в наш mock-сервер для симуляции ответа сервера
func (m *mockGoKeeperServer) SaveCard(ctx context.Context, in *pb.CardData) (*pb.EmptyMessage, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	return &pb.EmptyMessage{}, nil
}

func TestSaveCard_Client(t *testing.T) {
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

	// Тестовые данные валидной карты
	validCardData := model.Cards{
		UserLogin:  "mrechkunov_user",
		CardNumber: "4532718281828182", // Валидный номер по Луну
		ValidTo:    "12/29",
		CVVCode:    "123",
		MetaData:   "personal_visa_card",
	}

	// Тестовые данные невалидной карты (для провокации ошибки криптографии)
	invalidCardData := model.Cards{
		UserLogin:  "mrechkunov_user",
		CardNumber: "4532718281828183", // Невалидный номер по Луну
		ValidTo:    "12/29",
		CVVCode:    "123",
		MetaData:   "invalid_card",
	}

	tests := []struct {
		name      string
		cardData  model.Cards
		setupMock func()
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name:     "Успешно: Данные карты зашифрованы и сохранены на сервере",
			cardData: validCardData,
			setupMock: func() {
				mockServer.mockErr = nil
			},
			wantErr: false,
		},
		{
			name:     "Ошибка: Номер карты не прошёл проверку алгоритма Луна (Отклонение на клиенте)",
			cardData: invalidCardData,
			setupMock: func() {
				mockServer.mockErr = nil // До сети запрос даже не дойдёт
			},
			wantErr: true,
		},
		{
			name:     "Ошибка: Карта с такими метаданными уже существует на сервере (AlreadyExists)",
			cardData: validCardData,
			setupMock: func() {
				mockServer.mockErr = status.Error(codes.AlreadyExists, "server error card not saved")
			},
			wantErr:  true,
			wantCode: codes.AlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Вызываем тестируемый клиентский метод
			err := cliservice.SaveCard(ctx, client, tt.cardData)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("SaveCard() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка пришла со стороны gRPC-сервера, сверяем статус-коды
			if tt.wantErr && err != nil && tt.wantCode != codes.OK {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус gRPC, получили обычную ошибку: %v", err)
				}
				if st.Code() != tt.wantCode {
					t.Errorf("SaveCard() код ошибки = %v, ожидали = %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}

// Добавляем метод GetCard в наш mock-сервер для симуляции ответа сервера
func (m *mockGoKeeperServer) GetCard(ctx context.Context, in *pb.CardData) (*pb.CardData, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	login := in.GetLogin()
	metaData := in.GetMetadata()
	return pb.CardData_builder{
		Login:      &login,
		Cipherdata: &m.mockToken, // Используем mockToken для передачи зашифрованных данных карты
		Metadata:   &metaData,
	}.Build(), nil
}

func TestGetCard_Client(t *testing.T) {
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

	// Исходные чистые данные банковской карты
	targetNumber := "4532718281828182" // Валидный номер по Луну
	targetValid := "12/29"
	targetCVV := "123"

	// Генерируем валидный шифротекст в Base64 через родную функцию CryptoCard,
	// чтобы внутренний вызов DecryptCard прошел успешно
	validCiphertext, err := cryptodata.CryptoCard(targetNumber, targetValid, targetCVV)
	if err != nil {
		t.Fatalf("Не удалось подготовить зашифрованные тестовые данные карты: %v", err)
	}

	// Поисковый запрос от клиента
	searchFilter := model.Cards{
		UserLogin: "mrechkunov_user",
		MetaData:  "travel_card_credentials",
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantNumber string
		wantValid  string
		wantCVV    string
		wantErr    bool
		wantCode   codes.Code
	}{
		{
			name: "Успешно: Данные карты получены и успешно расшифрованы",
			setupMock: func() {
				mockServer.mockToken = validCiphertext
				mockServer.mockErr = nil
			},
			wantNumber: targetNumber,
			wantValid:  targetValid,
			wantCVV:    targetCVV,
			wantErr:    false,
		},
		{
			name: "Ошибка: Карта не найдена на сервере (NotFound)",
			setupMock: func() {
				mockServer.mockToken = ""
				mockServer.mockErr = status.Error(codes.NotFound, "card not found")
			},
			wantErr:  true,
			wantCode: codes.NotFound,
		},
		{
			name: "Ошибка: Сбой криптографии при расшифровке (битые байты в ответе)",
			setupMock: func() {
				mockServer.mockToken = "totally-broken-invalid-base64-payload"
				mockServer.mockErr = nil
			},
			wantErr: true, // DecryptCard вернет ошибку, и GetCard должен прокинуть её наверх
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Вызываем тестируемый метод получения и дешифрования карты
			result, err := cliservice.GetCard(ctx, client, searchFilter)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCard() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ожидалась ошибка от сервера, сверяем статус-коды gRPC
			if tt.wantErr && err != nil && tt.wantCode != codes.OK {
				st, ok := status.FromError(err)
				if ok && st.Code() != tt.wantCode {
					t.Errorf("GetCard() код ошибки = %v, ожидали = %v", st.Code(), tt.wantCode)
				}
				return
			}

			// 3. Если всё успешно, проверяем полное совпадение всех полей структуры карт
			if !tt.wantErr {
				if result.UserLogin != searchFilter.UserLogin {
					t.Errorf("GetCard() UserLogin = %q, want %q", result.UserLogin, searchFilter.UserLogin)
				}
				if result.MetaData != searchFilter.MetaData {
					t.Errorf("GetCard() MetaData = %q, want %q", result.MetaData, searchFilter.MetaData)
				}
				if result.CardNumber != tt.wantNumber {
					t.Errorf("GetCard() CardNumber = %q, want %q", result.CardNumber, tt.wantNumber)
				}
				if result.ValidTo != tt.wantValid {
					t.Errorf("GetCard() ValidTo = %q, want %q", result.ValidTo, tt.wantValid)
				}
				if result.CVVCode != tt.wantCVV {
					t.Errorf("GetCard() CVVCode = %q, want %q", result.CVVCode, tt.wantCVV)
				}
			}
		})
	}
}

// Добавляем метод EditCard в наш mock-сервер для симуляции ответа сервера
func (m *mockGoKeeperServer) EditCard(ctx context.Context, in *pb.CardData) (*pb.EmptyMessage, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	return &pb.EmptyMessage{}, nil
}

func TestEditCard_Client(t *testing.T) {
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

	// Тестовые данные для изменения карты
	testCardData := model.Cards{
		UserLogin:  "mrechkunov_user",
		CardNumber: "4532718281828182", // Валидный номер по Луну для CryptoCard
		ValidTo:    "11/30",
		CVVCode:    "999",
		MetaData:   "salary_visa_card",
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name: "Успешно: Данные карты зашифрованы и изменены на сервере",
			setupMock: func() {
				mockServer.mockErr = nil
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Изменяемая карта не найдена на сервере (NotFound)",
			setupMock: func() {
				mockServer.mockErr = status.Error(codes.NotFound, "card data does not exist in DB")
			},
			wantErr:  true,
			wantCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Вызываем клиентскую функцию редактирования карты
			err := cliservice.EditCard(ctx, client, testCardData)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("EditCard() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем её gRPC-код ответа
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус gRPC, получили обычную ошибку: %v", err)
				}
				if st.Code() != tt.wantCode {
					t.Errorf("EditCard() код ошибки = %v, ожидали = %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}

// Добавляем метод DeleteCard в наш mock-сервер для симуляции ответа сервера
func (m *mockGoKeeperServer) DeleteCard(ctx context.Context, in *pb.CardData) (*pb.EmptyMessage, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	return &pb.EmptyMessage{}, nil
}

func TestDeleteCard_Client(t *testing.T) {
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

	// Тестовые данные карты для удаления
	testCardData := model.Cards{
		UserLogin: "mrechkunov_user",
		MetaData:  "expired_credit_card",
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name: "Успешно: Данные карты удалены на сервере",
			setupMock: func() {
				mockServer.mockErr = nil
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Удаляемая карта не найдена в базе данных (NotFound)",
			setupMock: func() {
				mockServer.mockErr = status.Error(codes.NotFound, "card not found in DB")
			},
			wantErr:  true,
			wantCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Вызываем клиентскую функцию удаления карты
			err := cliservice.DeleteCard(ctx, client, testCardData)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("DeleteCard() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем её gRPC-код ответа
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус gRPC, получили обычную ошибку: %v", err)
				}
				if st.Code() != tt.wantCode {
					t.Errorf("DeleteCard() код ошибки = %v, ожидали = %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}
