package cliservice

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/mrechkunov/goKeeper.git/internal/cryptodata"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	pb "github.com/mrechkunov/goKeeper.git/proto"
)

// SaveFile client service for encrypt and save file data on server
func SaveFile(ctx context.Context, client pb.GoKeeperClient, file model.File) error {
	// узнать размер файла, если он больше чем 4 мб отказать
	fileInfo, err := os.Stat(file.FilePath)
	if err != nil {
		logger.Log.Infoln("error while file info get in os Stat:", err)
		return err
	}

	if fileInfo.Size() > 4000000 {
		err = errors.New("to big file to save")
		logger.Log.Infoln(err)
		return err
	}

	file.FileName = filepath.Base(file.FilePath)

	// прочитать файл в байты
	data, err := os.ReadFile(file.FilePath)
	if err != nil {
		logger.Log.Warnln("error while reading file", err)
		return err
	}

	// зашифровать байты
	file.CipherData, err = cryptodata.CryptoFile(data)
	if err != nil {
		logger.Log.Warnln("error while encrypting file", err)
		return err
	}

	// передать на сервер данные
	dataPb := pb.FileData_builder{
		Filename:   &file.FileName,
		Metadata:   &file.MetaData,
		Cipherdata: file.CipherData,
		Login:      &file.UserLogin,
	}.Build()

	_, err = client.SaveFile(ctx, dataPb)
	if err != nil {
		logger.Log.Warnln("error while save file", err)
		return err
	}
	return nil
}

// GetFile return file data from server and decrypt it
func GetFile(ctx context.Context, client pb.GoKeeperClient, file model.File) (out model.File, err error) {
	FilePb := pb.FileData_builder{
		Filename: &file.FileName,
		Metadata: &file.MetaData,
		Login:    &file.UserLogin,
	}.Build()
	data, err := client.GetFile(ctx, FilePb)
	if err != nil {
		return out, err
	}
	// создать папку для пользователя если ее нет
	dirPath := filepath.Join(".", "download", data.GetLogin())
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		logger.Log.Warnln("error while make dir on client size", err)
		return out, err
	}

	decryptData, err := cryptodata.DecryptFile(data.GetCipherdata())
	if err != nil {
		logger.Log.Warnln("error while decrypting file on client size", err)
		return out, err
	}

	// открываем (создаем, если нет) файл с правами чтения/записи
	filePath := filepath.Join(dirPath, data.GetFilename())
	fileToWrite, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Log.Warnln("error while open file on client size", err)
		return out, err
	}
	defer fileToWrite.Close()

	// записываем байты в файл
	_, err = fileToWrite.Write(decryptData)
	if err != nil {
		logger.Log.Warnln("error while write in file on client size", err)
		return out, err
	}
	out = model.File{
		FileName:  data.GetFilename(),
		FilePath:  filePath,
		MetaData:  data.GetMetadata(),
		UserLogin: data.GetLogin(),
	}
	return out, nil
}

// EditFile client service for encrypt and update file data on server
func EditFile(ctx context.Context, client pb.GoKeeperClient, file model.File) error {
	// Узнаем размер измененного файла, если он больше 4 МБ — отказываем
	fileInfo, err := os.Stat(file.FilePath)
	if err != nil {
		logger.Log.Infoln("error while file info get in os Stat during edit:", err)
		return err
	}
	if fileInfo.Size() > 4000000 {
		err = errors.New("to big file to save")
		logger.Log.Infoln(err)
		return err
	}

	file.FileName = filepath.Base(file.FilePath)

	// Прочитываем обновленный файл в байты
	data, err := os.ReadFile(file.FilePath)
	if err != nil {
		logger.Log.Warnln("error while reading file during edit", err)
		return err
	}

	// Зашифровываем новые байты файла
	file.CipherData, err = cryptodata.CryptoFile(data)
	if err != nil {
		logger.Log.Warnln("error while encrypting file during edit", err)
		return err
	}

	// Подготавливаем и передаем на сервер обновленные данные
	dataPb := pb.FileData_builder{
		Filename:   &file.FileName,
		Metadata:   &file.MetaData,
		Cipherdata: file.CipherData,
		Login:      &file.UserLogin,
	}.Build()

	_, err = client.EditFile(ctx, dataPb)
	if err != nil {
		logger.Log.Warnln("error while edit file on server", err)
		return err
	}
	return nil
}

// DeleteCard delete card data from server
func DeleteFile(ctx context.Context, client pb.GoKeeperClient, file model.File) (err error) {
	filePb := pb.FileData_builder{
		Login:    &file.UserLogin,
		Metadata: &file.MetaData,
	}.Build()
	_, err = client.DeleteFile(ctx, filePb)
	if err != nil {
		return err
	}
	return nil
}
