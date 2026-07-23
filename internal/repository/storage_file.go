package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/mrechkunov/goKeeper.git/internal/model"
)

// Бизнес-ошибки, которые может обрабатывать слой выше (сервис/транспорт)
var (
	ErrDataAlreadyExists = errors.New("data already exists in DB")
	ErrDataNotFound      = errors.New("data not found in DB")
)

type StorageFile struct {
	db *sql.DB
}

// NewFileStorage возвращает указатель на сторадж таблицы с файлами
func NewFileStorage(db *sql.DB) *StorageFile {
	return &StorageFile{db: db}
}

// InsertFile добавляет данные файла. Проверка на дубликаты делегирована БД.
func (sf *StorageFile) InsertFile(ctx context.Context, data model.File) error {
	sqlStatement := `INSERT INTO files (f_login, f_file_path, f_metadata) VALUES ($1, $2, $3);`

	_, err := sf.db.ExecContext(ctx, sqlStatement, data.UserLogin, data.FilePath, data.MetaData)
	if err != nil {
		// Проверяем ошибку уникальности Postgres (код 23505 - unique_violation)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrDataAlreadyExists
		}
		return fmt.Errorf("exec insert file: %w", err)
	}
	return nil
}

// SelectFile возвращает данные файла
func (sf *StorageFile) SelectFile(ctx context.Context, login string, metadata string) (model.File, error) {
	var resp model.File
	sqlStatement := `SELECT f_login, f_file_path, f_metadata FROM files WHERE f_login = $1 AND f_metadata = $2;`

	err := sf.db.QueryRowContext(ctx, sqlStatement, login, metadata).
		Scan(&resp.UserLogin, &resp.FilePath, &resp.MetaData)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.File{}, ErrDataNotFound
		}
		return model.File{}, fmt.Errorf("query row select file: %w", err)
	}
	return resp, nil
}

// UpdateFile обновляет путь к файлу за 1 запрос к БД.
func (sf *StorageFile) UpdateFile(ctx context.Context, data model.File) error {
	sqlStatement := `UPDATE files SET f_file_path = $1 WHERE f_login = $2 AND f_metadata = $3;`

	res, err := sf.db.ExecContext(ctx, sqlStatement, data.FilePath, data.UserLogin, data.MetaData)
	if err != nil {
		return fmt.Errorf("exec update file: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrDataNotFound
	}
	return nil
}

// DeleteFile удаляет запись о файле
func (sf *StorageFile) DeleteFile(ctx context.Context, data model.File) error {
	sqlStatement := `DELETE FROM files WHERE f_login = $1 AND f_metadata = $2;`

	res, err := sf.db.ExecContext(ctx, sqlStatement, data.UserLogin, data.MetaData)
	if err != nil {
		return fmt.Errorf("exec delete file: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrDataNotFound
	}
	return nil
}

// DeleteAllFilesByLogin удаляет записи из БД и возвращает список путей.
// Само удаление с диска должно происходить на уровне Service.
func (sf *StorageFile) DeleteAllFilesByLogin(ctx context.Context, login string) ([]string, error) {
	// Сначала забираем пути, чтобы слой выше знал, что стирать с HDD
	paths, err := sf.GetFilesPathByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("get files paths before deleting: %w", err)
	}

	sqlStatement := `DELETE FROM files WHERE f_login = $1;`
	_, err = sf.db.ExecContext(ctx, sqlStatement, login)
	if err != nil {
		return nil, fmt.Errorf("exec delete all files by login: %w", err)
	}
	return paths, nil
}

// GetFilesPathByLogin возвращает массив путей. Исправлены утечки ресурсов СУБД.
func (sf *StorageFile) GetFilesPathByLogin(ctx context.Context, login string) ([]string, error) {
	sqlStatement := `SELECT f_file_path FROM files WHERE f_login = $1;`

	rows, err := sf.db.QueryContext(ctx, sqlStatement, login)
	if err != nil {
		return nil, fmt.Errorf("query context files paths: %w", err)
	}
	defer rows.Close() // Гарантированно закроет rows при любом выходе из функции

	var resp []string
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return nil, fmt.Errorf("scan file path row: %w", err)
		}
		resp = append(resp, r)
	}

	// проверка, не прервался ли цикл из-за ошибки БД
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return resp, nil
}
