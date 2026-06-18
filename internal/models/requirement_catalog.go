package models

// RequirementCatalogItem represents a predefined requirement that
// teachers can select when creating writing exercises.
//
// The catalog is served via GET /api/v1/exercises/requirement-catalog
// so the frontend can render checkboxes instead of free-text inputs.
// Keeping it in Go ensures the frontend always gets the latest version.
type RequirementCatalogItem struct {
	ID       string `json:"id" example:"intro"`
	Text     string `json:"text" example:"Incluir una introducción clara del tema"`
	Category string `json:"category" example:"Cobertura del contenido"`
	Order    int    `json:"order" example:"1"`
}

// GetRequirementCatalog returns the canonical list of requirements.
// The frontend uses this to render selectable checkboxes grouped by category.
func GetRequirementCatalog() []RequirementCatalogItem {
	return []RequirementCatalogItem{
		// Cobertura del contenido
		{ID: "intro", Text: "Incluir una introducción clara del tema", Category: "Cobertura del contenido", Order: 1},
		{ID: "datos", Text: "Mencionar datos o estadísticas relevantes", Category: "Cobertura del contenido", Order: 2},
		{ID: "ejemplos", Text: "Incluir ejemplos concretos", Category: "Cobertura del contenido", Order: 3},
		{ID: "fuentes", Text: "Citar fuentes o referencias", Category: "Cobertura del contenido", Order: 4},
		{ID: "opinion", Text: "Expresar una opinión personal fundamentada", Category: "Cobertura del contenido", Order: 5},
		{ID: "experiencia", Text: "Describir una experiencia personal relacionada", Category: "Cobertura del contenido", Order: 6},
		{ID: "desafios", Text: "Identificar desafíos o problemas del tema", Category: "Cobertura del contenido", Order: 7},
		{ID: "soluciones", Text: "Proponer soluciones o recomendaciones", Category: "Cobertura del contenido", Order: 8},
		{ID: "beneficios", Text: "Mencionar beneficios o ventajas", Category: "Cobertura del contenido", Order: 9},

		// Estructura y organización
		{ID: "org_intro", Text: "Organizar el texto en introducción, desarrollo y conclusión", Category: "Estructura y organización", Order: 10},
		{ID: "conectores", Text: "Usar conectores textuales (sin embargo, además, por lo tanto)", Category: "Estructura y organización", Order: 11},
		{ID: "hilo", Text: "Mantener un hilo conductor claro entre párrafos", Category: "Estructura y organización", Order: 12},

		// Lenguaje y estilo
		{ID: "vocab_tecnico", Text: "Usar vocabulario técnico apropiado al tema", Category: "Lenguaje y estilo", Order: 13},
		{ID: "registro", Text: "Mantener un registro formal o académico", Category: "Lenguaje y estilo", Order: 14},
		{ID: "sin_repeticiones", Text: "Evitar repeticiones innecesarias", Category: "Lenguaje y estilo", Order: 15},
		{ID: "tono", Text: "Usar un tono persuasivo o argumentativo", Category: "Lenguaje y estilo", Order: 16},
	}
}
