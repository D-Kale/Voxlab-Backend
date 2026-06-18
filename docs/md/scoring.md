# Scoring — Puntuación de ejercicios de escritura

## 1. Para alumnos — ¿Cómo se calcula tu nota?

Cuando escribís un texto y lo enviás, el sistema lo analiza automáticamente y te da una nota de **0 a 100**. Esa nota no es un número al azar: mide **4 aspectos** de tu escritura, cada uno con su propio peso.

| Aspecto | ¿Qué mide? | Peso en la nota |
|---|---|---|
| **Contenido** | ¿Cubriste los requisitos del ejercicio? Si te pidieron "dar ejemplos", el sistema detecta si lo hiciste. | 30% |
| **Organización** | ¿Separaste en párrafos? ¿Usaste conectores como "además" o "por otro lado"? ¿Hay introducción y conclusión? | 25% |
| **Lenguaje** | ¿Usaste vocabulario variado? ¿Evitaste muletillas como "este" o "eh"? ¿El texto se entiende bien? | 25% |
| **Extensión** | ¿Respetaste el mínimo y máximo de palabras? | 20% |

Cada aspecto suma puntos. La nota final es el promedio ponderado: si te fue muy bien en contenido (30% de la nota) pero regular en organización (25%), tu nota refleja eso.

> **Ejemplo:** Si sacás 90 en contenido, 60 en organización, 70 en lenguaje y 100 en extensión:
> `(90 × 0.30) + (60 × 0.25) + (70 × 0.25) + (100 × 0.20) = **79**`

El sistema también detecta si escribiste cualquier cosa sin sentido (texto aleatorio, letras al azar). En ese caso la nota es **0** directamente.

---

## 2. Para docentes — ¿Cómo funciona cada análisis?

### 2.1 Detección de texto sin sentido (gibberish)

Antes de evaluar, el sistema verifica que el texto sea español real usando **3 filtros**:

1. **Letras y vocales** — el español tiene ~45% de vocales. Si un texto tiene menos de 20% o más de 70%, algo raro pasa. Ejemplo: `"khyucubjgyufuuj"` tiene 15% vocales → descartado.
2. **Combinaciones de letras** — compara las parejas de letras del texto contra las combinaciones más comunes del español. Si no se parecen, es sospechoso.
3. **Palabras conocidas** — usa el diccionario del sistema para ver si reconoce las palabras. Si más de la mitad son desconocidas, no es texto válido.

Si cualquiera de estos filtros se activa, el texto recibe nota 0 y no se analiza nada más.

### 2.2 Revisión de contenido (requisitos)

El sistema lee el texto completo y cada requisito, los convierte a "vectores matemáticos" (una representación numérica de su significado), y mide qué tan parecidos son.

Por ejemplo, si el requisito es *"Incluir ejemplos concretos"*, el sistema busca en el texto párrafos que hablen de ejemplos, aunque el alumno no use exactamente la palabra "ejemplo". Si la similitud es alta (>45%), el requisito se marca como cumplido.

### 2.3 Revisión de estructura

- **Párrafos:** cuenta cuántos párrafos tiene el texto. Más de 2 sugiere que hay introducción; más de 3 sugiere introducción, desarrollo y conclusión.
- **Variedad de oraciones:** mide si las oraciones tienen longitudes variadas (todas muy cortas o muy largas es monótono).
- **Conectores:** busca palabras como "además", "sin embargo", "por lo tanto", "en conclusión". Si más del 30% de las oraciones tienen al menos un conector, se considera buen uso.

### 2.4 Revisión de calidad del lenguaje

- **Riqueza léxica:** compara cuántas palabras *distintas* usó el alumno contra el total de palabras. Si usa muchas palabras repetidas, resta puntos.
- **Legibilidad:** usa la fórmula Fernández-Huerta (adaptación al español del método Flesch). Mide qué tan fácil o difícil es leer el texto. Textos muy simples o muy complejos puntúan menos.
- **Muletillas:** cuenta las palabras de relleno como "este...", "eh", "o sea". Cada una resta.
- **Palabras desconocidas:** si hay muchas palabras que el diccionario del sistema no reconoce, puede indicar problemas de vocabulario.

### 2.5 Revisión de extensión

Simplemente verifica que el texto esté dentro del rango de palabras que configuraste en el ejercicio. Si está abajo del mínimo o arriba del máximo, la nota se reduce proporcionalmente.

