-- 000005_add_updated_and_created_column_to_articles_table.down.sql

ALTER TABLE domain.articles
    DROP COLUMN created_at,
    DROP COLUMN updated_at;
