-- migrations/000004_create_file_table.up.sql
-- Создание таблицы для хранения данных о файлах
CREATE TABLE IF NOT EXISTS files (
    f_login VARCHAR(255) NOT NULL,
    f_file_path VARCHAR(255) NOT NULL,
    f_metadata VARCHAR(255) NOT NULL PRIMARY KEY
);

-- Базовый индекс для поиска по метаданным
CREATE INDEX idx_meta ON files(f_metadata);


	