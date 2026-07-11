-- Agregar columna completed_exercises (JSONB)
ALTER TABLE user_progress 
ADD COLUMN completed_exercises JSONB DEFAULT '[]'::jsonb;

-- Índice para querys rápidos
CREATE INDEX IF NOT EXISTS idx_user_progress_completed_exercises 
ON user_progress USING GIN (completed_exercises);