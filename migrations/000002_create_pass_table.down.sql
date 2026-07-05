-- migrations/000002_create_pass_table.down.sql
-- Откат создания таблицы паролей
DROP INDEX IF EXISTS idx_uuid;
DROP TABLE IF EXISTS passwords;