---

## 3. Respuesta de la API (lo que ve el frontend)

```json
{
  "score": 79,
  "gibberish_detected": false,
  "score_breakdown": {
    "cobertura_requisitos": { "score": 90, "weight": 0.30, "feedbacks": [...] },
    "estructura":           { "score": 60, "weight": 0.25, "feedbacks": [...] },
    "calidad_linguistica":  { "score": 70, "weight": 0.25, "feedbacks": [...] },
    "longitud":             { "score": 100, "weight": 0.20, "feedbacks": [...] }
  },
  "feedback": ["Requisito cumplido: \"Incluir ejemplos\"", "Buena estructura", "..."],
  "word_count": 245,
  "requirements": [...],
  ...
}
```

El `score_breakdown` se muestra en el frontend como barras de progreso individuales, para que el alumno vea exactamente en qué área le fue bien o mal.

---

---

# A. Pipeline técnico completo

```
texto_alumno
    │
    ├─ 1. Gibberish Detection (3 filtros secuenciales, más rápido primero)
    │   ├─ vowel_consonant_ratio: <20% o >70% → gibberish
    │   ├─ bigram_frequency: <5% bigramas españoles → gibberish
    │   └─ oov_detection: >50% tokens OOV (spaCy) → gibberish
    │
    │   Si es gibberish → score = 0, detener
    │
    └─ 2. Análisis completo
        ├─ word_count, sentence_count (spaCy)
        ├─ sentence_length: avg, min, max, std
        ├─ sentence_analysis: count, avg_length, std_length, connector_ratio
        ├─ paragraph_structure: count, has_introduction, has_conclusion
        ├─ vocabulary: lexical_richness (TTR), filler_words, keywords, oov_ratio
        ├─ readability: fernandez_huerta, flesch, label
        ├─ requirements: embeddings + cosine similarity vs cada requisito
        └─ scoring: fórmula ponderada con SCORE_FORMULA
```

# B. Fórmulas y pesos exactos

## B.1 SCORE_FORMULA

```python
SCORE_FORMULA = {
    "cobertura_requisitos": {"peso": 0.30},
    "estructura":           {"peso": 0.25},
    "calidad_linguistica":  {"peso": 0.25},
    "longitud":             {"peso": 0.20},
}
```

score_final = Σ (score_componente × peso_componente) para todos los componentes.

## B.2 Cobertura requisitos (30%)

Cada requisito se evalúa individualmente:

```python
text_embedding = model.encode(texto_alumno)
req_embedding = model.encode(cada_requisito)
similitud = cosine_similarity(text_embedding, req_embedding)
# similitud >= 0.45 → matched = True
# score_componente = promedio(similitudes) × 100
```

Modelo: `paraphrase-multilingual-MiniLM-L12-v2` (~130MB disco, ~80MB RAM int8).

## B.3 Estructura (25%)

| Sub-componente | Puntaje | Criterio |
|---|---|---|
| Párrafos | 0–40 | 1 párrafo = 10, 2 = 25, 3+ = 40 |
| Conclusión | 0–20 | presente si >= 3 párrafos |
| Variedad oraciones (std) | 0–20 | std > 4 = 20, std > 2 = 10 |
| Conectores textuales | 0–20 | ratio ≥ 30% = 20, ≥ 15% = 10 |

## B.4 Calidad lingüística (25%)

| Sub-componente | Puntaje | Criterio |
|---|---|---|
| TTR (type-token ratio) | 0–30 | ≥ 0.60 = 30, ≥ 0.45 = 20, ≥ 0.30 = 10 |
| Legibilidad (Fernández-Huerta) | 0–30 | 50–80 pts = 30, 30–50 o 80–90 = 15 |
| Muletillas | 0–20 | 0 = 20, 1–2 = 10, 3+ = 0 |
| OOV (palabras fuera de vocabulario) | 0–20 | < 10% = 20, < 25% = 10 |

## B.5 Longitud (20%)

- Sin min ni max definidos: score = 100
- Dentro del rango: score = 100
- Por debajo del mínimo: `score = (word_count / min_words) × 100`
- Por encima del máximo: `score = max(0, (1 - (word_count - max_words) / max_words) × 100)`

# C. Detección de gibberish — detalle algorítmico

## C.1 Vowel/consonant ratio

