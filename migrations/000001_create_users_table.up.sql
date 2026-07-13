-- migrations/000001_create_users_table.up.sql
-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    u_login VARCHAR(255) NOT NULL,
    u_password_hash VARCHAR(255) NOT NULL,
    u_uuid VARCHAR(255) NOT NULL,
    PRIMARY KEY (u_login, u_uuid)

);

-- Базовый индекс для поиска по login
CREATE INDEX idx_ulogin ON users(u_login);