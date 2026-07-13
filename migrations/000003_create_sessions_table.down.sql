-- migrations/000003_create_sessions_table.down.sql
-- Откат создания таблицы сессий пользователей
DROP INDEX IF EXISTS idx_token;
DROP TABLE IF EXISTS users_sessions;