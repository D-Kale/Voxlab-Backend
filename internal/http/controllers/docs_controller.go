package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var translations = map[string]string{
	// Info
	"Voxlab API": "Voxlab API — Documentación",
	"API for the Voxlab public speaking educational platform": "API de la plataforma educativa de oratoria Voxlab",
	"API Support": "Soporte API",

	// Tags
	"Auth":                         "Autenticación",
	"Tracks (Educational Content)": "Cursos",
	"Modules":                      "Módulos",
	"Lessons":                      "Lecciones",
	"Exercises":                    "Ejercicios",
	"Progress":                     "Progreso",
	"System":                       "Sistema",
	"Upload":                       "Subida de archivos",
	"Users":                        "Usuarios",

	// Auth endpoints
	"User Login":        "Iniciar sesión",
	"User Logout":       "Cerrar sesión",
	"Get Current User":  "Obtener usuario actual",
	"User Registration": "Registrar usuario",

	// Tracks
	"List all educational tracks": "Listar todos los cursos",
	"Get a single track by ID":    "Obtener curso por ID",
	"Create a new track":          "Crear curso nuevo",
	"Update an existing track":    "Actualizar curso",
	"Delete a track":              "Eliminar curso",

	// Modules
	"List modules for a track":  "Listar módulos del curso",
	"Get a single module by ID": "Obtener módulo por ID",
	"Create a new module":       "Crear módulo nuevo",
	"Update a module":           "Actualizar módulo",
	"Delete a module":           "Eliminar módulo",
	"Link a lesson to a module": "Vincular lección a módulo",

	// Lessons
	"List lessons in a module":  "Listar lecciones del módulo",
	"Get a single lesson by ID": "Obtener lección por ID",
	"Create a new lesson":       "Crear lección nueva",
	"Update a lesson":           "Actualizar lección",
	"Delete a lesson":           "Eliminar lección",

	// Exercises
	"List exercises for a lesson":                               "Listar ejercicios de la lección",
	"Get a single exercise by ID":                               "Obtener ejercicio por ID",
	"Create a new exercise":                                     "Crear ejercicio nuevo",
	"Update an exercise":                                        "Actualizar ejercicio",
	"Delete an exercise":                                        "Eliminar ejercicio",
	"Exercise UUID (e.g. 550e8400-e29b-41d4-a716-446655440000)": "UUID del ejercicio",

	// Progress
	"Get my learning progress": "Obtener mi progreso de aprendizaje",
	"Complete a lesson":        "Completar lección",

	// Upload
	"Upload track image":           "Subir imagen del curso",
	"Upload module image":          "Subir imagen del módulo",
	"Upload lesson image":          "Subir imagen de la lección",
	"Upload user avatar":           "Subir avatar del usuario",

	// Users
	"List all users (admin)":    "Listar usuarios (admin)",
	"Get user by ID (admin)":    "Obtener usuario por ID (admin)",
	"Update user (admin)":       "Actualizar usuario (admin)",
	"Delete user (admin)":       "Eliminar usuario (admin)",

	// Auth profile
	"Get user profile":    "Obtener perfil del usuario",
	"Update user profile": "Actualizar perfil del usuario",

	// Common responses
	"OK":           "OK",
	"Created":      "Creado",
	"Bad Request":  "Solicitud inválida",
	"Unauthorized": "No autorizado",
	"Not Found":    "No encontrado",
	"Conflict":     "Conflicto",
	"Server error": "Error del servidor",

	// Fields
	"Track ID (e.g. 1)":  "ID del curso",
	"Track data":         "Datos del curso",
	"Module ID (e.g. 1)": "ID del módulo",
	"Module data":        "Datos del módulo",
	"Lesson ID (e.g. 1)": "ID de la lección",
	"Lesson data":        "Datos de la lección",
	"Lesson ID to link":  "ID de la lección a vincular",
}

var prefixRules = []struct {
	prefix      string
	replacement string
}{
	{"Success: ", "Éxito: "},
	{"Created: ", "Creado: "},
	{"Updated: ", "Actualizado: "},
	{"Deleted: ", "Eliminado: "},
	{"Completed: ", "Completado: "},
	{"Linked: ", "Vinculado: "},
	{"Failed to ", "Error al "},
}

func translateValue(s string) string {
	if s == "" {
		return s
	}
	if translated, ok := translations[s]; ok {
		return translated
	}
	for _, rule := range prefixRules {
		if strings.HasPrefix(s, rule.prefix) {
			return rule.replacement + s[len(rule.prefix):]
		}
	}
	return s
}

func translateNode(node interface{}) interface{} {
	switch v := node.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			if str, ok := val.(string); ok {
				if key == "summary" || key == "description" || key == "title" {
					result[key] = translateValue(str)
				} else {
					result[key] = str
				}
			} else {
				result[key] = translateNode(val)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = translateNode(val)
		}
		return result
	default:
		return v
	}
}

type DocsController struct{}

func NewDocsController() *DocsController {
	return &DocsController{}
}

// ServeSwaggerUI godoc
// @Summary      Serve Spanish Swagger UI
// @Description  Serves an interactive Swagger UI page configured to load the
// @Description  Spanish-translated OpenAPI spec from /api/v1/docs/es/spec.
// @Description  This is the Spanish documentation portal for the Voxlab API.
// @Tags         Docs
// @Produce      html
// @Success      200  {string}  html  "HTML page with Swagger UI"
// @Router       /docs/es [get]
func (c *DocsController) ServeSwaggerUI(ctx *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Voxlab API - Documentación</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    body { margin: 0; padding: 0; background: #fff; }
    #swagger-ui { max-width: 1400px; margin: 0 auto; }
    .swagger-ui .topbar { background: #1e293b; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      SwaggerUIBundle({
        url: "/api/v1/docs/es/spec",
        dom_id: '#swagger-ui',
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
        layout: "BaseLayout",
        deepLinking: true
      });
    };
  </script>
</body>
</html>`
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Writer.Write([]byte(html))
}

// ServeTranslatedSpec godoc
// @Summary      Get Spanish OpenAPI specification
// @Description  Returns the full OpenAPI 2.0 specification translated to Spanish.
// @Description  Falls back to the English spec if the translation file is missing.
// @Description  This endpoint is consumed by the Swagger UI at /docs/es.
// @Tags         Docs
// @Produce      json
// @Success      200  {object}  object  "Translated OpenAPI specification"
// @Failure      500  {object}  map[string]interface{}  "Failed to read spec file"
// @Router       /api/v1/docs/es/spec [get]
func (c *DocsController) ServeTranslatedSpec(ctx *gin.Context) {
	specPaths := []string{
		"docs/es/openapi-es.json",
		"./docs/es/openapi-es.json",
		"/app/docs/es/openapi-es.json",
	}

	var data []byte
	var err error
	for _, path := range specPaths {
		data, err = os.ReadFile(path)
		if err == nil {
			ctx.Data(http.StatusOK, "application/json", data)
			return
		}
	}

	fallbackPaths := []string{
		"docs/swagger.json",
		"./docs/swagger.json",
		"/app/docs/swagger.json",
	}

	for _, path := range fallbackPaths {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(data, &spec); err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	translated := translateNode(spec)

	ctx.Header("Content-Type", "application/json")
	json.NewEncoder(ctx.Writer).Encode(translated)
}
