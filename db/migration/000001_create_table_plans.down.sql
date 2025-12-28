-- Rollback: drop indexes and table created in the UP migration

-- Drop indexes if they exist
DROP INDEX IF EXISTS idx_plans_created_at;
DROP INDEX IF EXISTS idx_plans_user_id;

-- Drop the table
DROP TABLE IF EXISTS plans;

