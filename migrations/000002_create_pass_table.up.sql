-- migrations/000002_create_pass_table.up.sql
-- Создание таблицы для хранения паролей
CREATE TABLE IF NOT EXISTS passwords (
    uuid VARCHAR(255) PRIMARY KEY,
    pair VARCHAR(255) NOT NULL,
    metadata VARCHAR(255) NOT NULL
);

-- Базовый индекс для поиска по uuid
CREATE INDEX idx_uuid ON passwords(uuid);