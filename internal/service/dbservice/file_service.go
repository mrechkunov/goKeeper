package dbservice

import (
	"context"
	"os"

	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/repository"
)

// AddFile insert data in storage
var AddFile = func(ctx context.Context, data model.File) error {
	return repository.R.FileStorage.InsertFile(ctx, data)
}

// GetFile return data from storage selected by login & metadata
var GetFile = func(ctx context.Context, login, metadata string) (data model.File, err error) {
	return repository.R.FileStorage.SelectFile(ctx, login, metadata)
}

// EditFile Edit file data in DB
var EditFile = func(ctx context.Context, dataIn model.File) error {
	return repository.R.FileStorage.UpdateFile(ctx, dataIn)
}

// DeleteFile delete row with file data by login and metadata
var DeleteFile = func(ctx context.Context, data model.File) error {
	return repository.R.FileStorage.DeleteFile(ctx, data)
}

// DeleteAllUserFiles delete all records by login
var DeleteAllUserFiles = func(ctx context.Context, login string) error {
	pathList, err := repository.R.FileStorage.DeleteAllFilesByLogin(ctx, login)
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
