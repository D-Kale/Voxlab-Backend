from typing import Optional
import spacy
from spacy.lang.es import Spanish

_nlp: Optional[Spanish] = None


def init_nlp() -> None:
    global _nlp
    _nlp = spacy.load("es_core_news_md")


def get_nlp() -> Spanish:
    if _nlp is None:
        raise RuntimeError(
            "spaCy model not initialized. Call init_nlp() before get_nlp()."
        )
    return _nlp
