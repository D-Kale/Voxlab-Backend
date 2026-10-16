# Crear contenido de Voxlab desde un agente

Guía completa para que un agente (por ejemplo, otro chat de Claude) cree y
edite el contenido del curso: tracks, módulos, lecciones y ejercicios. Todo
pasa por la API REST; no hace falta tocar la base de datos.

- API: `https://tvoxlab.tech/api/v1` (Swagger en `/api/v1/swagger/index.html`)
- Cliente listo para usar: [`tools/content/voxlab_content.py`](../../tools/content/voxlab_content.py)
- Plan de ejemplo: [`tools/content/example_plan.json`](../../tools/content/example_plan.json)

## 1. Acceso

Las escrituras de contenido exigen una cuenta con `role = "admin"`.

- El backend crea un admin al arrancar si no hay ninguno, con
  `ADMIN_EMAIL` / `ADMIN_PASSWORD` del `.env` del VPS.
- Para el agente conviene una cuenta propia: registrala en la app o con
  `POST /auth/register`, y que un admin la promueva con
  `PUT /users/{id}` y `{"role": "admin"}`.
- Login: `POST /auth/login` con `{"email", "password"}`. Responde
  `{"token", "expires_at", "user"}` **sin** el sobre `data`. Mandá el token
  como `Authorization: Bearer <token>`.

Nunca pegues la contraseña en el chat del agente: pasala por variables de
entorno (`VOXLAB_EMAIL`, `VOXLAB_PASSWORD`) en la máquina donde corre.

## 2. Modelo de contenido

```
Track (id int, order_index, is_active)
 └─ Module (id int, track_id, order_index)          = "Unidad" en el mapa de la app
     └─ ModuleLesson (module_id, lesson_id, order_index)
         └─ Lesson (id int)                          una lección puede estar en varios módulos
             └─ LessonExercise (lesson_id, exercise_id, order_index)
                 └─ Exercise (id uuid, name, type, content JSON)
```

- Lecciones y ejercicios son independientes: se crean solos y después se
  **vinculan**. Un mismo ejercicio puede reutilizarse en varias lecciones.
- El orden lo da `order_index` en cada nivel. `GET /tracks` devuelve todo
  ordenado.
- `GET /tracks` (público) devuelve el árbol completo:
  `modules[].lessons[].lesson.exercises[].exercise`. Es lo mismo que lee la
  app.

## 3. Cómo usa la app el contenido

- **Camino lineal:** la app recorre tracks, módulos, lecciones y ejercicios
  en orden. Desbloquea los completados más el siguiente. El orden importa:
  va de lo simple a lo complejo.
- **Borradores:** un track con `is_active = false` no aparece en la app.
  Armá todo oculto y publicá al final.
- **Tipos soportados:** `quiz`, `reading`, `writing` y `oratory_minigame`.
  La API también acepta `audio` y `video`, pero la app los muestra como "no
  disponible": **no los uses**.
- **XP:** la app muestra `+N XP` antes de empezar.
  - En `quiz`, N es `preguntas × points_per_question`.
  - En los demás tipos, N es `points`.
  - Referencia: lectura 5, quiz 10-15, escritura 15, oratoria 20.
- **Idioma y tono:** español rioplatense con voseo, igual que toda la app
  ("Hablá", "Escribí", "Contá"). Frases cortas y amables, tono Duolingo.
- **IA:**
  - Escritura se corrige con NLP. Cada `requirements` se compara por
    similitud semántica con el texto, y un LLM revisa los casos dudosos.
  - Oratoria se transcribe con Whisper, que mide ritmo, pausas y
    muletillas. Un LLM opina sobre el contenido usando `topic`.

## 4. Contenido de cada tipo (`content`)

La API no valida `content`. Si un campo falta, la app usa `0` o `""` y el
ejercicio se rompe en silencio. Validá siempre con
`voxlab_content.py validate` antes de crear.

### quiz
```json
{
  "points_per_question": 5,
  "questions": [
    {
      "question": "¿Qué ayuda a calmar los nervios antes de hablar?",
      "options": ["Respirar lento", "Hablar más rápido", "Evitar mirar al público"],
      "correct_index": 0,
      "explanation": "Respirar lento baja el ritmo cardíaco."
    }
  ]
}
```
- Reglas: 2 a 4 opciones por pregunta. `correct_index` empieza en 0.
- `explanation` es opcional; la app la muestra al responder.
- Recomendado: 2 a 5 preguntas.

