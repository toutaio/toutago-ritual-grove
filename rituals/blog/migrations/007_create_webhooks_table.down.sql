-- Migration: 007_create_webhooks_table
-- Description: Rollback webhooks table creation
-- Down Migration

DROP INDEX IF EXISTS idx_webhook_deliveries_created_at;
DROP INDEX IF EXISTS idx_webhook_deliveries_event_type;
DROP INDEX IF EXISTS idx_webhook_deliveries_webhook_id;
DROP TABLE IF EXISTS webhook_deliveries;
DROP TRIGGER IF EXISTS update_webhooks_updated_at ON webhooks;
DROP INDEX IF EXISTS idx_webhooks_created_at;
DROP INDEX IF EXISTS idx_webhooks_events;
DROP INDEX IF EXISTS idx_webhooks_enabled;
DROP TABLE IF EXISTS webhooks;
