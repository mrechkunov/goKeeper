package service_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/service"
	"github.com/mrechkunov/goKeeper.git/internal/service/dbservice"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGoKeeperServer_SaveFile(t *testing.T) {
	server := &service.GoKeeperServer{}
	ctx := context.Background()

	// Изолируем dbservice
	oldAddFile := dbservice.AddFile
	defer func() { dbservice.AddFile = oldAddFile }()

	// Перенаправляем создание папок во временную директорию ОС, чтобы не мусорить в проекте
	tmpDir := t.TempDir()

	// Для тестов подменим рабочую директорию на временную (через изменение путей в коде,
	// но так как в коде зашит префикс "./upload", мы просто временно перейдем в t.TempDir)
	oldWd, _ := os.Getwd()
	err := os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Не удалось переключить рабочую директорию для теста: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }() // Возвращаем рабочую директорию назад

	testLogin := "mrechkunov_user"
	testFileName := "secret_report.enc"
	testBytes := []byte("encrypted_file_payload_aes_gcm")
	testMeta := "financial_data_2026"
	duplicateFile := "duplicate_file.dat"
	tests := []struct {
		name         string
		req          *pb.FileData
		mockAct      func()
		wantErr      bool
		expectedCode codes.Code
	}{
		{
			name: "Успешно: Файл сохранен на диск и запись добавлена в БД",
			req: pb.FileData_builder{
				Login:      &testLogin,
				Filename:   &testFileName,
				Cipherdata: testBytes,
				Metadata:   &testMeta,
			}.Build(),
			mockAct: func() {
				dbservice.AddFile = func(ctx context.Context, data model.File) error {
					// Проверяем, что в БД уходят корректные параметры
					if data.UserLogin != testLogin || data.FileName != testFileName || data.MetaData != testMeta {
						return errors.New("mock db: data mismatch")
					}
					// Проверяем, что файл физически создался на диске в правильном месте
					expectedPath := filepath.Join("upload", testLogin, testFileName)
					if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
						return errors.New("mock db: file was not written to disk before db call")
					}
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Запись уже есть в БД (Файл должен удалиться с диска)",
			req: pb.FileData_builder{
				Login:      &testLogin,
				Filename:   &duplicateFile,
				Cipherdata: testBytes,
				Metadata:   &testMeta,
			}.Build(),
			mockAct: func() {
				dbservice.AddFile = func(ctx context.Context, data model.File) error {
					return errors.New("duplicate key in postgres")
				}
			},
			wantErr:      true,
			expectedCode: codes.AlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockAct()

			resp, err := server.SaveFile(ctx, tt.req)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("SaveFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Проверяем gRPC статус-код при ошибке
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус gRPC, получили: %v", err)
				}
				if st.Code() != tt.expectedCode {
					t.Errorf("SaveFile() код ошибки = %v, ожидали = %v", st.Code(), tt.expectedCode)
				}
				return
			}

			// 3. При успехе проверяем валидность EmptyMessage
			if !tt.wantErr {
				if resp == nil {
					t.Fatal("SaveFile() вернул nil вместо &pb.EmptyMessage{}")
				}
			}
		})
	}
}

