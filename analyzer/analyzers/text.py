"""Orquestador del análisis de texto.

Coordina todos los submódulos (gibberish, structure, vocabulary,
readability, requirements) y devuelve un dict unificado de métricas.

Flujo:
    1. Validación básica (texto vacío)
    2. Detección de gibberish (3 filtros rápidos)
    3. Análisis completo de métricas
    4. Cálculo de score ponderado
"""

from typing import List, Optional

from analyzers.nlp import get_nlp
from analyzers import gibberish, structure, vocabulary, readability, requirements
from analyzers.scoring import calculate_score


def _empty_result(word_count: int, feedback: list) -> dict:
    """Devuelve un resultado mínimo con todos los campos requeridos por AnalyzeResponse."""
    return {
        "word_count": word_count,
        "sentence_count": 0,
        "sentence_length": {"avg": 0, "min": 0, "max": 0, "std": 0},
        "sentence_analysis": {"sentence_count": 0, "avg_length": 0, "std_length": 0, "connector_ratio": 0.0},
        "paragraphs": {"paragraph_count": 0, "has_introduction": False, "has_conclusion": False},
        "vocabulary_richness": 0.0,
        "oov_ratio": 0.0,
        "readability": {"flesch": 0, "fernandez_huerta": 0, "label": "N/A"},
        "filler_words": 0,
        "keywords": [],
        "requirements": [],
        "gibberish_detected": True,
        "score": 0,
        "score_breakdown": {},
        "feedback": feedback,
    }


def analyze_text(
    text: str,
    requirements_list: Optional[List[str]] = None,
    min_words: Optional[int] = None,
    max_words: Optional[int] = None,
) -> dict:
    """Analiza un texto de escritura y devuelve métricas + score.

    Args:
        text: Texto del alumno en español.
        requirements_list: Lista de requisitos del ejercicio.
        min_words: Mínimo de palabras configurado.
        max_words: Máximo de palabras configurado.

    Returns:
        dict con word_count, sentence_count, sentence_length, paragraphs,
        sentence_analysis, vocabulary_richness, oov_ratio, readability,
        filler_words, keywords, requirements (matching semántico),
        gibberish_detected, score, score_breakdown, feedback.
    """
    nlp = get_nlp()
    stripped = text.strip()
    if not stripped:
        return _empty_result(0, ["El texto está vacío."])

    gibberish_detected, gibberish_reason = gibberish.is_gibberish(stripped)
    if gibberish_detected:
        return _empty_result(len(stripped.split()), [gibberish_reason])

    doc = nlp(stripped)
    words = [token for token in doc if token.is_alpha]
    word_count = len(words)

    result = {
        "word_count": word_count,
        "sentence_count": len(list(doc.sents)),
        "gibberish_detected": False,
    }

    result["sentence_length"] = _sentence_length_stats(doc)
    result["sentence_analysis"] = structure.sentence_analysis(stripped)
    result["paragraphs"] = structure.paragraph_structure(stripped)
    result["vocabulary_richness"] = vocabulary.lexical_richness(stripped)
    result["oov_ratio"] = vocabulary.oov_ratio(stripped)
    result["readability"] = readability.readability_score(stripped)
    result["filler_words"] = vocabulary.count_filler_words(stripped)
    result["keywords"] = vocabulary.extract_keywords(stripped, top_n=10)
    result["requirements"] = requirements.match_requirements(stripped, requirements_list or [])

    scoring_result = calculate_score(result, min_words, max_words)
    result["score"] = scoring_result["score"]
    result["score_breakdown"] = scoring_result["score_breakdown"]
    result["feedback"] = scoring_result["feedback"]

    return result


def _sentence_length_stats(doc) -> dict:
    sentences = list(doc.sents)
    if not sentences:
        return {"avg": 0, "min": 0, "max": 0, "std": 0}

    lengths = [len([t for t in s if t.is_alpha]) for s in sentences]
    avg = sum(lengths) / len(lengths)

    if len(lengths) > 1:
        variance = sum((x - avg) ** 2 for x in lengths) / len(lengths)
        std = variance ** 0.5
    else:
        std = 0.0

    return {
        "avg": round(avg, 1),
        "min": min(lengths),
        "max": max(lengths),
        "std": round(std, 1),
    }
