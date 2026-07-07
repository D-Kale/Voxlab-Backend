from typing import List, Optional

import numpy as np

_MODEL = None


def _load_model():
    """Carga el modelo de embeddings lazy (una sola vez).

    Usa fastembed (ONNX Runtime) con el mismo modelo multilingüe MiniLM.
    Sin PyTorch, ~20MB de dependencias vs ~1GB de sentence-transformers.
    El modelo se descarga automáticamente en la primera carga
    si no está cacheado (~130MB). Se cachea en memoria después.

    Returns:
        El modelo TextEmbedding, o None si falla la carga.
    """
    global _MODEL
    if _MODEL is None:
        try:
            from fastembed import TextEmbedding
            _MODEL = TextEmbedding(
                model_name="sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2",
            )
        except Exception:
            _MODEL = False
    return _MODEL if _MODEL else None


def match_requirements(text: str, requirements: List[str]) -> List[dict]:
    """Evalúa cada requirement contra el texto usando similitud coseno.

    Codifica tanto el texto completo como cada requirement usando el
    modelo multilingüe MiniLM via ONNX (fastembed), luego calcula la
    similitud coseno entre el embedding del texto y el de cada requirement.

    Args:
        text: Texto del alumno en español.
        requirements: Lista de frases requeridas (ej: "Incluir una introducción").

    Returns:
        Lista de dicts, uno por requirement:
            - requirement: texto del requirement
            - score: similitud coseno (0.0–1.0)
            - matched: True si score >= 0.45
            - keywords_found: lista vacía (se mantiene por compatibilidad)
    """
    if not requirements:
        return []

    model = _load_model()

    if model is None:
        return _fallback_keyword_match(text, requirements)

    try:
        text_embedding = list(model.embed([text]))[0]
        req_embeddings = np.array(list(model.embed(requirements)))
    except Exception:
        return _fallback_keyword_match(text, requirements)

    similarities = np.dot(req_embeddings, text_embedding)

    results = []
    for i, req in enumerate(requirements):
        sim = float(similarities[i])
        results.append({
            "requirement": req,
            "matched": sim >= 0.45,
            "score": round(sim, 4),
            "keywords_found": [],
        })

    return results


def _fallback_keyword_match(text: str, requirements: List[str]) -> List[dict]:
    """Fallback cuando el embedding model no está disponible.

    Usa matching por lemas (spaCy) como el analyzer original.
    """
    from analyzers.nlp import get_nlp
    nlp = get_nlp()
    text_lower = text.lower()
    text_doc = nlp(text_lower)
    text_lemmas = {token.lemma_ for token in text_doc if token.is_alpha}

    results = []
    for req in requirements:
        req_lower = req.lower().strip()
        if not req_lower:
            continue
        req_doc = nlp(req_lower)
        req_lemmas = {token.lemma_ for token in req_doc if token.is_alpha and not token.is_stop}
        matched = req_lemmas & text_lemmas
        score = len(matched) / max(len(req_lemmas), 1)
        results.append({
            "requirement": req,
            "matched": score >= 0.5,
            "score": round(score, 2),
            "keywords_found": list(matched),
        })
    return results


def precompute_embeddings(requirements: List[str]):
    """Precomputa los embeddings de los requirements (para guardar en DB).

    Útil para ejercicios donde los requirements son fijos y se evalúan
    muchas veces. Devuelve los embeddings serializables para almacenar
    junto con el ejercicio.

    Args:
        requirements: Lista de strings de requirements.

    Returns:
        Lista de embeddings (listas de floats) para almacenar.
    """
    if not requirements:
        return []

    model = _load_model()
    embeddings = list(model.embed(requirements))
    return [emb.tolist() for emb in embeddings]
