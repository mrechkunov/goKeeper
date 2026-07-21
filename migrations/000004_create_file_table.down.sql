-- migrations/000004_create_file_table.down.sql
-- Откат создания таблицы данных о файлах
DROP INDEX IF EXISTS idx_meta;
DROP TABLE IF EXISTS files;