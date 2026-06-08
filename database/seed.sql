-- Seed data para Voxlab
-- Este archivo se ejecuta DESPUÉS de que GORM haya creado las tablas mediante AutoMigrate

-- Insertar datos iniciales para ProgressStatus
INSERT INTO progress_statuses (id, name, color_code) VALUES
    (1, 'Not Started', '#6B7280'),
    (2, 'In Progress', '#3B82F6'),
    (3, 'Completed', '#10B981'),
    (4, 'Mastered', '#8B5CF6')
ON CONFLICT (id) DO NOTHING;

-- Insertar títulos gamificados iniciales
INSERT INTO gamified_titles (id, name, description, requirement_condition) VALUES
    (1, 'Primeros Pasos', 'Complete su primera lección', 'lessons_completed >= 1'),
    (2, 'Orador Novato', 'Complete 5 lecciones', 'lessons_completed >= 5'),
    (3, 'Habla Fluida', 'Mantenga una racha de 7 días', 'streak_days >= 7'),
    (4, 'Maestro del Discurso', 'Complete todos los módulos básicos', 'modules_completed >= 10'),
    (5, 'Leyenda de Voxlab', 'Alcance 1000 XP', 'xp >= 1000')
ON CONFLICT (id) DO NOTHING;

-- Insertar tracks iniciales
INSERT INTO tracks (id, title, description, icon_url, is_active) VALUES
    (1, 'Fundamentos de Oratoria', 'Aprende las bases del discurso público efectivo', '🎤', true),
    (2, 'Técnicas de Persuasión', 'Domina el arte de convencer a tu audiencia', '🎯', true),
    (3, 'Manejo del Escenario', 'Controla el espacio y conecta con tu público', '🎭', true)
ON CONFLICT (id) DO NOTHING;

-- Insertar módulos de ejemplo para el track 1
INSERT INTO modules (id, track_id, title, description, order_index) VALUES
    (1, 1, 'Introducción a la Oratoria', 'Conceptos básicos y fundamentos', 1),
    (2, 1, 'Lenguaje Corporal', 'Comunicación no verbal efectiva', 2),
    (3, 1, 'Estructura del Discurso', 'Cómo organizar tus ideas', 3)
ON CONFLICT (id) DO NOTHING;

-- Insertar lecciones de ejemplo
INSERT INTO lessons (id, title, description, estimated_time_seconds) VALUES
    (1, '¿Qué es la Oratoria?', 'Definición e importancia del discurso público', 300),
    (2, 'Tipos de Discurso', 'Informativo, persuasivo y conmemorativo', 420),
    (3, 'Postura y Presencia', 'Proyección corporal en el escenario', 360),
    (4, 'Gestos y Movimientos', 'Uso efectivo de las manos y desplazamiento', 480),
    (5, 'Apertura Impactante', 'Cómo comenzar tu discurso', 300),
    (6, 'Desarrollo del Contenido', 'Estructura lógica de ideas', 540),
    (7, 'Cierre Memorables', 'Terminar con impacto', 300)
ON CONFLICT (id) DO NOTHING;

-- Insertar relación módulo-lección
INSERT INTO module_lessons (module_id, lesson_id, order_index) VALUES
    (1, 1, 1),
    (1, 2, 2),
    (2, 3, 1),
    (2, 4, 2),
    (3, 5, 1),
    (3, 6, 2),
    (3, 7, 3)
ON CONFLICT DO NOTHING;

-- Crear índices para optimizar consultas
CREATE INDEX IF NOT EXISTS idx_user_reactions_created_at ON user_reactions(created_at);
CREATE INDEX IF NOT EXISTS idx_user_reactions_sender ON user_reactions(sender_id);
CREATE INDEX IF NOT EXISTS idx_user_reactions_receiver ON user_reactions(receiver_id);
CREATE INDEX IF NOT EXISTS idx_exercises_lesson ON exercises(lesson_id);
CREATE INDEX IF NOT EXISTS idx_module_lessons_module ON module_lessons(module_id);
CREATE INDEX IF NOT EXISTS idx_module_lessons_lesson ON module_lessons(lesson_id);
CREATE INDEX IF NOT EXISTS idx_user_titles_user ON user_titles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_titles_equipped ON user_titles(is_equipped);
