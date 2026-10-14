"""Análisis de oratoria a partir de un audio.

Transcribe con faster-whisper (timestamps por palabra) y mide la entrega:
ritmo (palabras por minuto), pausas, muletillas, fluidez y duración.

La parte de métricas y scoring (`analyze_words`) es pura y no depende del
modelo, así que se testea con listas de palabras sintéticas.
"""

from __future__ import annotations

import os
import re
import threading
import unicodedata
from dataclasses import dataclass
from typing import Iterable, List, Optional

# Silencio entre dos palabras a partir del cual se cuenta una pausa.
MIN_PAUSE_SECONDS = 0.6
# Pausas más largas que esto se marcan como "pausa larga".
LONG_PAUSE_SECONDS = 3.0

# Vacilaciones puras: siempre son muletillas.
HESITATIONS = {
    "eh", "ehh", "ehm", "em", "emm", "eem", "mm", "mmm", "hmm", "ah", "ahh",
    "aah", "uh", "um", "umm", "e", "ee", "eee",
}
# Palabras que solo son muletilla cuando van sueltas (seguidas de una pausa
# o de una coma): "este proyecto" no lo es, "este..., bueno" sí.
CONTEXTUAL_FILLERS = {
    "este", "esteee", "bueno", "pues", "tipo", "digamos", "entonces", "nada",
    "vale", "osea",
}
# Muletillas de varias palabras.
MULTIWORD_FILLERS = [("o", "sea"), ("o", "sea,"), ("como", "que"), ("es", "decir")]

# Le enseña al modelo a transcribir las vacilaciones en lugar de omitirlas.
# Probado con el modelo "base": una lista de muletillas sola (sobre todo con
# puntos suspensivos) hace que Whisper la repita como transcripción; una
# frase natural con muletillas no.
WHISPER_PROMPT = os.getenv(
    "WHISPER_PROMPT",
    "Bueno, eh, hoy les voy a contar, este, algo importante.",
)

SCORE_WEIGHTS = {
    "ritmo": 0.25,
    "muletillas": 0.25,
    "pausas": 0.20,
    "fluidez": 0.15,
    "duracion": 0.15,
}


@dataclass
class Word:
    text: str
    start: float
    end: float


def _normalize(token: str) -> str:
    token = token.strip().lower()
    token = "".join(
        c for c in unicodedata.normalize("NFD", token)
        if unicodedata.category(c) != "Mn"
    )
    return re.sub(r"[^\w]", "", token)


def count_fillers(words: List[Word]) -> dict:
    """Cuenta muletillas por forma normalizada."""
    counts: dict = {}
    norm = [_normalize(w.text) for w in words]
    skip = set()
    for i, n in enumerate(norm):
        if i in skip or not n:
            continue
        for first, second in MULTIWORD_FILLERS:
            if (
                n == _normalize(first)
                and i + 1 < len(norm)
                and norm[i + 1] == _normalize(second)
            ):
                key = f"{first} {_normalize(second)}"
                counts[key] = counts.get(key, 0) + 1
                skip.add(i + 1)
                break
        else:
            if n in HESITATIONS:
                key = "eh" if n.startswith("e") else n
                counts[key] = counts.get(key, 0) + 1
            elif n in CONTEXTUAL_FILLERS:
                raw = words[i].text.strip()
                gap = (
                    words[i + 1].start - words[i].end
                    if i + 1 < len(words)
                    else 0.0
                )
                if raw.endswith((",", "...", "…")) or gap >= 0.35:
                    counts[n] = counts.get(n, 0) + 1
    return counts


def _band_score(value: float, low: float, high: float, slope: float) -> int:
    """100 dentro de [low, high]; baja `slope` puntos por unidad afuera."""
    if low <= value <= high:
        return 100
    distance = low - value if value < low else value - high
    return int(max(0, min(100, round(100 - distance * slope))))


