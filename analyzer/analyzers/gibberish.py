import re
from typing import Tuple

VOWELS = set("aeiouáéíóúü")
ALPHA_PATTERN = re.compile(r"[a-záéíóúüñ]", re.IGNORECASE)

TOP_SPANISH_BIGRAMS = {
    "el", "es", "en", "la", "de", "qu", "ue", "ad", "al", "ar",
    "as", "co", "on", "os", "ra", "re", "ta", "te", "un", "nt",
    "an", "ac", "do", "ic", "ie", "io", "na", "nc", "no", "pa",
    "pr", "ro", "sa", "se", "si", "st", "tr", "ca", "ci", "ba",
    "be", "bi", "bo", "bu", "da", "di", "du", "ec", "ed", "em",
}


def vowel_consonant_ratio(text: str) -> Tuple[float, bool]:
    """Calcula la proporción de vocales respecto al total de letras.

    El español tiene ~45% de vocales en texto natural. Valores extremos
    (menos de 20% o más de 70%) indican que el texto no es español real.

    Returns:
        (ratio, is_gibberish) — is_gibberish=True si está fuera del rango válido.
    """
    letters = ALPHA_PATTERN.findall(text.lower())
    if not letters:
        return 1.0, True

    vowel_count = sum(1 for ch in letters if ch in VOWELS)
    ratio = vowel_count / len(letters)

    return round(ratio, 4), ratio < 0.20 or ratio > 0.70


def bigram_frequency(text: str) -> Tuple[float, bool]:
    """Mide qué porcentaje de bigramas del texto están en los top-100 bigramas del español.

    Texto real en español tiene una alta proporción de bigramas comunes.
    Si menos del 5% de los bigramas coinciden, probablemente no es español.

    Returns:
        (coverage_ratio, is_gibberish).
    """
    clean = ALPHA_PATTERN.findall(text.lower())
    if len(clean) < 4:
        return 1.0, False

    bigrams = {clean[i] + clean[i + 1] for i in range(len(clean) - 1)}
    if not bigrams:
        return 1.0, False

    matches = sum(1 for bg in bigrams if bg in TOP_SPANISH_BIGRAMS)
    ratio = matches / len(bigrams)

    return round(ratio, 4), ratio < 0.05


def oov_detection(text: str) -> Tuple[float, bool]:
    """Usa spaCy para detectar tokens fuera de vocabulario.

    Si más del 50% de los tokens alfabéticos son OOV, el texto
    probablemente no es español coherente.

    Returns:
        (oov_ratio, is_gibberish).
    """
    from analyzers.nlp import get_nlp

    nlp = get_nlp()
    doc = nlp(text)

    tokens = [t for t in doc if t.is_alpha]
    if not tokens:
        return 0.0, False

    oov_count = sum(1 for t in tokens if not t.has_vector and t.is_oov)
    ratio = oov_count / len(tokens)

    return round(ratio, 4), ratio > 0.50


def is_gibberish(text: str) -> Tuple[bool, str]:
    """Evalúa si un texto es gibberish usando 3 filtros secuenciales.

    Pipeline:
    1. Ratio vocales/consonantes (más rápido, sin dependencias pesadas)
    2. Frecuencia de bigramas españoles (rápido, solo regex)
    3. OOV con spaCy (más lento, requiere modelo cargado)

    Args:
        text: Texto a evaluar.

    Returns:
        (es_gibberish, razón). Si es gibberish, la razón explica qué filtro falló.
    """
    stripped = text.strip()
    if not stripped:
        return True, "El texto está vacío."

    _, bad = vowel_consonant_ratio(stripped)
    if bad:
        return True, "La proporción de vocales no corresponde a texto en español."

    _, bad = bigram_frequency(stripped)
    if bad:
        return True, "Los patrones de letras no corresponden al español."

    _, bad = oov_detection(stripped)
    if bad:
        return True, "Más del 50% de las palabras no fueron reconocidas en español."

    return False, ""
