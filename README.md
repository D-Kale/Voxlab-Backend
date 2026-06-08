# Voxlab Backend

Plataforma educativa de oratoria con progresión gamificada (estilo Duolingo) y prácticas en vivo mediante WebRTC.

## Stack

| Componente | Tecnología | Puerto |
|------------|------------|--------|
| API | Go (Gin) + GORM | `3000` |
| Documentación | Swagger | `3000/swagger/index.html` |
| Base de datos | PostgreSQL 15 + pg_cron | `5432` |
| Cache | Redis 7 | `6379` |
| Admin DB | pgAdmin 4 | `5050` |
| Archivos | MinIO (S3-compatible) | `9000` / `9001` |
| Llamadas | LiveKit | `7880` / `7881` / `7882` |

## Arquitectura del flujo

```
Usuario → Ruta (Track) → Módulo → Lección → Ejercicio (JSONB)
                                                    ↓
                                   +---------- LiveKit Call ----------+
                                   |    Reacciones (30d TTL via cron)  |
                                   +-------- Estados efímeros ---------+
                                                  ↓
                                           Redis (live_sessions,
                                           streak_days, progress cache)
```

### Datos persistentes (PostgreSQL)

Las tablas relacionales guardan la estructura rígida: `users`, `tracks`, `modules`, `lessons`, `module_lessons`, `user_reactions`.

La tabla `exercises` usa **JSONB** para el contenido polimórfico — un ejercicio puede ser una lectura, un quiz, un reto de audio, o un minijuego de oratoria, sin cambiar el esquema.

**pg_cron** ejecuta un job mensual que limpia reacciones viejas (>30 días):
```sql
SELECT cron.schedule('cleanup-old-reactions', '0 0 1 * *',
  $$DELETE FROM user_reactions WHERE created_at < NOW() - INTERVAL '30 days'$$);
```

### Datos efímeros (Redis)

Redis maneja todo lo temporal:
- `live_sessions:{room_id}` — sesiones activas de LiveKit con TTL de 2h
- `user:streak:active:{user_id}` — racha diaria del usuario (TTL 24h)
- `user:progress:{user_id}` — cache de XP y rachas
- `logs:errors` — últimos 1000 errores en cola

### Archivos (MinIO via go-storage)

MinIO almacena medios (audios, videos, avatares) usando `github.com/D-Kale/go-storage` como capa de abstracción S3. Bucket por defecto: `voxlab-media`.

## Cómo levantar el proyecto

```bash
# Clonar y entrar
git clone <repo> && cd backend

# Copiar configuración
cp .env.example .env

# Iniciar todo
docker compose up -d

# Ver logs
docker compose logs -f backend
```

### URLs de las herramientas

| Herramienta | URL | Credenciales |
|-------------|-----|--------------|
| API | http://localhost:3000 | — |
| Swagger | http://localhost:3000/swagger/index.html | — |
| pgAdmin | http://localhost:5050 | `admin@voxlab.com` / `admin` |
| MinIO Console | http://localhost:9001 | `voxlab_minio_admin` / `voxlab_minio_pass_2024` |

### Configurar pgAdmin

1. Abrir http://localhost:5050
2. Login con `admin@voxlab.com` / `admin`
3. Add New Server:
   - **Name**: `Voxlab DB`
   - **Host**: `postgres`
   - **Port**: `5432`
   - **Username**: `voxlab`
   - **Password**: `voxlab_pass_2024`

## Estructura del código

```
cmd/api/main.go              → Entry point (carga config, conecta DB, inicia router)
internal/
  config/config.go           → Variables de entorno
  database/
    postgres.go              → Conexión GORM + migraciones + seed
    redis.go                 → Conexión Redis + sesiones live + streaks
  http/
    router.go               → Definición de rutas (como routes/api.php)
    middleware/
      auth.go               → Validación JWT real
      cors.go               → CORS configurable por entorno
    controllers/
      auth_controller.go      → Login (JWT)
      health_controller.go    → Health check (DB + Redis)
      track_controller.go     → CRUD Tracks
      module_controller.go    → CRUD Modules + LinkLesson
      lesson_controller.go    → CRUD Lessons
      exercise_controller.go  → CRUD Exercises (JSONB)
      progress_controller.go  → Progreso del usuario
    resources/
      response.go           → Respuesta JSON estandarizada
  models/                    → Modelos GORM (User, Track, Module, Lesson, Exercise, UserProgress, etc.)
  repositories/              → Capa de acceso a datos (uno por entidad)
  services/                  → Lógica de negocio (uno por entidad)
  storage/storage.go         → Adapter para go-storage (MinIO/S3)
database/
  Dockerfile.postgres        → PostgreSQL con pg_cron compilado
  init.sql                   → Habilita pg_cron programa cleanup mensual
  seed.sql                   → Datos iniciales (tracks, módulos, lecciones, títulos)
docs/                        → Swagger autogenerado
```

