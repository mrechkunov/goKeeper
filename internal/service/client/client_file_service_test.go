package cliservice_test

import (
	"context"
	"net"
	"os"
	"path/filepath"
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

// Добавляем метод SaveFile в наш mock-сервер для симуляции ответа сервера
func (m *mockGoKeeperServer) SaveFile(ctx context.Context, in *pb.FileData) (*pb.EmptyMessage, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	return &pb.EmptyMessage{}, nil
}

func TestSaveFile_Client(t *testing.T) {
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

	// Создаем временную директорию для тестовых файлов на диске
	tmpDir := t.TempDir()

	// 1. Готовим нормальный файл (1 КБ)
	normalFilePath := filepath.Join(tmpDir, "normal.txt")
	_ = os.WriteFile(normalFilePath, []byte("some encrypted or raw file data payloads"), 0644)

	// 2. Готовим слишком большой файл (более 4 000 000 байт ~ 4.1 МБ)
	bigFilePath := filepath.Join(tmpDir, "too_big.bin")
	bigFile, _ := os.Create(bigFilePath)
	_ = bigFile.Truncate(4500000) // Устанавливаем размер 4.5 МБ
	bigFile.Close()

	tests := []struct {
		name      string
		filePath  string
		setupMock func()
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name:     "Успешно: Файл валидного размера отправлен на сервер",
			filePath: normalFilePath,
			setupMock: func() {
				mockServer.mockErr = nil
			},
			wantErr: false,
		},
		{
			name:     "Ошибка: Файл превышает лимит в 4 МБ (Отклонение на клиенте)",
			filePath: bigFilePath,
			setupMock: func() {
				mockServer.mockErr = nil // До сети даже не дойдет
			},
			wantErr: true,
		},
		{
			name:     "Ошибка: Файл не найден на диске (os.Stat error)",
			filePath: filepath.Join(tmpDir, "ghost_file_not_found.txt"),
			setupMock: func() {
				mockServer.mockErr = nil
			},
			wantErr: true,
		},
		{
			name:     "Ошибка: Сбой gRPC-сервера при сохранении (AlreadyExists)",
			filePath: normalFilePath,
			setupMock: func() {
				mockServer.mockErr = status.Error(codes.AlreadyExists, "server error file not saved")
			},
			wantErr:  true,
			wantCode: codes.AlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			fileInput := model.File{
				UserLogin: "mrechkunov_user",
				FilePath:  tt.filePath,
				MetaData:  "backup_archive",
			}

			// Вызываем тестируемый метод клиента
			err := cliservice.SaveFile(ctx, client, fileInput)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("SaveFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка пришла от gRPC-сервера, сверяем статус-коды
			if tt.wantErr && err != nil && tt.wantCode != codes.OK {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус gRPC, получили обычную ошибку: %v", err)
				}
				if st.Code() != tt.wantCode {
					t.Errorf("SaveFile() код ошибки = %v, ожидали = %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}

// Добавляем метод GetFile в наш mock-сервер для симуляции ответа сервера
func (m *mockGoKeeperServer) GetFile(ctx context.Context, in *pb.FileData) (*pb.FileData, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	fileName := in.GetFilename()
	metaData := in.GetMetadata()
	login := in.GetLogin()

	return pb.FileData_builder{
		Filename:   &fileName,
		Metadata:   &metaData,
		Login:      &login,
		Cipherdata: []byte(m.mockToken), // Используем mockToken для передачи сырых байт ответа
	}.Build(), nil
}

func TestGetFile_Client(t *testing.T) {
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

	// Изолируем создание папок во временную директорию ОС
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	testLogin := "mrechkunov_user"
	testFileName := "downloaded_doc.enc"
	testMetadata := "personal_archive"
	testContent := []byte("clean_unencrypted_file_data_bytes")

	// Генерируем валидный шифротекст через родную утилиту CryptoFile,
	// чтобы внутренний вызов DecryptFile прошел успешно.
	validEncryptedBytes, err := cryptodata.CryptoFile(testContent)
	if err != nil {
		t.Fatalf("Не удалось подготовить зашифрованные тестовые данные: %v", err)
	}

	searchParams := model.File{
		UserLogin: testLogin,
		FileName:  testFileName,
		MetaData:  testMetadata,
	}

	tests := []struct {
		name         string
		setupMock    func()
		wantErr      bool
		wantCode     codes.Code
		checkContent bool
	}{
		{
			name: "Успешно: Файл получен с сервера, расшифрован и сохранен на диск",
			setupMock: func() {
				mockServer.mockToken = string(validEncryptedBytes)
				mockServer.mockErr = nil
			},
			wantErr:      false,
			checkContent: true,
		},
		{
			name: "Ошибка: Запись о файле отсутствует на сервере (NotFound)",
			setupMock: func() {
				mockServer.mockToken = ""
				mockServer.mockErr = status.Error(codes.NotFound, "file not found")
			},
			wantErr:  true,
			wantCode: codes.NotFound,
		},
		{
			name: "Ошибка: Данные получены, но сбой расшифровки (битые байты)",
			setupMock: func() {
				mockServer.mockToken = "totally-corrupted-and-broken-crypto-bytes"
				mockServer.mockErr = nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Вызываем тестируемый метод получения файла
			result, err := cliservice.GetFile(ctx, client, searchParams)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ожидалась ошибка от сервера, сверяем статус-код
			if tt.wantErr && err != nil && tt.wantCode != codes.OK {
				st, ok := status.FromError(err)
				if ok && st.Code() != tt.wantCode {
					t.Errorf("GetFile() код ошибки = %v, ожидали = %v", st.Code(), tt.wantCode)
				}
				return
			}

			// 3. Если всё успешно, проверяем физическое существование файла и его контент
			if !tt.wantErr && tt.checkContent {
				expectedPath := filepath.Join(".", "download", testLogin, testFileName)
				if result.FilePath != expectedPath {
					t.Errorf("GetFile() FilePath = %q, want %q", result.FilePath, expectedPath)
				}

				// Читаем записанный файл с диска и сверяем байты контента
				diskBytes, err := os.ReadFile(result.FilePath)
				if err != nil {
					t.Fatalf("Не удалось прочитать созданный файл с диска: %v", err)
				}
				if string(diskBytes) != string(testContent) {
					t.Errorf("Содержимое сохраненного файла не совпадает. Получили %q, ожидали %q", string(diskBytes), string(testContent))
				}
			}
		})
	}
}

// Добавляем метод EditFile в наш существующий mock-сервер
func (m *mockGoKeeperServer) EditFile(ctx context.Context, in *pb.FileData) (*pb.EmptyMessage, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	return &pb.EmptyMessage{}, nil
}

func TestEditFile_Client(t *testing.T) {
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

	// Создаем временную изолированную директорию ОС
	tmpDir := t.TempDir()

	// 1. Готовим обновленный файл нормального размера (1 КБ)
	normalUpdatePath := filepath.Join(tmpDir, "updated_doc.txt")
	_ = os.WriteFile(normalUpdatePath, []byte("new fresh encrypted file payload logs"), 0644)

	// 2. Готовим слишком крупный файл обновления (4.5 МБ)
	bigUpdatePath := filepath.Join(tmpDir, "too_big_update.bin")
	bigFile, _ := os.Create(bigUpdatePath)
	_ = bigFile.Truncate(4500000)
	bigFile.Close()

	tests := []struct {
		name      string
		filePath  string
		setupMock func()
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name:     "Успешно: Измененный файл валидного размера отправлен на сервер",
			filePath: normalUpdatePath,
			setupMock: func() {
				mockServer.mockErr = nil
			},
			wantErr: false,
		},
		{
			name:     "Ошибка: Файл обновления превышает лимит 4 МБ (Отклонение клиентом)",
			filePath: bigUpdatePath,
			setupMock: func() {
				mockServer.mockErr = nil
			},
			wantErr: true,
		},
		{
			name:     "Ошибка: Сервер вернул NotFound (Редактируемый файл не найден в БД)",
			filePath: normalUpdatePath,
			setupMock: func() {
				mockServer.mockErr = status.Error(codes.NotFound, "file record not found")
			},
			wantErr:  true,
			wantCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			fileInput := model.File{
				UserLogin: "mrechkunov_user",
				FilePath:  tt.filePath,
				MetaData:  "important_archive",
			}

			// Вызываем разработанный метод
			err := cliservice.EditFile(ctx, client, fileInput)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("EditFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка пришла со стороны gRPC-сервера, сверяем статус-коды
			if tt.wantErr && err != nil && tt.wantCode != codes.OK {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус gRPC, получили: %v", err)
				}
				if st.Code() != tt.wantCode {
					t.Errorf("EditFile() код ошибки = %v, ожидали = %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}

// Добавляем метод DeleteFile в наш mock-сервер для симуляции ответа сервера
func (m *mockGoKeeperServer) DeleteFile(ctx context.Context, in *pb.FileData) (*pb.EmptyMessage, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	return &pb.EmptyMessage{}, nil
}

func TestDeleteFile_Client(t *testing.T) {
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

	// Тестовые данные файла для удаления
	testFileData := model.File{
		UserLogin: "mrechkunov_user",
		MetaData:  "old_unused_backup",
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name: "Успешно: Запись файла и сам файл удалены на сервере",
			setupMock: func() {
				mockServer.mockErr = nil
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Удаляемый файл не найден в базе данных (NotFound)",
			setupMock: func() {
				mockServer.mockErr = status.Error(codes.NotFound, "file not found in DB")
			},
			wantErr:  true,
			wantCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Вызываем клиентскую функцию удаления файла
			err := cliservice.DeleteFile(ctx, client, testFileData)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("DeleteFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась, проверяем её gRPC-код ответа
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус gRPC, получили обычную ошибку: %v", err)
				}
				if st.Code() != tt.wantCode {
					t.Errorf("DeleteFile() код ошибки = %v, ожидали = %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}
