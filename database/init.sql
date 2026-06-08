-- Configurar pg_cron para usar la base de datos voxlab_db
ALTER SYSTEM SET cron.database_name TO 'voxlab_db';

-- Habilitar extensión pg_cron para jobs programados
CREATE EXTENSION IF NOT EXISTS pg_cron;

-- Crear job para limpiar reacciones de usuarios mayores a 30 días
-- Se ejecuta el día 1 de cada mes a las 00:00
-- (La tabla user_reactions será creada por el backend Go mediante AutoMigrate)
SELECT cron.schedule(
    'cleanup-old-reactions',
    '0 0 1 * *',
    $$DELETE FROM user_reactions WHERE created_at < NOW() - INTERVAL '30 days'$$
);
