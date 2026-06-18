from typing import Optional
import spacy
from spacy.lang.es import Spanish
from spacy.cli import download as spacy_download

_nlp: Optional[Spanish] = None
_spacy_download_attempted = False


def get_nlp() -> Spanish:
    global _nlp, _spacy_download_attempted
    if _nlp is not None:
        return _nlp

    try:
        _nlp = spacy.load("es_core_news_md")
    except OSError:
        if not _spacy_download_attempted:
            _spacy_download_attempted = True
            try:
                spacy_download("es_core_news_md")
                _nlp = spacy.load("es_core_news_md")
            except Exception:
                pass

        if _nlp is None:
            _nlp = spacy.blank("es")

    return _nlp
