-- Migration: 006_create_media_table
-- Description: Rollback media table creation
-- Down Migration

DROP TRIGGER IF EXISTS update_media_updated_at ON media;
DROP INDEX IF EXISTS idx_media_storage_type;
DROP INDEX IF EXISTS idx_media_created_at;
DROP INDEX IF EXISTS idx_media_mime_type;
DROP INDEX IF EXISTS idx_media_post_id;
DROP INDEX IF EXISTS idx_media_user_id;
DROP TABLE IF EXISTS media;