def analyze_words(
    words: List[Word],
    duration: float,
    min_seconds: Optional[float] = None,
    target_seconds: Optional[float] = None,
) -> dict:
    """Métricas de entrega y score a partir de palabras con timestamps."""
    words = [w for w in words if _normalize(w.text)]
    feedback: List[str] = []

    if not words:
        breakdown = {
            key: {
                "score": 0,
                "weight": weight,
                "feedbacks": ["No detectamos tu voz en la grabación."],
            }
            for key, weight in SCORE_WEIGHTS.items()
        }
        return {
            "transcript": "",
            "duration_seconds": round(duration, 2),
            "speaking_seconds": 0.0,
            "word_count": 0,
            "words_per_minute": 0,
            "pause_count": 0,
            "long_pause_count": 0,
            "longest_pause_seconds": 0.0,
            "filler_words": 0,
            "filler_breakdown": {},
            "score": 0,
            "score_breakdown": breakdown,
            "feedback": [
                "No detectamos tu voz. Acercate al micrófono y hablá con un "
                "volumen claro."
            ],
        }

    span_start, span_end = words[0].start, words[-1].end
    span = max(span_end - span_start, 0.1)

    gaps = [b.start - a.end for a, b in zip(words, words[1:])]
    pauses = [g for g in gaps if g >= MIN_PAUSE_SECONDS]
    long_pauses = [g for g in pauses if g >= LONG_PAUSE_SECONDS]
    pause_time = sum(pauses)
    speaking = max(span - pause_time, 0.0)

    fillers = count_fillers(words)
    filler_total = sum(fillers.values())
    content_words = max(len(words) - filler_total, 0)
    minutes = span / 60
    wpm = round(content_words / minutes) if minutes > 0 else 0
    pauses_per_minute = len(pauses) / minutes if minutes > 0 else 0
    fillers_per_minute = filler_total / max(minutes, 0.25)
    fluency_ratio = speaking / span

    # Ritmo: 110-160 ppm es cómodo de escuchar en español.
    ritmo = _band_score(wpm, 110, 160, 1.5)
    ritmo_fb = []
    if wpm < 110:
        ritmo_fb.append(
            f"Hablaste a {wpm} palabras por minuto: un poco lento. Apuntá a "
            "120-150."
        )
    elif wpm > 160:
        ritmo_fb.append(
            f"Hablaste a {wpm} palabras por minuto: un poco rápido. Respirá "
            "entre ideas."
        )

    # Muletillas: hasta ~1 por minuto pasa desapercibida.
    muletillas = int(max(0, min(100, round(100 - max(0, fillers_per_minute - 1) * 15))))
    muletillas_fb = []
    if filler_total:
        top = sorted(fillers.items(), key=lambda kv: -kv[1])[:3]
        listed = ", ".join(f'"{k}" ({v})' for k, v in top)
        muletillas_fb.append(
            f"Usaste {filler_total} muletillas: {listed}. Reemplazalas por un "
            "silencio breve."
        )

    # Pausas: entre 3 y 10 por minuto marcan bien las ideas.
    pausas = _band_score(pauses_per_minute, 3, 10, 8)
    pausas -= 10 * len(long_pauses)
    pausas = int(max(0, min(100, pausas)))
    pausas_fb = []
    if pauses_per_minute < 3:
        pausas_fb.append(
            "Casi no hiciste pausas. Una pausa breve después de cada idea "
            "ayuda a que se entienda."
        )
    if long_pauses:
        pausas_fb.append(
            f"Hiciste {len(long_pauses)} pausa(s) de más de "
            f"{int(LONG_PAUSE_SECONDS)} segundos; prepará la siguiente idea "
            "antes de soltar la anterior."
        )

    fluidez = _band_score(fluency_ratio, 0.65, 0.95, 200)
    fluidez_fb = []
    if fluency_ratio < 0.65:
        fluidez_fb.append(
            f"Estuviste en silencio el {round((1 - fluency_ratio) * 100)} % "
            "del tiempo. Sostené la voz con frases completas."
        )

    duracion = 100
    duracion_fb = []
    if min_seconds and duration < min_seconds:
        duracion = int(round(duration / min_seconds * 70))
        duracion_fb.append(
            f"Tu toma duró {round(duration)} s; el mínimo era "
            f"{round(min_seconds)} s."
        )
    elif target_seconds and duration > target_seconds * 1.25:
        duracion = 80
        duracion_fb.append("Te pasaste del tiempo objetivo; practicá ser más breve.")

    scores = {
        "ritmo": (ritmo, ritmo_fb),
        "muletillas": (muletillas, muletillas_fb),
        "pausas": (pausas, pausas_fb),
        "fluidez": (fluidez, fluidez_fb),
        "duracion": (duracion, duracion_fb),
    }
    breakdown = {
        key: {"score": s, "weight": SCORE_WEIGHTS[key], "feedbacks": fb}
        for key, (s, fb) in scores.items()
    }
    total = round(sum(s * SCORE_WEIGHTS[k] for k, (s, _) in scores.items()))

    best = max(scores, key=lambda k: scores[k][0])
    praise = {
        "ritmo": f"Buen ritmo: {wpm} palabras por minuto se escuchan cómodas.",
        "muletillas": "Casi sin muletillas: tu mensaje se escuchó limpio.",
        "pausas": f"Buen uso de las pausas ({len(pauses)}): le diste aire a tus ideas.",
        "fluidez": "Hablaste con fluidez, sin cortes largos.",
        "duracion": "Sostuviste tu discurso el tiempo necesario.",
    }
    feedback.append(praise[best])

    return {
        "transcript": " ".join(w.text.strip() for w in words).strip(),
        "duration_seconds": round(duration, 2),
        "speaking_seconds": round(speaking, 2),
        "word_count": len(words),
        "words_per_minute": wpm,
        "pause_count": len(pauses),
        "long_pause_count": len(long_pauses),
        "longest_pause_seconds": round(max(pauses, default=0.0), 2),
        "filler_words": filler_total,
        "filler_breakdown": fillers,
        "score": int(total),
        "score_breakdown": breakdown,
        "feedback": feedback,
    }


