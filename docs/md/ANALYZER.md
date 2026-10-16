# Voxlab Text Analyzer Service

## Overview

The Text Analyzer is a Python microservice that performs Natural Language Processing (NLP) on writing exercise submissions. It provides detailed feedback and scoring for Spanish-language texts.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Frontend Request                             │
│  POST /api/v1/exercises/analyze-text                            │
│  Authorization: Bearer <token>                                  │
│  Body: { text, requirements, min_words, max_words }            │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Go Backend                                   │
│  ExerciseController.AnalyzeText()                               │
│  → analyzer.AnalyzeText()                                       │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Analyzer Client                              │
│  HTTP POST http://analyzer:8000/analyze/text                    │
│  Retry logic: 3 attempts, 180s timeout                        │
│  Fallback: localhost:8001                                       │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    FastAPI Server                               │
│  POST /analyze/text                                             │
│  Request: AnalyzeRequest                                          │
│  Response: AnalyzeResponse                                        │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    spaCy Pipeline (es_core_news_lg)               │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │  1. Tokenization                                          │ │
│  │  2. Sentence Segmentation                                 │ │
│  │  3. POS Tagging                                           │ │
│  │  4. Lemmatization                                         │ │
│  │  5. Named Entity Recognition (NER)                        │ │
│  │  6. Dependency Parsing                                    │ │
│  └───────────────────────────────────────────────────────────┘ │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Analysis Components                          │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │   textstat   │  │  Keyword     │  │   Requirement        │ │
│  │  (Readability│  │  Extractor   │  │   Matcher            │ │
│  │  metrics)    │  │              │  │                      │ │
│  └──────────────┘  └──────────────┘  └──────────────────────┘ │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │ Gibberish    │  │  Sentence    │  │   Feedback           │ │
│  │  Detector    │  │  Analyzer    │  │   Generator          │ │
│  └──────────────┘  └──────────────┘  └──────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Analysis Components

### 1. Basic Metrics

| Metric | Description | Calculation |
|---|---|---|
| `word_count` | Total words in text | Simple tokenization |
| `sentence_count` | Total sentences | spaCy sentence segmentation |
| `sentence_length` | Avg/min/max/std | From sentence lengths |
| `paragraphs` | Paragraph count | Split by newlines |
| `filler_words` | "este", "eh", "o sea", etc. | Keyword matching |

### 2. Vocabulary Richness

```python
vocabulary_richness = unique_lemmas / total_tokens
```

- Uses lemmatization to normalize words
- OOV (Out-of-Vocabulary) ratio indicates text complexity

### 3. Readability

Fernández-Huerta score (Spanish adaptation of Flesch Reading Ease):

```
FH = 206.25 - 1.099 * (total_words / total_sentences) - 84.26 * (total_polysyllables / total_words)
```

| Score Range | Level |
|---|---|
| 0-30 | Muy difícil |
| 30-50 | Difícil |
| 50-60 | Normal |
| 60-70 | Fácil |
| 70-100 | Muy fácil |

### 4. Structural Analysis

| Check | Description |
|---|---|
| `has_introduction` | Detects introduction patterns |
| `has_conclusion` | Detects conclusion patterns |
| `paragraphs` | Count and structure |

### 5. Requirement Matching

For each requirement in the `requirements` array:
- Tokenize and lemmatize the requirement
- Find matching keywords in the text
- Calculate match score (0-1)
- Return matched keywords

### 6. Gibberish Detection

Uses lexical diversity and coherence metrics:
- Very low unique words → potential gibberish
- Repeated phrases → potential gibberish
- No coherent sentence structure → potential gibberish

## Scoring Algorithm

### Score Breakdown

| Component | Weight | Description |
|---|---|---|
| Coverage de requisitos | 30% | How well requirements are covered |
| Estructura | 25% | Paragraph organization, transitions |
| Calidad lingüística | 25% | Grammar, vocabulary, readability |
| Longitud | 20% | Within min/max word bounds |

### Detailed Scoring

```python
score_breakdown = {
    "cobertura_requisitos": {
        "score": 0-100,
        "weight": 0.30,
        "feedbacks": ["Requirement X partially covered", ...]
    },
    "estructura": {
        "score": 0-100,
        "weight": 0.25,
        "feedbacks": ["Missing conclusion", ...]
    },
    "calidad_linguistica": {
        "score": 0-100,
        "weight": 0.25,
        "feedbacks": ["Good vocabulary richness", ...]
    },
    "longitud": {
        "score": 0-100,
        "weight": 0.20,
        "feedbacks": ["Within word limit", ...]
    }
}

final_score = sum(component.score * component.weight for component in score_breakdown)
```