## Endpoints de la API

### Convenciones

- `🔓` = Público (no requiere token)
- `🔒` = Protegido (requiere `Authorization: Bearer <token>`)
- Todas las respuestas tienen formato `{ "success": bool, "data": ... }` o `{ "error": "..." }`
- Swagger interactivo: http://localhost:3000/swagger/index.html

### Salud del sistema
```bash
🔓 GET /api/v1/health
# → { "status": "ok", "timestamp": "2026-01-01T00:00:00Z", "version": "1.0.0",
#     "services": { "postgres": "ok", "redis": "ok" } }
```

### Autenticación
```bash
🔓 POST /api/v1/auth/register                          # Crear cuenta (auto-login)
🔓 POST /api/v1/auth/login                             # Iniciar sesión
🔓 POST /api/v1/auth/logout                            # Revocar token actual
🔒 GET  /api/v1/auth/me                                # Perfil del usuario autenticado
```

**Login:**
```bash
🔓 POST /api/v1/auth/login
Content-Type: application/json
Body: { "email": "user@example.com", "password": "password123" }
# → { "token": "eyJ...", "expires_at": "2026-01-02T00:00:00Z",
#     "user": { "id": "uuid", "name": "...", "email": "...", "xp": 0, "streak_days": 0 } }
```

**Register (auto-login incluido):**
```bash
🔓 POST /api/v1/auth/register
Content-Type: application/json
Body: { "name": "John Doe", "email": "user@example.com", "password": "password123" }
# → 201 Created — misma respuesta que login (token + user)
# → 409 Conflict — si el email ya existe
```

**Logout:** Invalida el token actual agregándolo a una blacklist en Redis.
```bash
🔓 POST /api/v1/auth/logout
Authorization: Bearer <token>
# → { "success": true, "message": "logged out successfully" }
# El mismo token ya no funciona para endpoints protegidos.
```

**Me:** Devuelve el perfil del usuario autenticado.
```bash
🔒 GET /api/v1/auth/me
Authorization: Bearer <token>
# → { "success": true, "data": { "id": "uuid", "name": "...", "email": "...", "xp": 0, "streak_days": 0 } }
# → 401 "token has been revoked" — si se usó después de logout
```

### Tracks (Cursos)
```bash
🔓 GET  /api/v1/tracks              # Listar todos los cursos
🔓 GET  /api/v1/tracks/:id          # Obtener un curso
🔒 POST /api/v1/tracks              # Crear curso
🔒 PUT  /api/v1/tracks/:id          # Actualizar curso
🔒 DELETE /api/v1/tracks/:id        # Eliminar curso
```

### Modules (Módulos)
```bash
🔓 GET  /api/v1/tracks/:id/modules   # Listar módulos de un curso
🔓 GET  /api/v1/modules/:id          # Obtener un módulo
🔒 POST /api/v1/modules              # Crear módulo
🔒 PUT  /api/v1/modules/:id          # Actualizar módulo
🔒 DELETE /api/v1/modules/:id        # Eliminar módulo
🔒 POST /api/v1/modules/:id/lessons  # Vincular lección al módulo
```

### Lessons (Lecciones)
```bash
🔓 GET  /api/v1/modules/:id/lessons   # Listar lecciones de un módulo
🔓 GET  /api/v1/lessons/:id           # Obtener una lección
🔒 POST /api/v1/lessons               # Crear lección
🔒 PUT  /api/v1/lessons/:id           # Actualizar lección
🔒 DELETE /api/v1/lessons/:id         # Eliminar lección
```

### Exercises (Ejercicios — JSONB)
```bash
🔓 GET  /api/v1/lessons/:id/exercises   # Listar ejercicios de una lección
🔓 GET  /api/v1/exercises/:id           # Obtener un ejercicio
🔒 POST /api/v1/exercises               # Crear ejercicio
🔒 PUT  /api/v1/exercises/:id           # Actualizar ejercicio
🔒 DELETE /api/v1/exercises/:id         # Eliminar ejercicio
```

### Progress (Progreso del usuario)
```bash
🔒 GET  /api/v1/progress            # Mi progreso (todas las lecciones)
🔒 POST /api/v1/progress            # Completar una lección (+ XP)
```

## Ejemplos de Creación de Curso

Flujo completo para crear contenido educativo desde cero (con `curl`):

### 1. Crear cuenta o login

