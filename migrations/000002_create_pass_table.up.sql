-- migrations/000002_create_pass_table.up.sql
-- Создание таблицы для хранения паролей
CREATE TABLE IF NOT EXISTS passwords (
    p_uuid VARCHAR(255) PRIMARY KEY,
    p_pair VARCHAR(255) NOT NULL,
    p_metadata VARCHAR(255) NOT NULL
);

-- Базовый индекс для поиска по uuid
CREATE INDEX idx_uuid ON passwords(p_uuid);