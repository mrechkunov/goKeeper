-- migrations/000003_create_card_table.up.sql
-- Создание таблицы для хранения данных карт
CREATE TABLE IF NOT EXISTS cards (
    c_login VARCHAR(255) NOT NULL,
    c_chiperdata VARCHAR(255) NOT NULL,
    c_metadata VARCHAR(255) NOT NULL,
    PRIMARY KEY (c_chiperdata, c_metadata)
);

-- Базовый индекс для поиска по login
CREATE INDEX idx_login ON cards(c_login);


	