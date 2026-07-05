-- migrations/000001_create_users_table.up.sql
-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    u_login VARCHAR(255) PRIMARY KEY,
    u_password VARCHAR(255) NOT NULL,
    u_token VARCHAR(255) NOT NULL,
    uuid BIGINT NOT NULL
);

-- Базовый индекс для поиска по login
CREATE INDEX idx_ulogin ON users(u_login);