```python
# El español tiene ~45% de vocales en texto natural
letters = re.findall(r"[a-záéíóúüñ]", text, re.I)
vowels = sum(1 for ch in letters if ch in "aeiouáéíóúü")
ratio = vowels / len(letters)
if ratio < 0.20 or ratio > 0.70: gibberish = True
```

## C.2 Bigram frequency

```python
# Top-100 bigramas del español (el, es, en, la, de, qu, ue, ...)
TOP_BIGRAMS = {"el", "es", "qu", "ue", "la", "ad", ...}
bigrams = set(text[i] + text[i+1] for i in range(len(text)-1))
coverage = sum(1 for bg in bigrams if bg in TOP_BIGRAMS) / len(bigrams)
if coverage < 0.05: gibberish = True
```

## C.3 OOV detection (spaCy)

```python
doc = nlp(text)
tokens = [t for t in doc if t.is_alpha]
oov = sum(1 for t in tokens if not t.has_vector and t.is_oov)
if oov / len(tokens) > 0.50: gibberish = True
```

# D. Stack tecnológico

| Componente | Tecnología | Propósito |
|---|---|---|
| API | Python + FastAPI | Endpoint /analyze/text |
| NLP básico | spaCy + es_core_news_md | Tokenización, lematización, POS, OOV |
| Embeddings | sentence-transformers (MiniLM) | Matching semántico de requisitos |
| Legibilidad | textstat | Fernández-Huerta, Flesch |
| Scoring | Python puro (sin librerías externas) | Fórmula ponderada |
| Proxy API | Go (backend principal) | Router, auth, orquestación |

# E. Ejemplo paso a paso

### Ejercicio
- **Consigna:** "Escribí un texto sobre liderazgo"
- **Requisitos:** `["Incluir una introducción", "Dar ejemplos concretos"]`
- **Límites:** mínimo 100 palabras, máximo 500

### Texto del alumno (180 palabras)
```
El liderazgo es una habilidad fundamental en el mundo actual.
...
```

### Pipeline

1. **Gibberish check:** pasa los 3 filtros (ratio vocales 44%, bigramas 62%, OOV 8%) → texto válido
2. **Embeddings:** texto → vector 384d, cada requisito → vector 384d
   - "Incluir una introducción" vs texto → 0.82 similitud → **cumplido**
   - "Dar ejemplos concretos" vs texto → 0.55 similitud → **cumplido**
   - Cobertura: promedio (0.82 + 0.55) / 2 = 0.685 → score **69**
3. **Estructura:** 4 párrafos (40 pts) + std 5.2 (20 pts) + conectores 35% (20 pts) = **80**
4. **Calidad lingüística:** TTR 0.55 (20 pts) + legibilidad 65 (30 pts) + 0 muletillas (20 pts) + OOV 6% (20 pts) = **90**
5. **Longitud:** 180 palabras dentro de [100, 500] = **100**

### Cálculo final

| Componente | Score | Peso | Aporte |
|---|---|---|---|
| cobertura_requisitos | 69 | × 0.30 | = 20.7 |
| estructura | 80 | × 0.25 | = 20.0 |
| calidad_linguistica | 90 | × 0.25 | = 22.5 |
| longitud | 100 | × 0.20 | = 20.0 |
| **Total** | | | **83** |

# F. Guía para desarrolladores — Cómo agregar un componente

### Paso 1: Crear la función de scoring

En `analyzers/scoring.py`:

```python
def score_nuevo_componente(metrics: dict) -> Tuple[int, List[str]]:
    # calcular score 0-100 basado en metrics
    # generar feedbacks específicos
    return score, feedbacks
```

### Paso 2: Registrar en SCORE_FORMULA

```python
SCORE_FORMULA["nuevo_componente"] = {
    "peso": 0.10,
    "descripcion": "Qué mide este componente",
}
# Ajustar los pesos existentes para que sumen 1.0
```

### Paso 3: Agregar case en calculate_score()

```python
elif key == "nuevo_componente":
    s, fb = score_nuevo_componente(metrics)
```

### Paso 4: Agregar métricas necesarias

Si el componente necesita datos nuevos, agregarlos en `analyzers/text.py` → `analyze_text()`.

### Paso 5: Propagar el campo nuevo

- Agregar en `main.py` → `AnalyzeResponse` (si el dato debe exponerse)
- Agregar en `backend/internal/analyzer/client.go` → `AnalyzeResponse`
