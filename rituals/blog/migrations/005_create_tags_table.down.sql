-- Migration: 005_create_tags_table
-- Description: Rollback tags and post_tags tables creation
-- Down Migration

DROP TRIGGER IF EXISTS update_tags_updated_at ON tags;
DROP INDEX IF EXISTS idx_post_tags_tag_id;
DROP INDEX IF EXISTS idx_post_tags_post_id;
DROP INDEX IF EXISTS idx_tags_slug;
DROP INDEX IF EXISTS idx_tags_name;
DROP TABLE IF EXISTS post_tags;
DROP TABLE IF EXISTS tags;
