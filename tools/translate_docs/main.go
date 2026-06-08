package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	baseDir := filepath.Join("..", "..")
	srcPath := filepath.Join(baseDir, "docs", "swagger.json")
	dstPath := filepath.Join(baseDir, "docs", "es", "openapi-es.json")

	data, err := os.ReadFile(srcPath)
	if err != nil {
		log.Fatalf("reading %s: %v", srcPath, err)
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(data, &spec); err != nil {
		log.Fatalf("parsing JSON: %v", err)
	}

	translateNode(spec, "")

	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		log.Fatalf("creating dir: %v", err)
	}

	out, err := json.MarshalIndent(spec, "", "    ")
	if err != nil {
		log.Fatalf("marshaling: %v", err)
	}

	if err := os.WriteFile(dstPath, out, 0644); err != nil {
		log.Fatalf("writing %s: %v", dstPath, err)
	}

	fmt.Printf("✓ Generated %s\n", dstPath)

	if len(untranslated) > 0 {
		fmt.Printf("\n⚠️  %d untranslated string(s). Add to translation map:\n", len(untranslated))
		for _, s := range sortedUntranslated() {
			fmt.Printf("  • %q\n", s)
		}
	} else {
		fmt.Println("✓ All strings translated")
	}
}

var untranslated = make(map[string]bool)

func sortedUntranslated() []string {
	var keys []string
	for s := range untranslated {
		keys = append(keys, s)
	}
	return keys
}

var skipKeys = map[string]bool{
	"name": true, "in": true, "required": true, "type": true,
	"format": true, "additionalProperties": true, "items": true,
}

func translateNode(node interface{}, path string) {
	switch v := node.(type) {
	case map[string]interface{}:
		for key, val := range v {
			if skipKeys[key] {
				continue
			}
			if str, ok := val.(string); ok {
				if translated := translateString(str, path+"."+key); translated != str {
					v[key] = translated
				}
			} else {
				translateNode(val, path+"."+key)
			}
		}
	case []interface{}:
		for i, val := range v {
			if str, ok := val.(string); ok {
				if translated := translateString(str, path); translated != str {
					v[i] = translated
				}
			} else {
				translateNode(val, fmt.Sprintf("%s[%d]", path, i))
			}
		}
	}
}

func translateString(s, _ string) string {
	if s == "" || s == "application/json" || strings.HasPrefix(s, "http") ||
		strings.HasPrefix(s, "#/") || strings.HasPrefix(s, "Bearer") ||
		strings.HasPrefix(s, "550e8400") || s == "string" || s == "integer" ||
		s == "boolean" || s == "number" || s == "object" || s == "array" {
		return s
	}

	if translated, ok := exactMatches[s]; ok {
		return translated
	}

	for _, rule := range prefixRules {
		if strings.HasPrefix(s, rule.prefix) {
			return rule.replacement + s[len(rule.prefix):]
		}
	}

	nonTranslatable := map[string]bool{
		"swagger":                             true,
		"2.0":                                 true,
		"1.0":                                 true,
		"http://voxlab.com/terms":             true,
		"MIT":                                 true,
		"https://opensource.org/licenses/MIT": true,
		"localhost:3000":                      true,
		"/api/v1":                             true,
	}
	if nonTranslatable[s] {
		return s
	}

	untranslated[s] = true
	return s
}

