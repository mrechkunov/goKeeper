package service

import (
	"context"
	"os"

	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SaveFile save file data on server
func (gk *GoKeeperServer) SaveFile(ctx context.Context, in *pb.FileData) (out *pb.EmptyMessage, err error) {
	data := model.File{
		CipherData: in.GetCipherdata(),
		MetaData:   in.GetMetadata(),
		FileName:   in.GetFilename(),
	}

	// создать папку для пользователя если ее нет
	dirPath := "./upload/" + data.UserLogin
	err = os.MkdirAll(dirPath, 0755) // 0755 — права доступа для владельца, группы и остальных
	if err != nil {
		logger.Log.Warnln("error while make dir", err)
		return out, err
	}
	// открываем (создаем, если нет) файл с правами чтения/записи
	filePath := "./upload/" + data.UserLogin + "/" + data.FileName
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Log.Warnln("error while open file", err)
		return out, err
	}
	defer file.Close()
	// записываем байты в файл
	_, err = file.Write(data.CipherData)
	if err != nil {
		logger.Log.Warnln("error while write in file", err)
		return out, err
	}
	data.FilePath = filePath
	data.CipherData = nil
	// сходить в БД и добавить запись
	if err = db.AddFile(ctx, data); err != nil {
		logger.Log.Warnln("Error while save file in db", err)
		return out, status.Error(codes.AlreadyExists, "server error file not saved")
	}
	return out, nil
}

// GetFile return File data from server
func (gk *GoKeeperServer) GetFile(ctx context.Context, in *pb.FileData) (out *pb.FileData, err error) {
	// читаем данные из бд
	data, err := db.GetFile(ctx, in.GetLogin(), in.GetMetadata())
	if err != nil {
		logger.Log.Warnln("error while get file data", err)
	}
	// читаем данные из папки
	data.CipherData, err = os.ReadFile(data.FilePath)
	if err != nil {
		logger.Log.Warnln("error while reading file on server", err)
		return out, err
	}
	// подготавливаем и отправляем ответ
	out = pb.FileData_builder{
		Filename:   &data.FileName,
		Metadata:   &data.MetaData,
		Cipherdata: data.CipherData,
		Login:      &data.UserLogin,
	}.Build()
	return out, nil
}

// DeleteFile delete file data from storage and db
func (gk *GoKeeperServer) DeleteFile(ctx context.Context, in *pb.FileData) (out *pb.EmptyMessage, err error) {
	data := model.File{
		UserLogin: in.GetLogin(),
		MetaData:  in.GetMetadata(),
	}
	// сходим в бд и возьмем данные о файле
	inDBData, err := db.GetFile(ctx, data.UserLogin, data.MetaData)
	// удалим файл с диска
	err = os.Remove(inDBData.FilePath)
	if err != nil {
		logger.Log.Warnln("error while delete file from hdd", err)
		return out, nil
	}
	// удаляем запись в БД
	err = db.DeleteFile(ctx, data)
	if err != nil {
		logger.Log.Warnln("error while delete file", err)
		return out, err
	}
	return out, nil
}
