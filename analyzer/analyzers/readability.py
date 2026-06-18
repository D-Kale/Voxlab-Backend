import textstat


def readability_score(text: str) -> dict:
    """Calcula métricas de legibilidad del texto.

    Usa la fórmula de Fernández-Huerta (adaptación al español del
    Flesch Reading Ease). Contribuye al componente 'calidad_linguistica'.

    Rangos Fernández-Huerta:
        - 80–100: Muy fácil (texto infantil/simple)
        - 70–80:  Bastante fácil
        - 60–70:  Normal (texto estándar)
        - 50–60:  Bastante difícil
        - 0–50:   Muy difícil (texto técnico/académico)

    El score óptimo para escritura académica está en 50–70 (normal a
    bastante difícil). Textos muy fáciles restan calidad.

    Returns:
        dict con:
            - flesch: puntuación Flesch Reading Ease
            - fernandez_huerta: puntuación adaptada al español
            - label: etiqueta textual del nivel
    """
    try:
        flesch = textstat.flesch_reading_ease(text)
        fernandez_huerta = textstat.fernandez_huerta(text)
        return {
            "flesch": round(flesch, 1),
            "fernandez_huerta": round(fernandez_huerta, 1),
            "label": _readability_label(fernandez_huerta),
        }
    except Exception:
        return {"flesch": 0, "fernandez_huerta": 0, "label": "N/A"}


def _readability_label(score: float) -> str:
    if score >= 80:
        return "Muy fácil"
    elif score >= 70:
        return "Bastante fácil"
    elif score >= 60:
        return "Normal"
    elif score >= 50:
        return "Bastante difícil"
    else:
        return "Muy difícil"
