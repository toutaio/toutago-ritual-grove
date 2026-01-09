-- Migration: 004_update_posts_table
-- Description: Rollback posts table updates
-- Down Migration

DROP INDEX IF EXISTS idx_posts_deleted_at;
DROP INDEX IF EXISTS idx_posts_author_id;
ALTER TABLE posts DROP CONSTRAINT IF EXISTS posts_author_fk;
ALTER TABLE posts DROP COLUMN IF EXISTS deleted_at;
