-- Step 1: Drop the existing unique constraint
ALTER TABLE ema_scores DROP CONSTRAINT IF EXISTS unique_ema_score_entry;

-- Step 2: Add a new unique constraint including 'height'
ALTER TABLE ema_scores ADD CONSTRAINT unique_ema_score_entry UNIQUE (topic_id, type, address, height);