"""
Voxlab Analyzer — Microservicio de análisis de texto.

Endpoints:
    GET  /health          — Healthcheck con verificación de modelos
    POST /analyze/text    — Analiza un texto de escritura
    POST /analyze/speech  — Transcribe un audio de oratoria y mide la entrega

El análisis incluye detección de gibberish, métricas estructurales,
calidad lingüística, matching semántico de requisitos y scoring
ponderado. Ver docs/scoring.md para la documentación de la fórmula.
"""

from contextlib import asynccontextmanager

import os
import shutil
import tempfile

from fastapi import FastAPI, File, Form, HTTPException, UploadFile
from fastapi.concurrency import run_in_threadpool
from pydantic import BaseModel
from typing import Dict, List, Optional

from analyzers.text import analyze_text
from analyzers.nlp import init_nlp, get_nlp
from analyzers.requirements import init_embedding_model, _get_model
from analyzers.speech import analyze_speech, whisper_model_name


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

    # Whisper se carga bajo demanda; acá solo se informa cuál se usará.
    statuses["whisper"] = f"lazy ({whisper_model_name()})"

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


class SpeechResponse(BaseModel):
    transcript: str
    duration_seconds: float
    speaking_seconds: float
    word_count: int
    words_per_minute: int
    pause_count: int
    long_pause_count: int
    longest_pause_seconds: float
    filler_words: int
    filler_breakdown: Dict[str, int]
    score: int
    score_breakdown: dict
    feedback: List[str]


@app.post("/analyze/speech", response_model=SpeechResponse)
async def analyze_speech_endpoint(
    audio: UploadFile = File(...),
    min_seconds: Optional[float] = Form(None),
    target_seconds: Optional[float] = Form(None),
):
    """Transcribe un audio de oratoria (español) y mide la entrega.

    El scoring usa 5 componentes (ver analyzers/speech.py):
        - ritmo (25%): palabras por minuto, ideal 110-160
        - muletillas (25%): "eh", "este", "o sea"... por minuto
        - pausas (20%): pausas por minuto, ideal 3-10, penaliza las largas
        - fluidez (15%): proporción del tiempo hablando
        - duracion (15%): cumplimiento del mínimo / objetivo

    El modelo Whisper se carga en la primera llamada (WHISPER_MODEL).
    """
    suffix = os.path.splitext(audio.filename or "")[1] or ".m4a"
    with tempfile.NamedTemporaryFile(suffix=suffix) as tmp:
        shutil.copyfileobj(audio.file, tmp)
        tmp.flush()
        try:
            result = await run_in_threadpool(
                analyze_speech, tmp.name, min_seconds, target_seconds
            )
        except Exception as e:  # audio corrupto o formato no soportado
            raise HTTPException(
                status_code=422,
                detail=f"No se pudo procesar el audio: {e}",
            )
    return SpeechResponse(**result)
