import re
import textstat
import spacy
from spacy.lang.es import Spanish
from typing import List, Optional

_nlp: Optional[Spanish] = None


def get_nlp() -> Spanish:
    global _nlp
    if _nlp is None:
        try:
            _nlp = spacy.load("es_core_news_md")
        except OSError:
            _nlp = spacy.blank("es")
    return _nlp


FILLER_WORDS = {
    "este", "eh", "mmm", "em", "ah", "o sea", "tipo", "digamos",
    "entonces", "como que", "o sea que", "vale", "bueno", "pues",
    "este…", "ehm",
}


def extract_keywords(text: str) -> List[str]:
    nlp = get_nlp()
    doc = nlp(text.lower())
    lemmas = [token.lemma_ for token in doc if token.is_alpha and not token.is_stop]
    freq = {}
    for lemma in lemmas:
        freq[lemma] = freq.get(lemma, 0) + 1
    sorted_kw = sorted(freq.items(), key=lambda x: -x[1])
    return [kw for kw, count in sorted_kw[:20]]


def match_requirements(text: str, requirements: List[str]) -> List[dict]:
    text_lower = text.lower()
    results = []
    for req in requirements:
        req_lower = req.lower().strip()
        if not req_lower:
            continue
        nlp = get_nlp()
        req_doc = nlp(req_lower)
        req_lemmas = {token.lemma_ for token in req_doc if token.is_alpha and not token.is_stop}
        text_doc = nlp(text_lower)
        text_lemmas = {token.lemma_ for token in text_doc if token.is_alpha}
        matched = req_lemmas & text_lemmas
        score = len(matched) / max(len(req_lemmas), 1)
        results.append({
            "requirement": req,
            "matched": score >= 0.5,
            "score": round(score, 2),
            "keywords_found": list(matched),
        })
    return results


def count_filler_words(text: str) -> int:
    text_lower = text.lower()
    count = 0
    for filler in FILLER_WORDS:
        count += len(re.findall(r'\b' + re.escape(filler) + r'\b', text_lower))
    return count


def sentence_length_variation(text: str) -> dict:
    nlp = get_nlp()
    doc = nlp(text)
    sentences = list(doc.sents)
    if not sentences:
        return {"avg": 0, "min": 0, "max": 0, "std": 0}
    lengths = [len([t for t in s if t.is_alpha]) for s in sentences]
    avg = sum(lengths) / len(lengths)
    variance = sum((x - avg) ** 2 for x in lengths) / len(lengths)
    return {
        "avg": round(avg, 1),
        "min": min(lengths),
        "max": max(lengths),
        "std": round(variance ** 0.5, 1),
    }


def vocabulary_richness(text: str) -> float:
    nlp = get_nlp()
    doc = nlp(text.lower())
    lemmas = [token.lemma_ for token in doc if token.is_alpha]
    if not lemmas:
        return 0.0
    unique = set(lemmas)
    return round(len(unique) / len(lemmas), 4)


def paragraph_structure(text: str) -> dict:
    paragraphs = [p.strip() for p in text.split("\n") if p.strip()]
    return {
        "paragraph_count": len(paragraphs),
        "has_introduction": len(paragraphs) >= 2,
        "has_conclusion": len(paragraphs) >= 3,
    }


def readability_score(text: str) -> dict:
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


def analyze_text(text: str, requirements: Optional[List[str]] = None) -> dict:
    nlp = get_nlp()
    text_stripped = text.strip()
    if not text_stripped:
        return {"word_count": 0, "error": "Texto vacío"}

    doc = nlp(text_stripped)
    words = [token for token in doc if token.is_alpha]

    word_count = len(words)
    sentence_count = len(list(doc.sents))

    req_results = match_requirements(text_stripped, requirements or [])
    sentence_var = sentence_length_variation(text_stripped)
    richness = vocabulary_richness(text_stripped)
    paragraph = paragraph_structure(text_stripped)
    readability = readability_score(text_stripped)
    filler = count_filler_words(text_stripped)
    keywords = extract_keywords(text_stripped)

    return {
        "word_count": word_count,
        "sentence_count": sentence_count,
        "sentence_length": sentence_var,
        "vocabulary_richness": richness,
        "paragraphs": paragraph,
        "readability": readability,
        "filler_words": filler,
        "keywords": keywords[:10],
        "requirements": req_results,
    }