### reading
```json
{
  "title": "La regla de los tres",
  "content": "Primer párrafo...\n\nSegundo párrafo...",
  "reading_time_seconds": 60,
  "points": 5
}
```
- Separá los párrafos con `\n\n`.
- Usá 80 a 250 palabras. Calculá `reading_time_seconds` a unas 200
  palabras por minuto.

### writing
```json
{
  "prompt": "Escribí cómo te presentarías ante un grupo nuevo.",
  "min_words": 20,
  "max_words": 120,
  "requirements": ["Tu nombre", "A qué te dedicás", "Un dato curioso"],
  "points": 15
}
```
- `requirements`: entre 1 y 4. Escribí cada uno como un tema concreto
  ("Contar a qué te dedicás"), no como una regla de forma ("usar 3
  párrafos"). El matching es semántico.
- Hay un catálogo de requisitos típicos en
  `GET /exercises/requirement-catalog`.
- `min_words` < `max_words`. Para el celular, 20-150 palabras es lo razonable.

### oratory_minigame
```json
{
  "prompt": "Hablá sobre el tema sin detenerte.",
  "topic": "Presentación breve",
  "duration_seconds": 60,
  "min_duration_seconds": 20,
  "requirements": ["Volumen claro", "Sin muletillas"],
  "points": 20
}
```
- `duration_seconds` es el tiempo máximo, con 180 como tope. La toma se
  corta sola al llegar.
- `min_duration_seconds` es el mínimo para que cuente bien.
- `topic` va al LLM coach: hacelo específico ("Tu comida favorita y por
  qué"), no genérico.
- `requirements` acá son objetivos que se muestran al usuario. Los que mide
  la IA son ritmo, pausas, muletillas, fluidez y duración.

## 5. Endpoints

Lectura pública (sin token):

| Método | Ruta | Devuelve |
|---|---|---|
| GET | `/tracks` | Árbol completo ordenado |
| GET | `/tracks/{id}` | Un track con su árbol |
| GET | `/tracks/{id}/modules` | Módulos del track |
| GET | `/modules/{id}/lessons` | Lecciones del módulo (pivot con `order_index`) |
| GET | `/lessons` | Todas las lecciones |
| GET | `/lessons/{id}/exercises` | Ejercicios de la lección |
| GET | `/exercises` | Todos los ejercicios |
| GET | `/exercises/{id}` | Un ejercicio |
| GET | `/exercises/requirement-catalog` | Requisitos sugeridos para escritura |

Escritura (token de **admin**):

| Método | Ruta | Body |
|---|---|---|
| POST | `/tracks` | `{title, description, icon_url?}` (nace activo) |
| PUT | `/tracks/{id}` | `{title?, description?, icon_url?, is_active?}` |
| PUT | `/tracks/reorder` | `{items: [{id, order_index}]}` |
| DELETE | `/tracks/{id}` | — |
| POST | `/modules` | `{track_id, title, description, order_index}` |
| PUT | `/modules/{id}` | `{title?, description?, order_index?}` |
| PUT | `/tracks/{id}/modules/reorder` | `{items: [{id, order_index}]}` |
| DELETE | `/modules/{id}` | — |
| POST | `/lessons` | `{title, description, estimated_time_seconds}` |
| PUT | `/lessons/{id}` | `{title?, description?, estimated_time_seconds?}` |
| DELETE | `/lessons/{id}` | — |
| POST | `/modules/{id}/lessons` | `{lesson_id}` (queda con `order_index` 0: **reordená después**) |
| PUT | `/modules/{id}/lessons/reorder` | `{items: [{lesson_id, order_index}]}` |
| DELETE | `/modules/{id}/lessons/{lessonId}` | desvincula |
| POST | `/exercises` | `{name, type, content}` |
| PUT | `/exercises/{id}` | `{name?, type?, content?}` (`content` se reemplaza **entero**) |
| DELETE | `/exercises/{id}` | — |
| POST | `/lessons/{id}/exercises` | `{exercise_id}` (se agrega al final) |
| PUT | `/lessons/{id}/exercises/reorder` | `{items: [{exercise_id, order_index}]}` |
| DELETE | `/lessons/{id}/exercises/{exerciseId}` | desvincula |
| POST | `/upload/{track\|module\|lesson}/{id}` | multipart, campo `file` (imagen) |

Respuestas: `{"success": true, "data": ...}`. Errores: `{"error": "..."}` con
400 (body inválido), 401 (sin token o vencido), 403 (no es admin) o 404.

## 6. Flujo recomendado

1. **Leer** el estado actual: `python3 voxlab_content.py tree`. No dupliques
   títulos ni nombres; el cliente los usa para no crear dos veces.
2. **Planificar** un JSON con el formato de `example_plan.json`:
   un track (nuevo por `title` o existente por `id`), sus módulos,
   lecciones y ejercicios, en el orden en que se cursan.
3. **Validar**: `python3 voxlab_content.py validate plan.json`.
4. **Probar** sin escribir: `python3 voxlab_content.py apply plan.json --dry-run`.
5. **Aplicar**: `python3 voxlab_content.py apply plan.json`.
   - Un track nuevo queda como borrador.
   - Si lo corrés otra vez, solo crea lo que falta y vuelve a ordenar.
6. **Revisar** con `tree` y, si podés, en la app con un track de prueba.
7. **Publicar**: `python3 voxlab_content.py publish <track_id>`.

Para agregar contenido a un track **ya publicado**: usá `"track": {"id": N}`
en el plan. Los módulos nuevos aparecen apenas se vinculan sus lecciones.

## 7. Reglas para el agente

- Nunca uses `DELETE` sin confirmación explícita de una persona. Borrar un
  ejercicio afecta a todas las lecciones que lo usan y al progreso de los
  usuarios.
- No edites ejercicios publicados para cambiar su sentido. Creá uno nuevo y
  reemplazá el vínculo.
- Todo contenido nuevo va en borrador hasta que una persona lo revise.
- Validá cada plan antes de aplicarlo y mostrá el resumen (`tree`) al
  terminar.
- No inventes datos ni estadísticas en las lecturas. Si citás un dato,
  ponelo como orientativo.
- Contenido apto para todo público, inclusivo y sin marcas ni personas
  reales.

## 8. Guía pedagógica

- Cada **módulo** es una habilidad (por ejemplo, "Voz y respiración",
  "Estructura del mensaje", "Manejo de nervios") con 2 a 4 lecciones.
- Cada **lección** tiene 3 a 5 ejercicios en este ciclo:
  1. `reading`: el concepto, breve.
  2. `quiz`: comprensión.
  3. `writing`: aplicarlo por escrito.
  4. `oratory_minigame`: decirlo en voz alta.
- Dificultad creciente:
  - Al principio, oratorias de 30-45 s con mínimo de 10-15 s.
  - Después, 60-120 s con mínimo de 30-60 s.
- Un mismo hilo temático por lección; el `topic` de la oratoria retoma lo
  que el usuario escribió.

## 9. Prompt para otro chat de Claude

Pegá esto en el chat que va a crear contenido, junto con este archivo:

> Sos el autor de contenido de Voxlab, una app tipo Duolingo para aprender
> oratoria en español rioplatense. Leé `CONTENT_AGENT.md`. Para crear
> contenido, escribí un plan JSON con el formato de `example_plan.json`. Usá
> solo los tipos `reading`, `quiz`, `writing` y `oratory_minigame`. Validalo
> con `python3 tools/content/voxlab_content.py validate plan.json`, probalo
> con `apply --dry-run` y aplicalo. Dejá el track en borrador y mostrame el
> resultado de `tree` para que lo revise antes de publicar. Nunca borres
> nada sin que te lo confirme.

## 10. Problemas comunes

| Síntoma | Causa |
|---|---|
| 403 `admin access required` | La cuenta no es admin |
| 401 `invalid or expired token` | Token vencido: volvé a hacer login |
| Lecciones en otro orden en la app | Faltó `PUT /modules/{id}/lessons/reorder` (el vínculo nace con orden 0) |
| Ejercicio vacío o "no disponible" en la app | `content` con campos faltantes, o tipo `audio`/`video` |
| Track no aparece en la app | Está en borrador (`is_active = false`) |
