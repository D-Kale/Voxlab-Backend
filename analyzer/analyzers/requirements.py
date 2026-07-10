from typing import List, Optional

import numpy as np

_MODEL = None


def init_embedding_model() -> None:
    global _MODEL
    from fastembed import TextEmbedding
    _MODEL = TextEmbedding(
        model_name="sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2",
    )


def _get_model():
    if _MODEL is None:
        raise RuntimeError(
            "Embedding model not initialized. Call init_embedding_model() first."
        )
    return _MODEL


def match_requirements(text: str, requirements: List[str]) -> List[dict]:
    if not requirements:
        return []

    model = _get_model()

    text_embedding = list(model.embed([text]))[0]
    req_embeddings = np.array(list(model.embed(requirements)))

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


def precompute_embeddings(requirements: List[str]) -> List[List[float]]:
    if not requirements:
        return []

    model = _get_model()
    embeddings = list(model.embed(requirements))
    return [emb.tolist() for emb in embeddings]