## API Request/Response

### Request

```json
POST /analyze/text
Content-Type: application/json

{
  "text": "El liderazgo es una habilidad fundamental para el éxito profesional...",
  "requirements": [
    "Incluir una introducción clara del tema",
    "Dar ejemplos concretos",
    "Mencionar beneficios"
  ],
  "min_words": 100,
  "max_words": 500
}
```

### Response

```json
{
  "word_count": 245,
  "sentence_count": 12,
  "sentence_length": {
    "avg": 20.4,
    "min": 8,
    "max": 35,
    "std": 6.2
  },
  "sentence_analysis": {
    "sentence_count": 12,
    "avg_length": 20.4,
    "std_length": 6.2,
    "connector_ratio": 0.35
  },
  "paragraphs": {
    "paragraph_count": 4,
    "has_introduction": true,
    "has_conclusion": true
  },
  "vocabulary_richness": 0.42,
  "oov_ratio": 0.05,
  "readability": {
    "flesch": 58.5,
    "fernandez_huerta": 55.2,
    "label": "Normal"
  },
  "filler_words": 3,
  "keywords": ["liderazgo", "habilidad", "profesional", ...],
  "requirements": [
    {
      "requirement": "Incluir una introducción clara",
      "matched": true,
      "score": 0.95,
      "keywords_found": ["introducción", "clave"]
    }
  ],
  "gibberish_detected": false,
  "score": 87,
  "score_breakdown": {
    "cobertura_requisitos": {"score": 95, "weight": 0.30, "feedbacks": [...]},
    "estructura": {"score": 85, "weight": 0.25, "feedbacks": [...]},
    "calidad_linguistica": {"score": 88, "weight": 0.25, "feedbacks": [...]},
    "longitud": {"score": 100, "weight": 0.20, "feedbacks": [...]}
  },
  "feedback": [
    "Muy buena estructura conclusiva",
    "Considera usar más conectores",
    "Excelente vocabulario técnico"
  ]
}
```

## Deployment (Docker)

The analyzer runs as a separate service that the backend connects to via internal Docker network.

### docker-compose.yml entry
```yaml
analyzer:
  build: ./analyzer
  container_name: voxlab-analyzer
  ports:
    - "8001:8000"
  environment:
    - PYTHONPATH=/app
```