# ── Transcripción ──────────────────────────────────────────────────────────

_MODEL = None
_MODEL_LOCK = threading.Lock()


def whisper_model_name() -> str:
    return os.getenv("WHISPER_MODEL", "base")


def _get_model():
    """Carga el modelo bajo demanda: el analizador de texto no paga su RAM."""
    global _MODEL
    with _MODEL_LOCK:
        if _MODEL is None:
            from faster_whisper import WhisperModel

            _MODEL = WhisperModel(
                whisper_model_name(),
                device="cpu",
                compute_type="int8",
                cpu_threads=int(os.getenv("WHISPER_THREADS", "2")),
            )
        return _MODEL


def transcribe(path: str) -> tuple[List[Word], float]:
    model = _get_model()
    segments, info = model.transcribe(
        path,
        language="es",
        word_timestamps=True,
        initial_prompt=WHISPER_PROMPT,
        # El filtro VAD hizo entrar al modelo en bucles de repetición con
        # pausas largas; sin él las pausas quedan medidas igual.
        vad_filter=False,
        condition_on_previous_text=False,
    )
    words: List[Word] = []
    for segment in segments:
        for w in segment.words or []:
            words.append(Word(text=w.word, start=w.start, end=w.end))
    return drop_prompt_echo(words), float(info.duration)


def drop_prompt_echo(words: List[Word]) -> List[Word]:
    """Descarta una transcripción que solo repite el prompt (alucinación de
    Whisper sobre audio sin voz)."""
    said = "".join(_normalize(w.text) for w in words)
    prompt = "".join(_normalize(t) for t in WHISPER_PROMPT.split())
    return [] if said and said == prompt else words


def analyze_speech(
    path: str,
    min_seconds: Optional[float] = None,
    target_seconds: Optional[float] = None,
) -> dict:
    words, duration = transcribe(path)
    return analyze_words(words, duration, min_seconds, target_seconds)


def words_from(items: Iterable[tuple]) -> List[Word]:
    """Atajo para tests: [(texto, inicio, fin), ...]."""
    return [Word(t, s, e) for t, s, e in items]