func TestGoKeeperServer_GetFile(t *testing.T) {
	server := &service.GoKeeperServer{}
	ctx := context.Background()

	// Изолируем dbservice
	oldGetFile := dbservice.GetFile
	defer func() { dbservice.GetFile = oldGetFile }()

	// Создаем временную директорию ОС для тестовых файлов
	tmpDir := t.TempDir()

	testLogin := "mrechkunov_client"
	testFileName := "backup.key"
	testMetadata := "server_keys"
	testContent := []byte("secret_binary_content_aes_gcm")
	missingMeta := "missing_meta"
	deletedMeta := "deleted_meta"
	// Физически создаем файл во временной папке для успешного сценария
	validFilePath := filepath.Join(tmpDir, testFileName)
	err := os.WriteFile(validFilePath, testContent, 0644)
	if err != nil {
		t.Fatalf("Не удалось подготовить тестовый файл на диске: %v", err)
	}

	tests := []struct {
		name         string
		req          *pb.FileData
		mockAct      func()
		wantFilename string
		wantMetadata string
		wantContent  []byte
		wantLogin    string
		wantErr      bool
		expectedErr  string
	}{
		{
			name: "Успешно: Файл найден в БД и прочитан с диска",
			req: pb.FileData_builder{
				Login:    &testLogin,
				Metadata: &testMetadata,
			}.Build(),
			mockAct: func() {
				dbservice.GetFile = func(ctx context.Context, login, metadata string) (model.File, error) {
					if login == testLogin && metadata == testMetadata {
						return model.File{
							UserLogin: testLogin,
							FileName:  testFileName,
							MetaData:  testMetadata,
							FilePath:  validFilePath, // Путь указывает на наш временный файл
						}, nil
					}
					return model.File{}, errors.New("not found")
				}
			},
			wantFilename: testFileName,
			wantMetadata: testMetadata,
			wantContent:  testContent,
			wantLogin:    testLogin,
			wantErr:      false,
		},
		{
			name: "Ошибка: Запись о файле отсутствует в базе данных",
			req: pb.FileData_builder{
				Login:    &testLogin,
				Metadata: &missingMeta,
			}.Build(),
			mockAct: func() {
				dbservice.GetFile = func(ctx context.Context, login, metadata string) (model.File, error) {
					return model.File{}, errors.New("file record not found in DB")
				}
			},
			wantErr:     true,
			expectedErr: "file record not found in DB",
		},
		{
			name: "Ошибка: Запись в БД есть, но сам файл физически удален с диска",
			req: pb.FileData_builder{
				Login:    &testLogin,
				Metadata: &deletedMeta,
			}.Build(),
			mockAct: func() {
				dbservice.GetFile = func(ctx context.Context, login, metadata string) (model.File, error) {
					return model.File{
						UserLogin: testLogin,
						FileName:  "ghost.bin",
						MetaData:  "deleted_meta",
						FilePath:  filepath.Join(tmpDir, "non_existent_ghost_file.bin"), // Файла нет на диске
					}, nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockAct()

			// Вызываем тестируемый метод получения файла
			resp, err := server.GetFile(ctx, tt.req)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ошибка ожидалась и мы знаем её точный текст, проверяем его
			if tt.wantErr && err != nil && tt.expectedErr != "" {
				if err.Error() != tt.expectedErr {
					t.Errorf("GetFile() error text = %q, want %q", err.Error(), tt.expectedErr)
				}
				return
			}

			// 3. Если сценарий успешный, проверяем корректность сборки ответа из builder
			if !tt.wantErr {
				if resp.GetFilename() != tt.wantFilename {
					t.Errorf("GetFile() resp.Filename = %q, want %q", resp.GetFilename(), tt.wantFilename)
				}
				if resp.GetMetadata() != tt.wantMetadata {
					t.Errorf("GetFile() resp.Metadata = %q, want %q", resp.GetMetadata(), tt.wantMetadata)
				}
				if string(resp.GetCipherdata()) != string(tt.wantContent) {
					t.Errorf("GetFile() resp.Cipherdata = %q, want %q", string(resp.GetCipherdata()), string(tt.wantContent))
				}
				if resp.GetLogin() != tt.wantLogin {
					t.Errorf("GetFile() resp.Login = %q, want %q", resp.GetLogin(), tt.wantLogin)
				}
			}
		})
	}
}

func TestGoKeeperServer_DeleteFile(t *testing.T) {
	server := &service.GoKeeperServer{}
	ctx := context.Background()

	// Изолируем dbservice
	oldGetFile := dbservice.GetFile
	oldDeleteFile := dbservice.DeleteFile
	defer func() {
		dbservice.GetFile = oldGetFile
		dbservice.DeleteFile = oldDeleteFile
	}()

	tmpDir := t.TempDir()
	testLogin := "mrechkunov_user"
	testMetadata := "delete_meta"
	testFileName := "to_be_deleted.txt"
	unknownMeta := "unknown_meta"
	// Готовим физический файл на диске во временной папке
	filePath := filepath.Join(tmpDir, testFileName)

	tests := []struct {
		name         string
		req          *pb.FileData
		mockAct      func()
		wantErr      bool
		expectedCode codes.Code
	}{
		{
			name: "Успешно: Файл удален с диска и запись стерта из БД",
			req: pb.FileData_builder{
				Login:    &testLogin,
				Metadata: &testMetadata,
			}.Build(),
			mockAct: func() {
				// Создаем файл перед запуском кейса
				_ = os.WriteFile(filePath, []byte("temp"), 0644)

				dbservice.GetFile = func(ctx context.Context, login, metadata string) (model.File, error) {
					return model.File{FilePath: filePath}, nil
				}
				dbservice.DeleteFile = func(ctx context.Context, data model.File) error {
					return nil // Успешное удаление из БД
				}
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Файл не найден в базе данных",
			req: pb.FileData_builder{
				Login:    &testLogin,
				Metadata: &unknownMeta,
			}.Build(),
			mockAct: func() {
				dbservice.GetFile = func(ctx context.Context, login, metadata string) (model.File, error) {
					return model.File{}, errors.New("file not found in db")
				}
			},
			wantErr:      true,
			expectedCode: codes.Unknown,
		},
		{
			name: "Ошибка: Сбой СУБД на этапе удаления записи",
			req: pb.FileData_builder{
				Login:    &testLogin,
				Metadata: &testMetadata,
			}.Build(),
			mockAct: func() {
				// Создаем файл
				_ = os.WriteFile(filePath, []byte("temp"), 0644)

				dbservice.GetFile = func(ctx context.Context, login, metadata string) (model.File, error) {
					return model.File{FilePath: filePath}, nil
				}
				dbservice.DeleteFile = func(ctx context.Context, data model.File) error {
					return errors.New("db delete crash")
				}
			},
			wantErr:      true,
			expectedCode: codes.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockAct()

			resp, err := server.DeleteFile(ctx, tt.req)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("DeleteFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если упало по плану, проверяем статус код (если это gRPC статус)
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if ok && tt.expectedCode != codes.Unknown && st.Code() != tt.expectedCode {
					t.Errorf("DeleteFile() код ошибки = %v, ожидали = %v", st.Code(), tt.expectedCode)
				}
				return
			}

			// 3. При успехе проверяем, что файла больше нет на диске и возвращен EmptyMessage
			if !tt.wantErr {
				if resp == nil {
					t.Fatal("DeleteFile() вернул nil вместо &pb.EmptyMessage{}")
				}
				if _, err := os.Stat(filePath); !os.IsNotExist(err) {
					t.Error("Физический файл остался на диске после успешного вызова DeleteFile")
				}
			}
		})
	}
}

func TestGoKeeperServer_EditFile(t *testing.T) {
	server := &service.GoKeeperServer{}
	ctx := context.Background()

	// Изолируем dbservice
	oldGetFile := dbservice.GetFile
	oldEditFile := dbservice.EditFile
	defer func() {
		dbservice.GetFile = oldGetFile
		dbservice.EditFile = oldEditFile
	}()

	tmpDir := t.TempDir()
	testLogin := "mrechkunov_user"
	testMetadata := "important_doc"
	testFileName := "document.enc"
	ghostMeta := "ghost_meta"
	filePath := filepath.Join(tmpDir, testFileName)
	newContent := []byte("new_updated_encrypted_file_payload_aes_gcm")

	tests := []struct {
		name         string
		req          *pb.FileData
		mockAct      func()
		wantErr      bool
		expectedCode codes.Code
	}{
		{
			name: "Успешно: Содержимое файла и метаданные изменены",
			req: pb.FileData_builder{
				Login:      &testLogin,
				Filename:   &testFileName,
				Cipherdata: newContent,
				Metadata:   &testMetadata,
			}.Build(),
			mockAct: func() {
				// Создаем старую версию файла перед запуском
				_ = os.WriteFile(filePath, []byte("old_payload"), 0644)

				dbservice.GetFile = func(ctx context.Context, login, metadata string) (model.File, error) {
					return model.File{FilePath: filePath, UserLogin: login, MetaData: metadata}, nil
				}
				dbservice.EditFile = func(ctx context.Context, data model.File) error {
					return nil // Успешный апдейт в БД
				}
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Изменяемый файл отсутствует в СУБД (NotFound)",
			req: pb.FileData_builder{
				Login:      &testLogin,
				Filename:   &testFileName,
				Cipherdata: newContent,
				Metadata:   &ghostMeta,
			}.Build(),
			mockAct: func() {
				dbservice.GetFile = func(ctx context.Context, login, metadata string) (model.File, error) {
					return model.File{}, errors.New("record not found")
				}
			},
			wantErr:      true,
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockAct()

			resp, err := server.EditFile(ctx, tt.req)

			// 1. Проверяем флаг ошибки
			if (err != nil) != tt.wantErr {
				t.Fatalf("EditFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Если ожидалась ошибка, проверяем gRPC status-код
			if tt.wantErr && err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Ожидался статус gRPC, получили: %v", err)
				}
				if st.Code() != tt.expectedCode {
					t.Errorf("EditFile() код ошибки = %v, ожидали = %v", st.Code(), tt.expectedCode)
				}
				return
			}

			// 3. При успешном сценарии проверяем запись на диск и возврат EmptyMessage
			if !tt.wantErr {
				if resp == nil {
					t.Fatal("EditFile() вернул nil вместо &pb.EmptyMessage{}")
				}

				// Проверяем, что файл на диске реально перезаписался новыми байтами
				diskBytes, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Не удалось прочитать измененный файл с диска: %v", err)
				}
				if string(diskBytes) != string(newContent) {
					t.Errorf("Данные на диске не обновились. Получили %q, ожидали %q", string(diskBytes), string(newContent))
				}
			}
		})
	}
}
