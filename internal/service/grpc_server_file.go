package service

import (
	"context"
	"os"
	"path/filepath"

	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/service/dbservice"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SaveFile save file data on server
func (gk *GoKeeperServer) SaveFile(ctx context.Context, in *pb.FileData) (*pb.EmptyMessage, error) {
	// Безопасно извлекаем имя файла (без путей типа ../)
	safeFileName := filepath.Base(in.GetFilename())
	if safeFileName == "." || safeFileName == "/" {
		return nil, status.Error(codes.InvalidArgument, "invalid filename")
	}

	data := model.File{
		UserLogin:  in.GetLogin(),
		CipherData: in.GetCipherdata(),
		MetaData:   in.GetMetadata(),
		FileName:   safeFileName,
	}

	// Создаем папку для пользователя если ее нет
	dirPath := filepath.Join(".", "upload", data.UserLogin)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		logger.Log.Warnln("error while make dir", err)
		return nil, err
	}

	// Открываем файл с правами чтения/записи
	filePath := filepath.Join(dirPath, data.FileName)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Log.Warnln("error while open file", err)
		return nil, err
	}
	defer file.Close()

	// Записываем байты в файл
	_, err = file.Write(data.CipherData)
	if err != nil {
		logger.Log.Warnln("error while write in file", err)
		return nil, err
	}

	data.FilePath = filePath
	data.CipherData = nil // Обнуляем байты перед отправкой структуры в БД, как в исходном коде

	// Сходить в БД и добавить запись
	if err = dbservice.AddFile(ctx, data); err != nil {
		logger.Log.Warnln("Error while save file in db", err)
		// удаляем файл с диска, если запись в БД сорвалась
		os.Remove(filePath)
		return nil, status.Error(codes.AlreadyExists, "server error file not saved")
	}
	return &pb.EmptyMessage{}, nil
}

// GetFile return File data from server
func (gk *GoKeeperServer) GetFile(ctx context.Context, in *pb.FileData) (*pb.FileData, error) {
	// читаем данные из бд
	data, err := dbservice.GetFile(ctx, in.GetLogin(), in.GetMetadata())
	if err != nil {
		logger.Log.Warnln("error while get file data", err)
		return nil, err
	}

	// читаем данные из папки
	cipherData, err := os.ReadFile(data.FilePath)
	if err != nil {
		logger.Log.Warnln("error while reading file on server", err)
		return nil, err
	}

	// подготавливаем и отправляем ответ (без именованных параметров)
	out := pb.FileData_builder{
		Filename:   &data.FileName,
		Metadata:   &data.MetaData,
		Cipherdata: cipherData,
		Login:      &data.UserLogin,
	}.Build()

	return out, nil
}

// DeleteFile delete file data from storage and db
func (gk *GoKeeperServer) DeleteFile(ctx context.Context, in *pb.FileData) (*pb.EmptyMessage, error) {
	data := model.File{
		UserLogin: in.GetLogin(),
		MetaData:  in.GetMetadata(),
	}

	// 1. Сходим в бд и возьмем данные о файле
	inDBData, err := dbservice.GetFile(ctx, data.UserLogin, data.MetaData)
	if err != nil {
		logger.Log.Warnln("error while get file data from db for deletion", err)
		return nil, err // Досрочный возврат при ошибке БД
	}

	// 2. Удалим файл с диска
	err = os.Remove(inDBData.FilePath)
	if err != nil && !os.IsNotExist(err) { // Если файла уже нет на диске, не паникуем, идем дальше
		logger.Log.Warnln("error while delete file from hdd", err)
		return nil, status.Error(codes.Internal, "failed to delete physical file")
	}

	// 3. Удаляем запись в БД
	err = dbservice.DeleteFile(ctx, data)
	if err != nil {
		logger.Log.Warnln("error while delete file record from db", err)
		return nil, err
	}

	return &pb.EmptyMessage{}, nil
}

// EditFile изменяет содержимое или метаданные существующего файла
func (gk *GoKeeperServer) EditFile(ctx context.Context, in *pb.FileData) (*pb.EmptyMessage, error) {
	safeFileName := filepath.Base(in.GetFilename())
	if safeFileName == "." || safeFileName == "/" {
		return nil, status.Error(codes.InvalidArgument, "invalid filename")
	}

	data := model.File{
		UserLogin:  in.GetLogin(),
		CipherData: in.GetCipherdata(),
		MetaData:   in.GetMetadata(),
		FileName:   safeFileName,
	}

	// 1. Сначала проверяем в СУБД, существует ли вообще редактируемый файл
	inDBData, err := dbservice.GetFile(ctx, data.UserLogin, data.MetaData)
	if err != nil {
		logger.Log.Warnln("error while finding file for edit", err)
		return nil, status.Error(codes.NotFound, "file not found in database")
	}

	// Использовать старый путь или сгенерировать новый, если имя файла изменилось
	data.FilePath = inDBData.FilePath

	// 2. Перезаписываем бинарные данные (cipherdata) на диске
	file, err := os.OpenFile(data.FilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Log.Warnln("error while opening file for editing", err)
		return nil, status.Error(codes.Internal, "failed to write changes to disk")
	}
	defer file.Close()

	_, err = file.Write(data.CipherData)
	if err != nil {
		logger.Log.Warnln("error while writing updated bytes to file", err)
		return nil, status.Error(codes.Internal, "failed to save updated content")
	}

	data.CipherData = nil // Обнуляем байты перед отправкой структуры в репозиторий

	// 3. Обновляем запись о файле в базе данных
	err = dbservice.EditFile(ctx, data)
	if err != nil {
		logger.Log.Warnln("error while updating file data in db", err)
		return nil, status.Error(codes.Internal, "failed to update file registry in database")
	}
	return &pb.EmptyMessage{}, nil
}
