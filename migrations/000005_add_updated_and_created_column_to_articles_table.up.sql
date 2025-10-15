-- 000005_add_updated_and_created_column_to_articles_table.up.sql

ALTER TABLE domain.articles
    ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