var exactMatches = map[string]string{
	// ── Info ──
	"Voxlab API": "Voxlab API — Documentación en Español",
	"API for the Voxlab public speaking educational platform": "API de la plataforma educativa de oratoria Voxlab. " +
		"Documentación en español para desarrolladores frontend, diseñadores y stakeholders.",
	"API Support": "Soporte API",

	// ── Tags ──
	"Auth":                         "Autenticación",
	"Tracks (Educational Content)": "Cursos (Tracks)",
	"Modules":                      "Módulos",
	"Lessons":                      "Lecciones",
	"Exercises":                    "Ejercicios",
	"Progress":                     "Progreso",
	"System":                       "Sistema",

	// ── Auth ──
	"User Login":        "Iniciar sesión",
	"User Logout":       "Cerrar sesión",
	"Get Current User":  "Obtener usuario actual",
	"User Registration": "Registrar usuario",
	"Login credentials": "Credenciales de inicio de sesión",
	"Registration data": "Datos de registro",
	"Health Check":      "Verificación de Salud",

	"Verifies API status, database and Redis connectivity": "Verifica el estado de la API, la conexión a la base de datos y Redis",

	"Type \"Bearer \" followed by your JWT token": "Escribe \"Bearer \" seguido de tu token JWT",

	"Authenticates with email + password and returns a JWT token.\nThe token must be sent as `Authorization: Bearer \u003ctoken\u003e` for protected endpoints.\nTokens expire after 24 hours.": "Autentica con email y contraseña y devuelve un token JWT.\n" +
		"El token debe enviarse como `Authorization: Bearer \u003ctoken\u003e` en los endpoints protegidos.\n" +
		"Los tokens expiran después de 24 horas.",

	"Creates a new user account and returns a JWT token (auto-login).\nThe password must be at least 6 characters.\nIf the email is already registered, returns a 409 conflict error.": "Crea una cuenta nueva y devuelve un token JWT (inicio de sesión automático).\n" +
		"La contraseña debe tener al menos 6 caracteres.\n" +
		"Si el email ya está registrado, devuelve un error 409 (conflicto).",

	"Invalidates the current JWT token by adding it to a Redis blacklist.\nAfter calling this, the token can no longer be used for authenticated requests.\nThe frontend should also discard the token locally.": "Invalida el token JWT actual agregándolo a una lista negra en Redis.\n" +
		"Después de esto, el token ya no puede usarse para peticiones autenticadas.\n" +
		"El frontend también debe descartar el token localmente.",

	"Returns the authenticated user's profile (name, email, XP, streak, lives).\nUse this to verify the token is valid and load the user's data on page refresh.": "Devuelve el perfil del usuario autenticado (nombre, email, XP, racha, vidas).\n" +
		"Úsalo para verificar que el token es válido y cargar los datos del usuario al recargar la página.",

	// ── Tracks ──
	"List all educational tracks": "Listar todos los cursos",
	"Get a single track by ID":    "Obtener un curso por ID",
	"Create a new track":          "Crear un nuevo curso",
	"Update an existing track":    "Actualizar un curso",
	"Delete a track":              "Eliminar un curso",
	"Track ID (e.g. 1)":           "ID del curso (ejemplo: 1)",
	"Track data":                  "Datos del curso",
	"Track fields to update":      "Campos del curso a actualizar",
	"Track not found":             "Curso no encontrado",

	// Track descriptions
	"Returns ALL available tracks with their modules, lessons, and exercises nested inside.\nThis is the main endpoint for the course catalog — frontends should call this once\nto build the full navigation tree. Each track contains modules, each module contains\nlessons (through the pivot table), and each lesson contains exercises.\n\n🔓 Public — no authentication required.": "Devuelve TODOS los cursos disponibles con sus módulos, lecciones y ejercicios anidados.\n" +
		"Este es el endpoint principal del catálogo — el frontend debe llamarlo una vez\n" +
		"para construir el árbol de navegación completo. Cada curso contiene módulos,\n" +
		"cada módulo contiene lecciones y cada lección contiene ejercicios.\n\n🔓 Público — no requiere autenticación.",

	"Returns one track with its nested modules, lessons, and exercises.\nUse this when you need to reload or fetch details for a specific track.\n\n🔓 Public — no authentication required.": "Devuelve un curso con sus módulos, lecciones y ejercicios anidados.\n" +
		"Úsalo cuando necesites recargar u obtener detalles de un curso específico.\n\n🔓 Público — no requiere autenticación.",

	"Adds a new educational track (course) to the catalog.\nAfter creating a track, you can add modules to it using POST /api/v1/modules.\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Agrega un nuevo curso (track) al catálogo educativo.\n" +
		"Después de crear un curso, puedes agregarle módulos usando POST /api/v1/modules.\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	"Modifies the title, description, or icon of a track.\nSend only the fields you want to update.\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Modifica el título, descripción o icono de un curso.\n" +
		"Envía solo los campos que deseas actualizar.\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	"Permanently removes a track and all its associated modules and module-lesson links.\n⚠️ This action cannot be undone.\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Elimina permanentemente un curso y todos sus módulos y vínculos asociados.\n" +
		"⚠️ Esta acción no se puede deshacer.\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	// ── Modules ──
	"List modules for a track":   "Listar módulos de un curso",
	"Get a single module by ID":  "Obtener un módulo por ID",
	"Create a new module":        "Crear un nuevo módulo",
	"Update a module":            "Actualizar un módulo",
	"Delete a module":            "Eliminar un módulo",
	"Link a lesson to a module":  "Vincular una lección a un módulo",
	"Module ID (e.g. 1)":         "ID del módulo (ejemplo: 1)",
	"Module data":                "Datos del módulo",
	"Module not found":           "Módulo no encontrado",
	"Module or Lesson not found": "Módulo o Lección no encontrados",

	"Returns all modules belonging to a specific track, ordered by order_index.\nEach module includes its linked lessons and their exercises.\n\n🔓 Public — no authentication required.": "Devuelve todos los módulos de un curso específico, ordenados por order_index.\n" +
		"Cada módulo incluye sus lecciones vinculadas y los ejercicios de cada lección.\n\n🔓 Público — no requiere autenticación.",

	"Returns one module with its linked lessons and exercises.\n\n🔓 Public — no authentication required.": "Devuelve un módulo con sus lecciones vinculadas y ejercicios.\n\n🔓 Público — no requiere autenticación.",

	"Adds a module inside a specific track (course).\nThe track_id must reference an existing track. Modules appear in order_index order.\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Agrega un módulo dentro de un curso específico.\n" +
		"El track_id debe referenciar un curso existente. Los módulos se ordenan por order_index.\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	"Changes the title, description, or order of a module.\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Cambia el título, descripción u orden de un módulo.\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	"Permanently removes a module and its lesson links. Lessons themselves are NOT deleted,\nonly the link between the module and the lesson is removed.\n⚠️ This action cannot be undone.\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Elimina permanentemente un módulo y sus vínculos a lecciones. Las lecciones NO se eliminan,\n" +
		"solo se elimina el vínculo entre el módulo y la lección.\n⚠️ Esta acción no se puede deshacer.\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	"Associates an existing lesson with a module using the ModuleLesson pivot table.\nA lesson can be linked to MULTIPLE modules. This does NOT move or copy the lesson.\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Asocia una lección existente con un módulo usando la tabla pivote ModuleLesson.\n" +
		"Una lección puede vincularse a MÚLTIPLES módulos. Esto NO mueve ni copia la lección.\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	// ── Lessons ──
	"List lessons in a module":  "Listar lecciones de un módulo",
	"Get a single lesson by ID": "Obtener una lección por ID",
	"Create a new lesson":       "Crear una nueva lección",
	"Update a lesson":           "Actualizar una lección",
	"Delete a lesson":           "Eliminar una lección",
	"Lesson ID (e.g. 1)":        "ID de la lección (ejemplo: 1)",
	"Lesson data":               "Datos de la lección",
	"Lesson ID to link":         "ID de la lección a vincular",
	"Lesson not found":          "Lección no encontrada",

	"Returns all lessons linked to a specific module, with their exercises.\nLessons are returned through the ModuleLesson pivot and include an order_index.\n\n🔓 Public — no authentication required.": "Devuelve todas las lecciones vinculadas a un módulo específico, con sus ejercicios.\n" +
		"Las lecciones se devuelven a través de la tabla pivote ModuleLesson e incluyen order_index.\n\n🔓 Público — no requiere autenticación.",

	"Returns one lesson with its exercises. Use this to load the full lesson content\nincluding all exercise data.\n\n🔓 Public — no authentication required.": "Devuelve una lección con sus ejercicios. Úsalo para cargar el contenido completo\n" +
		"de la lección incluyendo todos los ejercicios.\n\n🔓 Público — no requiere autenticación.",

	"Creates a standalone lesson. After creation, link it to a module using\nPOST /api/v1/modules/:id/lessons. Each lesson contains exercises (created separately).\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Crea una lección independiente. Después de crearla, vincúlala a un módulo usando\n" +
		"POST /api/v1/modules/:id/lessons. Cada lección tiene ejercicios (creados por separado).\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	"Changes the title, description, or estimated time of a lesson.\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Cambia el título, descripción o tiempo estimado de una lección.\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	"Permanently removes a lesson and its exercises.\n⚠️ This action cannot be undone.\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Elimina permanentemente una lección y sus ejercicios.\n⚠️ Esta acción no se puede deshacer.\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	// ── Exercises ──
	"List exercises for a lesson":                               "Listar ejercicios de una lección",
	"Get a single exercise by ID":                               "Obtener un ejercicio por ID",
	"Create a new exercise":                                     "Crear un nuevo ejercicio",
	"Update an exercise":                                        "Actualizar un ejercicio",
	"Delete an exercise":                                        "Eliminar un ejercicio",
	"Exercise UUID (e.g. 550e8400-e29b-41d4-a716-446655440000)": "UUID del ejercicio (ejemplo: 550e8400-e29b-41d4-a716-446655440000)",
	"Exercise data":                                             "Datos del ejercicio",
	"Exercise not found":                                        "Ejercicio no encontrado",

	"Returns all exercises in a lesson, ordered by order_index.\nEach exercise has a \"type\" field that defines the JSON structure of its \"content\" field.\nSee the \"content\" field descriptions below for each exercise type.\n\n🔓 Public — no authentication required.": "Devuelve todos los ejercicios de una lección, ordenados por order_index.\n" +
		"Cada ejercicio tiene un campo \"type\" que define la estructura JSON del campo \"content\".\n" +
		"Ver las descripciones abajo para cada tipo de ejercicio.\n\n🔓 Público — no requiere autenticación.",

	"Returns one exercise with its full JSONB content. The content structure depends\non the exercise type (quiz, reading, oratory_minigame, audio, video).\n\n🔓 Public — no authentication required.": "Devuelve un ejercicio con su contenido JSONB completo. La estructura del contenido\n" +
		"depende del tipo de ejercicio (quiz, reading, oratory_minigame, audio, video).\n\n🔓 Público — no requiere autenticación.",

	"Modifies the type, content (JSONB), or order of an exercise.\nWhen updating the content field, send the FULL new content object for the exercise type.\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Modifica el tipo, contenido (JSONB) u orden de un ejercicio.\n" +
		"Al actualizar content, envía el objeto NUEVO COMPLETO para el tipo de ejercicio.\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	"Permanently removes an exercise from a lesson.\n⚠️ This action cannot be undone.\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Elimina permanentemente un ejercicio de una lección.\n⚠️ Esta acción no se puede deshacer.\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	// ── Exercise creation with JSON examples ──
	"Creates an exercise inside a lesson. The \"type\" field determines the JSONB \"content\" structure.\n\n📝 Supported exercise types and their content structure:\n\n**quiz** — Multiple choice question:\n```json\n{\n\"type\": \"quiz\",\n\"content\": {\n\"question\": \"What is public speaking?\",\n\"options\": [\"Option A\", \"Option B\", \"Option C\", \"Option D\"],\n\"correct_index\": 0,\n\"explanation\": \"Option A is correct because...\",\n\"points\": 10\n}\n}\n```\n\n**reading** — Reading passage:\n```json\n{\n\"type\": \"reading\",\n\"content\": {\n\"title\": \"The Art of Speech\",\n\"content\": \"Full reading text here...\",\n\"reading_time_seconds\": 120,\n\"points\": 5\n}\n}\n```\n\n**oratory_minigame** — Oratory challenge with requirements:\n```json\n{\n\"type\": \"oratory_minigame\",\n\"content\": {\n\"prompt\": \"Record a 30-second speech about...\",\n\"topic\": \"Leadership\",\n\"duration_seconds\": 30,\n\"min_duration_seconds\": 15,\n\"requirements\": [\"Clear introduction\", \"Use at least 3 key points\", \"Strong conclusion\"],\n\"points\": 20\n}\n}\n```\n\n**audio** — Audio recording exercise:\n```json\n{\n\"type\": \"audio\",\n\"content\": {\n\"prompt\": \"Read this paragraph aloud...\",\n\"duration_seconds\": 60,\n\"points\": 15\n}\n}\n```\n\n**video** — Video recording exercise:\n```json\n{\n\"type\": \"video\",\n\"content\": {\n\"prompt\": \"Record a video introducing yourself...\",\n\"duration_seconds\": 120,\n\"points\": 25\n}\n}\n```\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Crea un ejercicio dentro de una lección. El campo \"type\" determina la estructura del campo \"content\" (JSONB).\n\n📝 Tipos de ejercicio soportados y su estructura:\n\n**quiz** — Pregunta de opción múltiple:\n```json\n{\n\"type\": \"quiz\",\n\"content\": {\n\"question\": \"¿Cuál es la función principal del diafragma al hablar?\",\n\"options\": [\"Proyectar la voz\", \"Respirar\", \"Articular\", \"Deglutir\"],\n\"correct_index\": 0,\n\"explanation\": \"El diafragma es el músculo clave para proyectar la voz\",\n\"points\": 10\n}\n}\n```\n\n**reading** — Lectura informativa:\n```json\n{\n\"type\": \"reading\",\n\"content\": {\n\"title\": \"La importancia de la pausa en el discurso\",\n\"content\": \"Texto completo de la lectura en español...\",\n\"reading_time_seconds\": 120,\n\"points\": 5\n}\n}\n```\n\n**oratory_minigame** — Reto de oratoria con requisitos:\n```json\n{\n\"type\": \"oratory_minigame\",\n\"content\": {\n\"prompt\": \"Graba un discurso de 30 segundos presentándote como líder\",\n\"topic\": \"Liderazgo\",\n\"duration_seconds\": 30,\n\"min_duration_seconds\": 15,\n\"requirements\": [\"Saludo inicial\", \"Menciona tu experiencia\", \"Cierre motivador\"],\n\"points\": 20\n}\n}\n```\n\n**audio** — Ejercicio de grabación de audio:\n```json\n{\n\"type\": \"audio\",\n\"content\": {\n\"prompt\": \"Lee este párrafo en voz alta...\",\n\"duration_seconds\": 60,\n\"points\": 15\n}\n}\n```\n\n**video** — Ejercicio de grabación de video:\n```json\n{\n\"type\": \"video\",\n\"content\": {\n\"prompt\": \"Grábate presentándote por 2 minutos...\",\n\"duration_seconds\": 120,\n\"points\": 25\n}\n}\n```\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	// ── Progress ──
	"Get my learning progress": "Obtener mi progreso",
	"Complete a lesson":        "Completar una lección",
	"Lesson completion data":   "Datos de finalización de lección",

	"Returns ALL progress records for the authenticated user (every lesson they've started or completed).\nEach record shows: status (in_progress/completed), xp_earned, and timestamps.\nUse this on the frontend to determine which lessons are completed and show user progress bars.\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Devuelve TODOS los registros de progreso del usuario autenticado.\n" +
		"Cada registro muestra: estado (in_progress/completed), XP ganado y timestamps.\n" +
		"Úsalo en el frontend para determinar qué lecciones están completadas.\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	"Marks a lesson as completed for the authenticated user. This endpoint:\n1. Sets the progress status to \"completed\"\n2. Adds XP to the user's total (calculated from exercises + score)\n3. Stores the completion timestamp\n\nIf the lesson was already completed before, it UPDATES the existing record\n(the final score replaces the previous one).\n\n🔒 Requires JWT token (Authorization: Bearer \u003ctoken\u003e)": "Marca una lección como completada para el usuario autenticado. Este endpoint:\n" +
		"1. Cambia el estado del progreso a \"completed\"\n" +
		"2. Agrega XP al total del usuario (según ejercicios + puntuación)\n" +
		"3. Guarda la fecha de finalización\n\n" +
		"Si la lección ya estaba completada, ACTUALIZA el registro existente.\n\n🔒 Requiere token JWT (Authorization: Bearer \u003ctoken\u003e).",

	// ── Common response descriptions ──
	"OK":               "OK",
	"Created":          "Creado",
	"Bad Request":      "Solicitud inválida",
	"Unauthorized":     "No autorizado",
	"Not Found":        "No encontrado",
	"Conflict":         "Conflicto",
	"Server error":     "Error del servidor",
	"Validation error": "Error de validación",

	// ── Fields to update (shared) ──
	"Fields to update": "Campos a actualizar",

	// ── Security ──
	"Unauthorized — token missing or invalid": "No autorizado — token faltante o inválido",

	// ── Definition field descriptions ──
	"support@voxlab.com":     "soporte@voxlab.com",
	"User email":             "Email del usuario",
	"User password":          "Contraseña del usuario",
	"User name":              "Nombre del usuario",
	"Lesson ID":              "ID de la lección",
	"Score obtained (0-100)": "Puntuación obtenida (0-100)",

	// ── Example values in schemas ──
	"user@example.com": "usuario@ejemplo.com",
	"password123":      "contraseña123",
	"John Doe":         "Juan Pérez",
}

var prefixRules = []prefixRule{
	{"Success: {", "Éxito: {"},
	{"Created: {", "Creado: {"},
	{"Updated: {", "Actualizado: {"},
	{"Deleted: {", "Eliminado: {"},
	{"Completed: {", "Completado: {"},
	{"Linked: {", "Vinculado: {"},
}

type prefixRule struct {
	prefix      string
	replacement string
}
