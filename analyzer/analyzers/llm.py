"""Coaching con un LLM vía OpenRouter, solo cuando aporta.

El análisis local (spaCy, embeddings, Whisper) mide forma: longitud,
estructura, ritmo, pausas, muletillas. Un LLM sirve para lo que esas
métricas no ven: si el contenido responde a la consigna y un consejo
concreto sobre las ideas. Por costo y latencia se llama solo si:

- hay API key (OPENROUTER_API_KEY); sin ella todo sigue funcionando igual,
- el texto o la transcripción tiene contenido suficiente para opinar,
- quedan llamadas en el presupuesto diario (LLM_DAILY_LIMIT),
- la misma consulta no está ya en caché.

Por defecto usa modelos gratuitos (`:free` y el router `openrouter/free`).
Cualquier error (timeout, rate limit, JSON inválido) devuelve None y el
análisis local se responde sin cambios.
"""

from __future__ import annotations

import datetime
import hashlib
import json
import logging
import os
import threading
import urllib.error
import urllib.request
from collections import OrderedDict
from typing import List, Optional

log = logging.getLogger("analyzer.llm")

OPENROUTER_URL = "https://openrouter.ai/api/v1/chat/completions"
DEFAULT_MODELS = (
    "nvidia/nemotron-3-ultra-550b-a55b:free,"
    "nvidia/nemotron-3-super-120b-a12b:free,"
    "dots-studio/dots-3-note-preview:free"
)

# Mínimos de contenido para pedir opinión al LLM.
MIN_TEXT_WORDS = 15
MIN_SPEECH_WORDS = 25

_CACHE_SIZE = 256
_cache: "OrderedDict[str, dict]" = OrderedDict()
_lock = threading.Lock()
_budget_day: Optional[datetime.date] = None
_budget_used = 0


def api_key() -> str:
    return os.getenv("OPENROUTER_API_KEY", "").strip()


def models() -> List[str]:
    raw = os.getenv("OPENROUTER_MODELS", DEFAULT_MODELS)
    return [m.strip() for m in raw.split(",") if m.strip()]


def enabled() -> bool:
    return bool(api_key())


def _daily_limit() -> int:
    try:
        return int(os.getenv("LLM_DAILY_LIMIT", "300"))
    except ValueError:
        return 300


def _take_budget() -> bool:
    global _budget_day, _budget_used
    today = datetime.date.today()
    with _lock:
        if _budget_day != today:
            _budget_day, _budget_used = today, 0
        if _budget_used >= _daily_limit():
            return False
        _budget_used += 1
        return True


def _cache_get(key: str) -> Optional[dict]:
    with _lock:
        value = _cache.get(key)
        if value is not None:
            _cache.move_to_end(key)
        return value


def _cache_put(key: str, value: dict) -> None:
    with _lock:
        _cache[key] = value
        _cache.move_to_end(key)
        while len(_cache) > _CACHE_SIZE:
            _cache.popitem(last=False)


def _extract_json(content: str) -> Optional[dict]:
    """Los modelos chicos a veces envuelven el JSON en ```json ... ```."""
    start, end = content.find("{"), content.rfind("}")
    if start < 0 or end <= start:
        return None
    try:
        data = json.loads(content[start : end + 1])
    except json.JSONDecodeError:
        return None
    return data if isinstance(data, dict) else None


def ask_json(system: str, user: str, max_tokens: int = 350) -> Optional[dict]:
    """Pide una respuesta JSON. Devuelve {"data": dict, "model": str} o None."""
    if not enabled():
        return None
    key = hashlib.sha256(f"{system}\n{user}".encode()).hexdigest()
    cached = _cache_get(key)
    if cached is not None:
        return cached
    if not _take_budget():
        log.info("LLM daily budget exhausted; skipping")
        return None

    # Free models vary a lot: some are classifiers or spend every token
    # reasoning. Try them one by one and keep the first valid JSON answer.
    for model in models():
        result = _ask_model(model, system, user, max_tokens)
        if result is not None:
            _cache_put(key, result)
            return result
    return None


