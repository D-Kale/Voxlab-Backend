"""
Voxlab Analyzer — Microservicio de análisis de texto.

Endpoints:
    GET  /health          — Healthcheck con verificación de modelos
    POST /analyze/text    — Analiza un texto de escritura

El análisis incluye detección de gibberish, métricas estructurales,
calidad lingüística, matching semántico de requisitos y scoring
ponderado. Ver docs/scoring.md para la documentación de la fórmula.
"""

from contextlib import asynccontextmanager

from fastapi import FastAPI
from pydantic import BaseModel
from typing import List, Optional

from analyzers.text import analyze_text
from analyzers.nlp import init_nlp, get_nlp
from analyzers.requirements import init_embedding_model, _get_model


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_nlp()
    init_embedding_model()
    yield


app = FastAPI(
    title="Voxlab Analyzer",
    version="0.2.0",
    description="Microservicio de análisis NLP para ejercicios de escritura y oratoria.",
    lifespan=lifespan,
)


class AnalyzeRequest(BaseModel):
    text: str
    requirements: Optional[List[str]] = None
    min_words: Optional[int] = None
    max_words: Optional[int] = None


class SentenceLength(BaseModel):
    avg: float
    min: int
    max: int
    std: float


class Paragraphs(BaseModel):
    paragraph_count: int
    has_introduction: bool
    has_conclusion: bool


class SentenceAnalysis(BaseModel):
    sentence_count: int
    avg_length: float
    std_length: float
    connector_ratio: float


class Readability(BaseModel):
    flesch: float
    fernandez_huerta: float
    label: str


class RequirementResult(BaseModel):
    requirement: str
    matched: bool
    score: float
    keywords_found: List[str]


class ScoreBreakdownItem(BaseModel):
    score: int
    weight: float
    feedbacks: List[str]


class AnalyzeResponse(BaseModel):
    word_count: int
    sentence_count: int
    sentence_length: SentenceLength
    sentence_analysis: SentenceAnalysis
    paragraphs: Paragraphs
    vocabulary_richness: float
    oov_ratio: float
    readability: Readability
    filler_words: int
    keywords: List[str]
    requirements: List[RequirementResult]
    gibberish_detected: bool
    score: int
    score_breakdown: dict
    feedback: List[str]


@app.get("/health")
def health():
    statuses = {}
    ok = True

    try:
        nlp = get_nlp()
        nlp("prueba")
        statuses["spacy"] = "ok"
    except Exception as e:
        statuses["spacy"] = f"error: {e}"
        ok = False

    try:
        model = _get_model()
        list(model.embed(["prueba"]))
        statuses["embeddings"] = "ok"
    except Exception as e:
        statuses["embeddings"] = f"error: {e}"
        ok = False

    code = 200 if ok else 503
    return {
        "status": "ok" if ok else "degraded",
        "models": statuses,
    }


@app.post("/analyze/text", response_model=AnalyzeResponse)
def analyze(req: AnalyzeRequest):
    """Analiza un texto de escritura y devuelve métricas + score.

    El scoring usa una fórmula ponderada con 4 componentes:
        - cobertura_requisitos (30%): similitud semántica MiniLM
        - estructura (25%): párrafos, conectores, variedad
        - calidad_linguistica (25%): TTR, legibilidad, muletillas
        - longitud (20%): cumplimiento de min/max palabras

    Si se detecta gibberish, el score es 0 y se indica en gibberish_detected.
    """
    result = analyze_text(
        text=req.text,
        requirements_list=req.requirements,
        min_words=req.min_words,
        max_words=req.max_words,
    )
    return AnalyzeResponse(**result)
