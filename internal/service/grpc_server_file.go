package service

import (
	"context"
	"os"

	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/service/db"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SaveFile save file data on server
func (gk *GoKeeperServer) SaveFile(ctx context.Context, in *pb.FileData) (out *pb.EmptyMessage, err error) {
	data := model.File{
		FilePath:   in.GetFilename(),
		CipherData: in.GetCipherdata(),
		MetaData:   in.GetMetadata(),
		FileName:   in.GetFilename(),
	}
	// сходить в БД и добавить запись
	if err = db.AddFile(ctx, data); err != nil {
		logger.Log.Warnln("Error while save file in db", err)
		return out, status.Error(codes.AlreadyExists, "server error file not saved")
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
	return out, nil
}

// // GetCard return card data from storage
// func (gk *GoKeeperServer) GetCard(ctx context.Context, in *pb.CardData) (out *pb.CardData, err error) {
// 	data, err := db.GetCard(ctx, in.GetLogin(), in.GetMetadata())
// 	if err != nil {
// 		logger.Log.Warnln("error while get card data", err)
// 	}
// 	out = pb.CardData_builder{
// 		Login:      &data.UserLogin,
// 		Cipherdata: &data.CipherData,
// 		Metadata:   &data.MetaData,
// 	}.Build()
// 	return out, nil
// }

// // EditCard edit card data in storage
// func (gk *GoKeeperServer) EditCard(ctx context.Context, in *pb.CardData) (out *pb.EmptyMessage, err error) {
// 	data := model.Cards{
// 		UserLogin:  in.GetLogin(),
// 		CipherData: in.GetCipherdata(),
// 		MetaData:   in.GetMetadata(),
// 	}
// 	err = db.EditCard(ctx, data)
// 	if err != nil {
// 		logger.Log.Warnln("error while edit card data", err)
// 		return out, err
// 	}
// 	return out, nil
// }

// // DeleteCard delete card data from storage
// func (gk *GoKeeperServer) DeleteCard(ctx context.Context, in *pb.CardData) (out *pb.EmptyMessage, err error) {
// 	data := model.Cards{
// 		UserLogin:  in.GetLogin(),
// 		CipherData: in.GetCipherdata(),
// 		MetaData:   in.GetMetadata(),
// 	}
// 	err = db.DeleteCard(ctx, data)
// 	if err != nil {
// 		logger.Log.Warnln("error while delete card", err)
// 		return out, err
// 	}
// 	return out, nil
// }
