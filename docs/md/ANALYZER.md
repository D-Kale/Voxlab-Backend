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