def _ask_model(model: str, system: str, user: str, max_tokens: int) -> Optional[dict]:
    body = {
        "model": model,
        "messages": [
            {"role": "system", "content": system},
            {"role": "user", "content": user},
        ],
        "temperature": 0.3,
        # Room for models that think before answering; the thinking itself
        # is not returned.
        "max_tokens": max(max_tokens, 900),
        "reasoning": {"exclude": True},
    }
    request = urllib.request.Request(
        OPENROUTER_URL,
        data=json.dumps(body).encode(),
        headers={
            "Authorization": f"Bearer {api_key()}",
            "Content-Type": "application/json",
            "HTTP-Referer": os.getenv("OPENROUTER_REFERER", "https://tvoxlab.tech"),
            "X-Title": "Voxlab",
        },
        method="POST",
    )
    try:
        timeout = float(os.getenv("LLM_TIMEOUT_SECONDS", "20"))
        with urllib.request.urlopen(request, timeout=timeout) as response:
            payload = json.loads(response.read())
        content = payload["choices"][0]["message"]["content"] or ""
        data = _extract_json(content)
        if data is None:
            log.warning("LLM %s returned non-JSON content", model)
            return None
        return {"data": data, "model": payload.get("model", model)}
    except (urllib.error.URLError, TimeoutError, KeyError, IndexError, ValueError) as e:
        log.warning("LLM %s call failed: %s", model, e)
        return None


# ── Escritura ──────────────────────────────────────────────────────────────

_TEXT_SYSTEM = (
    "Sos un profesor de comunicación que corrige ejercicios de escritura en "
    "español de estudiantes. Respondé SOLO con JSON válido, sin texto extra."
)

# Similitud de embeddings donde el matching local es poco confiable.
AMBIGUOUS_LOW, AMBIGUOUS_HIGH = 0.30, 0.55


def review_text(text: str, requirements: List[dict], score: int) -> Optional[dict]:
    """Revisa requisitos dudosos y da un consejo sobre el contenido.

    `requirements` es la salida de match_requirements. Solo se pide
    veredicto de los que quedaron en la franja ambigua. Devuelve
    {"feedback": str, "requirements": {texto: bool}, "model": str} o None.
    """
    if len(text.split()) < MIN_TEXT_WORDS:
        return None
    ambiguous = [
        r["requirement"]
        for r in requirements
        if AMBIGUOUS_LOW <= r.get("score", 0) <= AMBIGUOUS_HIGH
    ]
    # Un texto excelente y sin requisitos dudosos no necesita el LLM.
    if not ambiguous and score >= 85:
        return None

    reqs = "\n".join(f"- {r}" for r in ambiguous) or "(ninguno)"
    user = (
        f"Texto del estudiante:\n\"\"\"\n{text[:4000]}\n\"\"\"\n\n"
        f"Requisitos a verificar:\n{reqs}\n\n"
        "Devolvé este JSON:\n"
        '{"requirements": {"<requisito>": true|false}, '
        '"feedback": "<un consejo concreto sobre las ideas o la claridad, '
        'máximo 2 oraciones, en español rioplatense, tuteando con vos>"}'
    )
    answer = ask_json(_TEXT_SYSTEM, user)
    if answer is None:
        return None
    data = answer["data"]
    verdicts = data.get("requirements") or {}
    feedback = str(data.get("feedback") or "").strip()
    return {
        "feedback": feedback[:400],
        "requirements": {
            r: bool(verdicts[r]) for r in ambiguous if isinstance(verdicts.get(r), bool)
        },
        "model": answer["model"],
    }


# ── Oratoria ───────────────────────────────────────────────────────────────

_SPEECH_SYSTEM = (
    "Sos un coach de oratoria. Recibís la transcripción de un estudiante que "
    "practica hablar en público y sus métricas de entrega. Respondé SOLO con "
    "JSON válido, sin texto extra."
)


def review_speech(transcript: str, topic: Optional[str], metrics: dict) -> Optional[dict]:
    """Consejo sobre el contenido del discurso (estructura, claridad).

    Las métricas de entrega ya las mide Whisper; acá se le pide al LLM solo
    lo que ellas no ven. Devuelve {"feedback": str, "model": str} o None.
    """
    if len(transcript.split()) < MIN_SPEECH_WORDS:
        return None
    topic_line = f"Tema pedido: {topic}\n" if topic else ""
    user = (
        f"{topic_line}"
        f"Transcripción:\n\"\"\"\n{transcript[:4000]}\n\"\"\"\n\n"
        f"Métricas: {metrics.get('words_per_minute')} palabras/min, "
        f"{metrics.get('pause_count')} pausas, "
        f"{metrics.get('filler_words')} muletillas.\n\n"
        "Devolvé este JSON:\n"
        '{"feedback": "<un consejo concreto sobre la estructura o claridad del '
        "mensaje (no sobre ritmo ni muletillas), máximo 2 oraciones, en "
        'español rioplatense, tuteando con vos>"}'
    )
    answer = ask_json(_SPEECH_SYSTEM, user, max_tokens=250)
    if answer is None:
        return None
    feedback = str(answer["data"].get("feedback") or "").strip()
    if not feedback:
        return None
    return {"feedback": feedback[:400], "model": answer["model"]}
