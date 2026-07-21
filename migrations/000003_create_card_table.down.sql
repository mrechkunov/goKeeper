-- migrations/000003_create_card_table.down.sql
-- Откат создания таблицы данных карт
DROP INDEX IF EXISTS idx_login;
DROP TABLE IF EXISTS cards;