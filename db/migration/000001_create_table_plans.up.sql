-- Create table to store full FinancialPlan document as JSONB
CREATE TABLE IF NOT EXISTS plans (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    data JSONB NOT NULL
);

-- Helpful indexes for common queries (optional)
CREATE INDEX IF NOT EXISTS idx_plans_user_id ON plans(user_id);
CREATE INDEX IF NOT EXISTS idx_plans_created_at ON plans(created_at);