**Opción A — Registrarse (primera vez):**
```bash
curl -s -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Juan Pérez","email":"juan@example.com","password":"password123"}'
# → 201 Created + token + user data
```

**Opción B — Login (usuario existente):**
```bash
curl -s -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"juan@example.com","password":"password123"}' \
  | jq .token -r
```
Guardar el token: `export TOKEN="eyJ..."`

### 2. Crear un Track (curso)
```bash
curl -s -X POST http://localhost:3000/api/v1/tracks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Oratoria para Líderes",
    "description": "Aprende a comunicarte con impacto en el escenario",
    "icon_url": "https://cdn.voxlab.com/icons/leadership.png"
  }'
# → { "success": true, "data": { "id": 4, "title": "Oratoria para Líderes", ... } }
```

### 3. Crear Módulos
```bash
curl -s -X POST http://localhost:3000/api/v1/modules \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "track_id": 4,
    "title": "Voz y Proyección",
    "description": "Técnicas para proyectar la voz sin esfuerzo",
    "order_index": 1
  }'
```

### 4. Crear Lecciones
```bash
curl -s -X POST http://localhost:3000/api/v1/lessons \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Respiración Diafragmática",
    "description": "Aprende a respirar desde el diafragma",
    "estimated_time_seconds": 300
  }'
# → { "success": true, "data": { "id": 7, "title": "Respiración Diafragmática", ... } }
```

### 5. Vincular Lección al Módulo
```bash
curl -s -X POST http://localhost:3000/api/v1/modules/4/lessons \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{ "lesson_id": 7 }'
# → { "success": true, "message": "Lesson linked to module successfully" }
```

### 6. Crear Ejercicios (JSONB)

**Tipo quiz (multiple choice):**
```bash
curl -s -X POST http://localhost:3000/api/v1/exercises \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "lesson_id": 7,
    "type": "quiz",
    "order_index": 1,
    "content": {
      "question": "¿Cuál es el músculo principal para respirar al hablar?",
      "options": ["Diafragma", "Pecho", "Hombros", "Abdomen"],
      "correct_index": 0,
      "explanation": "El diafragma es el músculo clave para una respiración eficiente al hablar",
      "points": 10
    }
  }'
```

**Tipo reading (lectura):**
```bash
curl -s -X POST http://localhost:3000/api/v1/exercises \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "lesson_id": 7,
    "type": "reading",
    "order_index": 2,
    "content": {
      "title": "La importancia de la respiración",
      "content": "Texto completo del artículo...",
      "reading_time_seconds": 120,
      "points": 5
    }
  }'
```

**Tipo oratory_minigame (reto de oratoria):**
```bash
curl -s -X POST http://localhost:3000/api/v1/exercises \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "lesson_id": 7,
    "type": "oratory_minigame",
    "order_index": 3,
    "content": {
      "prompt": "Graba un discurso de 30s presentándote como líder",
      "topic": "Liderazgo",
      "duration_seconds": 30,
      "min_duration_seconds": 15,
      "requirements": [
        "Saludo inicial",
        "Menciona tu experiencia",
        "Cierra con una frase motivadora"
      ],
      "points": 20
    }
  }'
```

**Tipo audio (grabación de audio):**
```bash
curl -s -X POST http://localhost:3000/api/v1/exercises \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "lesson_id": 7,
    "type": "audio",
    "order_index": 4,
    "content": {
      "prompt": "Lee el siguiente párrafo en voz alta...",
      "duration_seconds": 60,
      "points": 15
    }
  }'
```

**Tipo video (grabación de video):**
```bash
curl -s -X POST http://localhost:3000/api/v1/exercises \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "lesson_id": 7,
    "type": "video",
    "order_index": 5,
    "content": {
      "prompt": "Grábate presentando un tema de tu elección por 2 minutos",
      "duration_seconds": 120,
      "points": 25
    }
  }'
```

### 7. Marcar Lección como Completada
```bash
curl -s -X POST http://localhost:3000/api/v1/progress \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "lesson_id": 7,
    "score": 85
  }'
# → { "success": true, "data": { "user_id": "uuid", "lesson_id": 7, "status": "completed", "xp_earned": 55, ... } }
```

### 8. Ver el Progreso del Usuario
```bash
curl -s http://localhost:3000/api/v1/progress \
  -H "Authorization: Bearer $TOKEN"
# → { "success": true, "data": [{ "user_id": "uuid", "lesson_id": 7, "status": "completed", ... }] }
```

## Consola MinIO

Para acceder a los archivos subidos:
1. Ir a http://localhost:9001
2. Login con las credenciales de arriba
3. Explorar el bucket `voxlab-media`
