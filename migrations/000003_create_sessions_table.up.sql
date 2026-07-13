-- migrations/000003_create_sessions_table.up.sql
-- Создание таблицы для хранения сессий пользователей
CREATE TABLE IF NOT EXISTS users_sessions (
    us_login VARCHAR(255) PRIMARY KEY,
    us_token VARCHAR(255) NOT NULL
);

-- Базовый индекс для поиска по token
CREATE INDEX idx_token ON users_sessions(us_token);