# translate-docs

Traduce `docs/swagger.json` (inglés) a `docs/es/openapi-es.json` (español latinoamericano) usando la API de NVIDIA.

## Requisitos

- Python >= 3.13
- [uv](https://docs.astral.sh/uv/) (gestor de paquetes)
- Clave de API de NVIDIA (gratuita en [build.nvidia.com](https://build.nvidia.com/))

## Instalación

```bash
# uv ya instalado — las dependencias se instalan automáticamente al ejecutar
cd tools/translate-docs
uv sync
```

## Configuración

La clave de API se lee desde (en orden de prioridad):

1. `tools/translate-docs/.env`
2. `.env` (raíz del proyecto)
3. `.env.example` (raíz del proyecto)

Crea `tools/translate-docs/.env`:

```env
NVIDIA_API_KEY=tu_clave_aqui
```

## Uso

```bash
# Traducción completa
uv run python main.py --verbose

# Sin verbose (solo output relevante)
uv run python main.py

# Mostrar diff git después de traducir
uv run python main.py --show-diff

# Re-traducir todo (ignorando traducciones existentes)
uv run python main.py --force

# Solo validar archivo existente
uv run python main.py --validate-only

# Timeout personalizado (default: 300s)
uv run python main.py --timeout 600
```

## Arquitectura

```
tools/translate-docs/
├── main.py          # Entry point, argparse, orquestación (~70 líneas)
├── swagger.py       # Recorrido del árbol swagger, extraer/aplicar/validar (~130 líneas)
├── translate.py     # Prompt + llamada a NVIDIA API (~60 líneas)
├── util.py          # Constantes, load_env, merge, show_diff (~50 líneas)
├── pyproject.toml   # Configuración uv
└── README.md
```

## Flujo

1. Lee `docs/swagger.json` y extrae strings traducibles (summary, description, title, tags, etc.)
2. Si existe `docs/es/openapi-es.json`, preserva traducciones existentes (modo incremental)
3. Envía strings nuevos o modificados a `moonshotai/kimi-k2.6` vía NVIDIA API
4. Aplica las traducciones a una copia profunda del spec original
5. Escribe a `docs/es/tmp_swagger_es.json` → valida Swagger 2.0 → renombra a `docs/es/openapi-es.json`

## Modelo

Usa `moonshotai/kimi-k2.6` (tier gratuito de NVIDIA). ~2-3 min para 178 strings.

## Docker

El archivo `docs/es/openapi-es.json` se copia en la imagen Docker durante el build.
El endpoint `GET /api/v1/docs/es/spec` lo sirve directamente.
`GET /docs/es` muestra Swagger UI en español cargando desde ese endpoint.
