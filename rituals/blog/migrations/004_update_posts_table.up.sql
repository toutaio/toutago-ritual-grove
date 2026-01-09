-- Migration: 004_update_posts_table
-- Description: Update posts table with author_id FK, soft deletes, and better constraints
-- Up Migration

-- Add deleted_at column for soft deletes
ALTER TABLE posts ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;

-- Add author_id foreign key if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'posts_author_fk'
    ) THEN
        ALTER TABLE posts ADD CONSTRAINT posts_author_fk 
            FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
END$$;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);
CREATE INDEX IF NOT EXISTS idx_posts_deleted_at ON posts(deleted_at);

-- Add check constraint for published_at (only published posts should have it)
-- This is informational and can be enforced at application level
