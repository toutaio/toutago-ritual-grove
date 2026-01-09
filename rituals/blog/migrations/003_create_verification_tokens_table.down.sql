-- Migration: 003_create_verification_tokens_table
-- Description: Rollback verification_tokens table creation
-- Down Migration

DROP INDEX IF EXISTS idx_verification_tokens_created_at;
DROP INDEX IF EXISTS idx_verification_tokens_expires_at;
DROP INDEX IF EXISTS idx_verification_tokens_type;
DROP INDEX IF EXISTS idx_verification_tokens_token;
DROP INDEX IF EXISTS idx_verification_tokens_user_id;
DROP TABLE IF EXISTS verification_tokens;
