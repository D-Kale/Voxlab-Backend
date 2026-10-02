-- Create user_exercise_progress table
CREATE TABLE IF NOT EXISTS user_exercise_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exercise_id VARCHAR(64) NOT NULL,
    lesson_id INT,
    score INT NOT NULL CHECK (score >= 0 AND score <= 100),
    completed_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (user_id, exercise_id)
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_user_exercise_progress_user ON user_exercise_progress(user_id);
CREATE INDEX IF NOT EXISTS idx_user_exercise_progress_user_completed ON user_exercise_progress(user_id, completed_at);

-- Add columns to users table for efficient streak calculation
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_completed_at TIMESTAMPTZ;