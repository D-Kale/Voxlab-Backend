import re
from typing import List

CONECTORES_ADICION = {"además", "también", "asimismo", "igualmente", "incluso"}
CONECTORES_CONTRASTE = {"sin embargo", "no obstante", "a pesar de", "empero", "en cambio", "por el contrario", "ahora bien"}
CONECTORES_CAUSA = {"por lo tanto", "en consecuencia", "por consiguiente", "así que", "de modo que", "por eso", "entonces", "por ende"}
CONECTORES_ORDEN = {"en primer lugar", "por un lado", "por otro lado", "primero", "luego", "después", "finalmente", "para empezar"}
CONECTORES_EJEMPLO = {"por ejemplo", "es decir", "o sea", "en otras palabras", "tal como"}
CONECTORES_RESUMEN = {"en resumen", "en conclusión", "para finalizar", "en síntesis", "concluyendo", "en resumidas cuentas"}

ALL_CONECTORES = (
    CONECTORES_ADICION
    | CONECTORES_CONTRASTE
    | CONECTORES_CAUSA
    | CONECTORES_ORDEN
    | CONECTORES_EJEMPLO
    | CONECTORES_RESUMEN
)


def paragraph_structure(text: str) -> dict:
    """Analiza la estructura de párrafos del texto.

    Contribuye al componente 'estructura' (25% del score total).
    Un texto bien estructurado tiene al menos 2 párrafos (introducción
    implícita) y 3+ párrafos (indica introducción + desarrollo + conclusión).

    Args:
        text: Texto del alumno en español.

    Returns:
        dict con:
            - paragraph_count: número de párrafos
            - has_introduction: True si hay >= 2 párrafos
            - has_conclusion: True si hay >= 3 párrafos
    """
    paragraphs = [p.strip() for p in text.split("\n") if p.strip()]
    count = len(paragraphs)
    return {
        "paragraph_count": count,
        "has_introduction": count >= 2,
        "has_conclusion": count >= 3,
    }


def sentence_analysis(text: str) -> dict:
    """Analiza las oraciones del texto.

    Evalúa la variedad de longitud de oraciones (textos monótonos
    tienen desviación estándar baja) y la presencia de conectores.

    Args:
        text: Texto del alumno.

    Returns:
        dict con:
            - sentence_count: total de oraciones
            - avg_length: palabras por oración (promedio)
            - std_length: desviación estándar de longitud
            - connector_ratio: proporción de oraciones con al menos un conector
    """
    from analyzers.nlp import get_nlp

    nlp = get_nlp()
    doc = nlp(text)
    sentences = list(doc.sents)

    if not sentences:
        return {"sentence_count": 0, "avg_length": 0, "std_length": 0, "connector_ratio": 0.0}

    lengths = [len([t for t in s if t.is_alpha]) for s in sentences]
    avg = sum(lengths) / len(lengths)

    if len(lengths) > 1:
        variance = sum((x - avg) ** 2 for x in lengths) / len(lengths)
        std = variance ** 0.5
    else:
        std = 0.0

    sentences_with_connector = 0
    for sent in sentences:
        sent_lower = sent.text.lower()
        for conector in ALL_CONECTORES:
            if conector in sent_lower:
                sentences_with_connector += 1
                break

    return {
        "sentence_count": len(sentences),
        "avg_length": round(avg, 1),
        "std_length": round(std, 1),
        "connector_ratio": round(sentences_with_connector / len(sentences), 4),
    }
