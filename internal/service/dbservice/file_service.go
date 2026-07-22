package dbservice

import (
	"context"
	"os"

	"github.com/mrechkunov/goKeeper.git/internal/config"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/repository"
)

// AddFile insert data in storage
func AddFile(ctx context.Context, data model.File) error {
	fileStorage := repository.NewFileStorage(config.DBconn)
	return fileStorage.InsertFile(ctx, data)
}

// GetFile return data from storage selected by login & metadata
func GetFile(ctx context.Context, login, metadata string) (data model.File, err error) {
	fileStorage := repository.NewFileStorage(config.DBconn)
	return fileStorage.SelectFile(ctx, login, metadata)
}

// EditFile Edit file data in DB
func EditFile(ctx context.Context, dataIn model.File) error {
	fileStorage := repository.NewFileStorage(config.DBconn)
	return fileStorage.UpdateFile(ctx, dataIn)
}

// DeleteFile delete row with file data by login and metadata
func DeleteFile(ctx context.Context, data model.File) error {
	fileStorage := repository.NewFileStorage(config.DBconn)
	return fileStorage.DeleteFile(ctx, data)
}

// DeleteAllUserFiles delete all records by login
func DeleteAllUserFiles(ctx context.Context, login string) error {
	fileStorage := repository.NewFileStorage(config.DBconn)
	pathList, err := fileStorage.DeleteAllFilesByLogin(ctx, login)
	if err != nil {
		logger.Log.Warnln("error while delete user file data", err)
		return err
	}
	// удалим файлы с диска
	for _, fp := range pathList {
		err = os.Remove(fp)
		if err != nil {
			logger.Log.Warnln("error while delete file:", fp, "from hdd", err)
		}
	}
	return nil
}
