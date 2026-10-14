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

import asyncio
import os
import shutil
import tempfile
import threading

from fastapi import FastAPI, File, Form, HTTPException, UploadFile
from fastapi.concurrency import run_in_threadpool
from pydantic import BaseModel
from typing import Dict, List, Optional

from analyzers.text import analyze_text
from analyzers.nlp import init_nlp, get_nlp
from analyzers.requirements import init_embedding_model, _get_model
from analyzers import llm
from analyzers.speech import analyze_speech, warm_up, whisper_model_name

# Whisper usa todos los hilos que le dan: dos transcripciones a la vez en
# un VPS chico se pisan y tardan el doble cada una. Las demás esperan.
_SPEECH_SLOTS = asyncio.Semaphore(int(os.getenv("SPEECH_CONCURRENCY", "1")))
# Más de esto en cola: 503 en lugar de dejar al usuario esperando minutos.
_SPEECH_MAX_WAITING = int(os.getenv("SPEECH_MAX_WAITING", "4"))
_speech_waiting = 0
# Límite de audio que se acepta (el ejercicio más largo es de 3 minutos).
MAX_AUDIO_BYTES = 15 * 1024 * 1024


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_nlp()
    init_embedding_model()
    if os.getenv("WHISPER_WARMUP", "1") == "1":
        threading.Thread(target=warm_up, daemon=True).start()
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
    # Consejo del LLM (OpenRouter) cuando se usó; null si no hizo falta o no
    # hay API key.
    ai_feedback: Optional[str] = None
    ai_model: Optional[str] = None


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

    # Whisper se precarga en segundo plano; acá solo se informa cuál se usa.
    statuses["whisper"] = whisper_model_name()
    statuses["llm"] = "openrouter" if llm.enabled() else "disabled"

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
    ai_feedback: Optional[str] = None
    ai_model: Optional[str] = None


@app.post("/analyze/speech", response_model=SpeechResponse)
async def analyze_speech_endpoint(
    audio: UploadFile = File(...),
    min_seconds: Optional[float] = Form(None),
    target_seconds: Optional[float] = Form(None),
    topic: Optional[str] = Form(None),
):
    """Transcribe un audio de oratoria (español) y mide la entrega.

    El scoring usa 5 componentes (ver analyzers/speech.py):
        - ritmo (25%): palabras por minuto, ideal 110-160
        - muletillas (25%): "eh", "este", "o sea"... por minuto
        - pausas (20%): pausas por minuto, ideal 3-10, penaliza las largas
        - fluidez (15%): proporción del tiempo hablando
        - duracion (15%): cumplimiento del mínimo / objetivo

    El modelo Whisper se precarga al arrancar (WHISPER_MODEL). Si hay
    OPENROUTER_API_KEY y la transcripción tiene contenido, se agrega un
    consejo sobre el mensaje en `ai_feedback` (ver analyzers/llm.py).
    """
    global _speech_waiting
    if _speech_waiting >= _SPEECH_MAX_WAITING:
        raise HTTPException(
            status_code=503,
            detail="Hay muchas grabaciones en análisis. Probá en un momento.",
        )
    suffix = os.path.splitext(audio.filename or "")[1] or ".m4a"
    with tempfile.NamedTemporaryFile(suffix=suffix) as tmp:
        shutil.copyfileobj(audio.file, tmp)
        tmp.flush()
        if tmp.tell() > MAX_AUDIO_BYTES:
            raise HTTPException(status_code=413, detail="El audio es demasiado grande.")
        if tmp.tell() == 0:
            raise HTTPException(status_code=422, detail="El audio está vacío.")
        _speech_waiting += 1
        try:
            async with _SPEECH_SLOTS:
                result = await run_in_threadpool(
                    analyze_speech, tmp.name, min_seconds, target_seconds,
                    (topic or "").strip()[:200] or None,
                )
        except Exception as e:  # audio corrupto o formato no soportado
            raise HTTPException(
                status_code=422,
                detail=f"No se pudo procesar el audio: {e}",
            )
        finally:
            _speech_waiting -= 1
    return SpeechResponse(**result)