### Build Dockerfile
```dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### Requirements
```
fastapi>=0.100.0
uvicorn>=0.22.0
spacy>=3.5.0
es-core-news-lg @ https://github.com/explosion/spacy-models/releases/download/es_core_news_lg-3.5.0/es_core_news_lg-3.5.0.tar.gz
textstat>=0.7.2
scikit-learn>=1.2.0
numpy>=1.24.0
```

## Performance Considerations

- **Timeout**: 180 seconds (Go client side)
- **Retry**: 3 attempts before returning 502
- **Memory**: ~500MB RAM (spaCy model is large)
- **CPU**: Moderate (parsing and analysis)

## Extending the Analyzer

To add new analysis features:

1. Add new field to `AnalyzeRequest` and `AnalyzeResponse` in `analyzer/client.go`
2. Add analysis function in the Python service
3. Update the feedback generator
4. Update scoring weights if needed
## Speech analysis (`POST /analyze/speech`)

Exposed by the API as `POST /api/v1/exercises/analyze-speech` (JWT required,
`multipart/form-data`):

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `audio` | file | yes | m4a/aac, wav, webm/opus or mp3, max 15 MB |
| `min_seconds` | number | no | Minimum expected duration |
| `target_seconds` | number | no | Target duration |
| `topic` | string | no | Exercise topic, used by the AI coaching tip (max 200 chars) |

Pipeline (`analyzer/analyzers/speech.py`):

1. **Transcription** with [faster-whisper](https://github.com/SYSTRAN/faster-whisper)
   on CPU (int8), Spanish, word-level timestamps, `beam_size=5` (greedy
   decoding looped and triggered temperature fallback: ~6× slower and worse
   on a 28 s test take). The model is downloaded at
   image build time and preloaded in a background thread at startup
   (`WHISPER_WARMUP=0` to disable), so the first take does not pay the load.
   One transcription runs at a time (`SPEECH_CONCURRENCY`); with more than
   `SPEECH_MAX_WAITING` takes queued the analyzer answers 503 and the API
   forwards it as 503 so the app can retry.
2. **Metrics** from the word timestamps: words per minute (fillers excluded),
   pauses (gaps ≥ 0.6 s), long pauses (≥ 3 s), speaking time, filler words.
3. **Fillers**: pure hesitations (`eh`, `mmm`, `ehm`…) always count;
   contextual ones (`este`, `bueno`, `pues`, `tipo`…) only when they stand
   alone (followed by a comma or a pause), so "este proyecto" is not a filler;
   multi-word ones (`o sea`, `como que`, `es decir`).
4. **Score** (0-100), weighted:

| Component | Weight | Ideal |
|-----------|--------|-------|
| `ritmo` | 25 % | 110-160 words per minute |
| `muletillas` | 25 % | ≤ 1 filler per minute |
| `pausas` | 20 % | 3-10 pauses per minute, −10 per pause > 3 s |
| `fluidez` | 15 % | speaking ≥ 65 % of the take |
| `duracion` | 15 % | reach `min_seconds`, don't exceed `target_seconds` by 25 % |

Response (`data` in the API envelope): `transcript`, `duration_seconds`,
`speaking_seconds`, `word_count`, `words_per_minute`, `pause_count`,
`long_pause_count`, `longest_pause_seconds`, `filler_words`,
`filler_breakdown`, `score`, `score_breakdown` (same shape as the text
analyzer), `feedback`, and `ai_feedback` / `ai_model` (see below; `null`
when the LLM was not used).

### Configuration

| Variable | Default | Notes |
|----------|---------|-------|
| `WHISPER_MODEL` | `base` | `small` transcribes better but was ~10× slower on CPU in tests |
| `WHISPER_THREADS` | `2` | CPU threads for transcription |
| `WHISPER_PROMPT` | natural sentence with fillers | Nudges Whisper to keep hesitations |
| `WHISPER_BEAM_SIZE` | `5` | 1 (greedy) was slower in tests, see above |
| `WHISPER_WARMUP` | `1` | Preload the model at startup |
| `SPEECH_CONCURRENCY` | `1` | Simultaneous transcriptions |
| `SPEECH_MAX_WAITING` | `4` | Queued takes before answering 503 |

Findings from testing with `base` (kept in the code comments):
- A prompt made only of fillers ("Eh... este... mmm") is echoed back as the
  transcript; a natural sentence with fillers is not. Echoes are dropped.
- `vad_filter=True` sent the model into repetition loops after long pauses,
  so it stays off; pauses are measured from word timestamps instead.
- faster-whisper 1.2 calls `av.open(metadata_errors=...)`, removed in PyAV 16,
  so `av` is pinned below 16.

Tests: `cd analyzer && uv run --group dev pytest tests` (no model needed).

## AI coaching with an LLM (OpenRouter, optional)

`analyzer/analyzers/llm.py` adds a short coaching tip on top of the local
analysis. The local metrics measure form (length, structure, pace, pauses,
fillers); the LLM is used only for what they cannot see: whether the content
answers the prompt, and one concrete tip about the ideas.

It is called **only when needed**:

- `OPENROUTER_API_KEY` is set (without it nothing changes),
- writing: ≥ 15 words, and either a requirement is ambiguous (embedding
  similarity between 0.30 and 0.55, where MiniLM is unreliable) or the
  score is below 85. The LLM verdict on ambiguous requirements replaces the
  embedding verdict and the score is recomputed,
- speech: the transcript has ≥ 25 words,
- the daily budget (`LLM_DAILY_LIMIT`) is not exhausted,
- the same request is not in the in-memory cache (256 entries).

Any failure (timeout, rate limit, invalid JSON) falls back silently to the
local result. The tip is returned as `ai_feedback` (also prepended to
`feedback`) and the model used as `ai_model`.

| Variable | Default | Notes |
|----------|---------|-------|
| `OPENROUTER_API_KEY` | empty | Key from https://openrouter.ai/keys |
| `OPENROUTER_MODELS` | `google/gemma-4-31b-it:free,openrouter/free` | Comma-separated, tried in order (OpenRouter `models` fallback). Free models change over time; check https://openrouter.ai/models?max_price=0 |
| `LLM_DAILY_LIMIT` | `300` | Calls per day per analyzer process |
| `LLM_TIMEOUT_SECONDS` | `20` | Per call |

Free models on OpenRouter are rate limited (about 20 requests/minute, and
50 requests/day on accounts without credits, 1000/day after buying $10 of
credits once). Keep `LLM_DAILY_LIMIT` under the account limit.

Tests: `tests/test_llm.py` (no network).
