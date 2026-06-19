-- 001: Exercise <-> Lesson Many-to-Many migration
-- 
-- This migration:
-- 1. Adds 'name' column to exercises (GORM creates it via AutoMigrate, this is a safety net)
-- 2. Creates the lesson_exercises pivot table
-- 3. Migrates existing exercise->lesson relationships into the pivot
-- 4. Drops the now-redundant lesson_id and order_index columns from exercises
-- 5. Creates performance indexes
--
-- GORM's AutoMigrate creates the lesson_exercises table and name column,
-- but does NOT drop old columns. SQL is needed for that.

DO $$
BEGIN
    -- Step 1: Add name column if it doesn't exist (safety: GORM usually does this)
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'exercises' AND column_name = 'name'
    ) THEN
        ALTER TABLE exercises ADD COLUMN name VARCHAR(255) NOT NULL DEFAULT 'Sin nombre';
    END IF;

    -- Step 2: Migrate existing data into lesson_exercises (only if lesson_id still exists)
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'exercises' AND column_name = 'lesson_id'
    ) THEN
        INSERT INTO lesson_exercises (lesson_id, exercise_id, order_index)
        SELECT e.lesson_id, e.id, e.order_index
        FROM exercises e
        WHERE e.lesson_id IS NOT NULL
        ON CONFLICT (lesson_id, exercise_id) DO NOTHING;
    END IF;

    -- Step 3: Drop old FK constraint (if exists) before dropping columns
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'fk_exercises_lesson'
    ) THEN
        ALTER TABLE exercises DROP CONSTRAINT fk_exercises_lesson;
    END IF;

    -- Step 4: Drop lesson_id and order_index from exercises (they live in the pivot now)
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'exercises' AND column_name = 'lesson_id'
    ) THEN
        ALTER TABLE exercises DROP COLUMN lesson_id;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'exercises' AND column_name = 'order_index'
    ) THEN
        ALTER TABLE exercises DROP COLUMN order_index;
    END IF;

    -- Step 5: Indexes for performance
    CREATE INDEX IF NOT EXISTS idx_lesson_exercises_lesson ON lesson_exercises(lesson_id, order_index);
    CREATE INDEX IF NOT EXISTS idx_lesson_exercises_exercise ON lesson_exercises(exercise_id);

END $$;
