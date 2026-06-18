"""
Sistema de puntuación ponderada para ejercicios de escritura.

La fórmula está definida como constante (SCORE_FORMULA) para que sea
transparente y documentable. Cada componente devuelve un score 0-100
y feedbacks específicos.

Ver docs/scoring.md para la documentación completa de la fórmula.
"""

from typing import Dict, List, Optional

SCORE_FORMULA = {
    "cobertura_requisitos": {
        "peso": 0.30,
        "descripcion": (
            "Evalúa si el texto cubre los requisitos del ejercicio "
            "usando similitud semántica (embeddings MiniLM)."
        ),
    },
    "estructura": {
        "peso": 0.25,
        "descripcion": (
            "Analiza la organización del texto: párrafos, conectores "
            "textuales y variedad en la longitud de oraciones."
        ),
    },
    "calidad_linguistica": {
        "peso": 0.25,
        "descripcion": (
            "Mide riqueza léxica (TTR), legibilidad "
            "(Fernández-Huerta), ausencia de muletillas y "
            "proporción de palabras conocidas."
        ),
    },
    "longitud": {
        "peso": 0.20,
        "descripcion": (
            "Verifica que el texto cumpla con los límites de palabras "
            "mínimos y máximos configurados en el ejercicio."
        ),
    },
}


def score_requirements(requirements: List[dict]) -> tuple:
    """Evalúa cobertura de requisitos mediante similitud semántica.

    Cada requirement tiene un score entre 0 y 1. Se promedian todos.
    Un score de 0.6+ (promedio) equivale a 60+ puntos.

    Args:
        requirements: Lista de resultados de match_requirements().

    Returns:
        (score_0_100, feedbacks).
    """
    if not requirements:
        return 0, []

    avg_sim = sum(r["score"] for r in requirements) / len(requirements)
    score = int(avg_sim * 100)
    score = max(0, min(100, score))

    feedback = []
    for r in requirements:
        if r["matched"]:
            feedback.append(f'Requisito cumplido: "{r["requirement"]}"')
        else:
            feedback.append(f'Requisito no cumplido: "{r["requirement"]}" — intentá incluir este tema en tu texto.')

    return score, feedback


def score_structure(metrics: dict) -> tuple:
    """Evalúa la estructura del texto.

    Componentes:
        - Párrafos (0-40 pts): 1 párrafo = 10, 2 = 25, 3+ = 40
        - Conclusión (0-20 pts): presente si >= 3 párrafos
        - Variedad de oraciones (0-20 pts): std > 4 = 20, > 2 = 10
        - Conectores (0-20 pts): ratio >= 0.30 = 20, >= 0.15 = 10

    Returns:
        (score_0_100, feedbacks).
    """
    feedback = []
    score = 0

    paragraphs = metrics.get("paragraphs", {})
    par_count = paragraphs.get("paragraph_count", 0)

    if par_count >= 3:
        score += 40
        feedback.append("Buena estructura: el texto tiene introducción, desarrollo y conclusión.")
    elif par_count >= 2:
        score += 25
        feedback.append("El texto tiene una estructura básica con varios párrafos.")
    else:
        score += 10
        feedback.append("El texto tiene un solo párrafo. Separar en varios mejora la organización.")

    if paragraphs.get("has_conclusion"):
        score += 20

    sent = metrics.get("sentence_analysis", {})
    std_length = sent.get("std_length", 0)
    if std_length > 4:
        score += 20
        feedback.append("Buena variedad en la longitud de las oraciones.")
    elif std_length > 2:
        score += 10

    connector_ratio = sent.get("connector_ratio", 0)
    if connector_ratio >= 0.30:
        score += 20
        feedback.append("Uso adecuado de conectores textuales.")
    elif connector_ratio >= 0.15:
        score += 10
    elif sent.get("sentence_count", 0) > 3:
        feedback.append("Podrías usar más conectores textuales para mejorar la fluidez.")

    return min(score, 100), feedback


def score_linguistic_quality(metrics: dict) -> tuple:
    """Evalúa la calidad lingüística del texto.

    Componentes:
        - Riqueza léxica TTR (0-30 pts): lineal entre 0.30 y 0.60
        - Legibilidad (0-30 pts): óptimo en 50-80 (Normal)
        - Muletillas (0-20 pts): 0 = 20, 1-2 = 10, 3+ = 0
        - OOV (0-20 pts): < 10% = 20, 10-25% = 10

    Returns:
        (score_0_100, feedbacks).
    """
    feedback = []
    score = 0

    richness = metrics.get("vocabulary_richness", 0)
    if richness >= 0.60:
        score += 30
        feedback.append("Excelente variedad de vocabulario.")
    elif richness >= 0.45:
        score += 20
        feedback.append("Buena variedad de vocabulario.")
    elif richness >= 0.30:
        score += 10
    else:
        feedback.append("Repetís muchas palabras. Intentá usar sinónimos.")

    readability = metrics.get("readability", {})
    label = readability.get("label", "")
    fh_score = readability.get("fernandez_huerta", 0)
    if 50 <= fh_score <= 80:
        score += 30
        feedback.append("Nivel de legibilidad adecuado.")
    elif 30 <= fh_score < 50 or 80 < fh_score <= 90:
        score += 15
        if fh_score > 80:
            feedback.append("El texto es muy fácil de leer. Podrías usar vocabulario más específico.")
        else:
            feedback.append("El texto es algo difícil de leer. Intentá oraciones más cortas.")
    else:
        if label in ("Muy difícil",):
            feedback.append("El texto es difícil de leer. Usá oraciones más cortas.")

    filler = metrics.get("filler_words", 0)
    if filler == 0:
        score += 20
    elif filler <= 2:
        score += 10
        feedback.append(f"Encontramos {filler} palabra(s) de relleno.")
    else:
        feedback.append(f"Evitá las muletillas ('este', 'eh', 'o sea'). Encontramos {filler}.")

    oov = metrics.get("oov_ratio", 0)
    if oov < 0.10:
        score += 20
    elif oov < 0.25:
        score += 10

    return min(score, 100), feedback


def score_length(metrics: dict, min_words: Optional[int], max_words: Optional[int]) -> tuple:
    """Evalúa el cumplimiento de la longitud esperada.

    Si no hay min ni max definidos, se asume 100 (sin restricción).
    La puntuación es lineal: a los 50% del mínimo da 50 pts.

    Args:
        metrics: Diccionario con word_count.
        min_words: Mínimo de palabras requerido (del ejercicio).
        max_words: Máximo de palabras permitido (del ejercicio).

    Returns:
        (score_0_100, feedbacks).
    """
    feedback = []
    word_count = metrics.get("word_count", 0)

    if word_count == 0:
        return 0, ["El texto está vacío."]

    if not min_words and not max_words:
        return 100, []

    if min_words and word_count < min_words:
        pct = word_count / min_words
        score = int(pct * 100)
        feedback.append(f"El texto tiene {word_count} palabras ({pct*100:.0f}% del mínimo de {min_words}).")
        return max(0, min(100, score)), feedback

    if max_words and word_count > max_words:
        excess_ratio = (word_count - max_words) / max_words
        score = max(0, int((1 - excess_ratio) * 100))
        feedback.append(f"El texto excede el máximo de {max_words} palabras por {word_count - max_words}.")
        return max(0, min(100, score)), feedback

    if min_words and max_words:
        mid = (min_words + max_words) / 2
        if word_count == mid:
            feedback.append("Longitud ideal.")
        else:
            feedback.append("Dentro del rango de palabras esperado.")
    elif min_words:
        feedback.append(f"Cumple el mínimo de {min_words} palabras.")
    elif max_words:
        feedback.append(f"Dentro del máximo de {max_words} palabras.")

    return 100, feedback


def calculate_score(
    metrics: dict,
    min_words: Optional[int] = None,
    max_words: Optional[int] = None,
) -> dict:
    """Calcula el score total ponderado y el desglose.

    Recorre cada componente de SCORE_FORMULA, ejecuta su scorer,
    pondera y acumula. Devuelve score total + breakdown completo.

    Args:
        metrics: Diccionario completo de métricas (de analyze_text).
        min_words: Mínimo de palabras configurado en el ejercicio.
        max_words: Máximo de palabras configurado en el ejercicio.

    Returns:
        dict con:
            - score: int 0-100
            - score_breakdown: dict {componente: {score, weight, feedbacks}}
            - feedback: lista de feedbacks agregados
    """
    breakdown = {}
    all_feedback = []

    for key, cfg in SCORE_FORMULA.items():
        if key == "cobertura_requisitos":
            s, fb = score_requirements(metrics.get("requirements", []))
        elif key == "estructura":
            s, fb = score_structure(metrics)
        elif key == "calidad_linguistica":
            s, fb = score_linguistic_quality(metrics)
        elif key == "longitud":
            s, fb = score_length(metrics, min_words, max_words)
        else:
            s, fb = 0, []

        breakdown[key] = {
            "score": s,
            "weight": cfg["peso"],
            "feedbacks": fb,
        }
        all_feedback.extend(fb)

    total = sum(bd["score"] * bd["weight"] for bd in breakdown.values())
    total = max(0, min(100, int(round(total))))

    return {
        "score": total,
        "score_breakdown": breakdown,
        "feedback": all_feedback,
    }
