import re
from typing import List

FILLER_WORDS = {
    "este", "eh", "mmm", "em", "ah", "o sea", "tipo", "digamos",
    "entonces", "como que", "o sea que", "vale", "bueno", "pues",
    "este…", "ehm", "osea",
}

_FILLER_PATTERNS = [
    re.compile(r"\b" + re.escape(w) + r"\b", re.IGNORECASE)
    for w in sorted(FILLER_WORDS, key=len, reverse=True)
]


def lexical_richness(text: str) -> float:
    """Calcula la riqueza léxica (Type-Token Ratio) del texto.

    TTR = lemas únicos / total de lemas. Un valor alto indica
    vocabulario variado. Rangos típicos en español escrito:
        - < 0.30: muy repetitivo
        - 0.30–0.50: normal
        - 0.50–0.70: buena variedad
        - > 0.70: muy rico (poco común en textos extensos)

    Contribuye al componente 'calidad_linguistica' (25% del score total).

    Returns:
        float entre 0.0 y 1.0.
    """
    from analyzers.nlp import get_nlp

    nlp = get_nlp()
    doc = nlp(text.lower())
    lemmas = [token.lemma_ for token in doc if token.is_alpha]

    if not lemmas:
        return 0.0

    unique = set(lemmas)
    return round(len(unique) / len(lemmas), 4)


def count_filler_words(text: str) -> int:
    """Cuenta las palabras de relleno (muletillas) en el texto.

    Las muletillas ('este', 'eh', 'o sea', etc.) restan calidad
    al texto escrito formal. Este conteo contribuye negativamente
    al componente 'calidad_linguistica'.

    Returns:
        int: cantidad total de muletillas encontradas.
    """
    count = 0
    for pattern in _FILLER_PATTERNS:
        count += len(pattern.findall(text))
    return count


def extract_keywords(text: str, top_n: int = 20) -> List[str]:
    """Extrae las palabras clave más frecuentes del texto.

    Usa lematización de spaCy y excluye stop words. Ordena por
    frecuencia descendente.

    Args:
        text: Texto a analizar.
        top_n: Cantidad máxima de keywords a devolver.

    Returns:
        Lista de keywords ordenadas por frecuencia.
    """
    from analyzers.nlp import get_nlp

    nlp = get_nlp()
    doc = nlp(text.lower())

    lemmas = [token.lemma_ for token in doc if token.is_alpha and not token.is_stop]
    freq = {}
    for lemma in lemmas:
        freq[lemma] = freq.get(lemma, 0) + 1

    sorted_kw = sorted(freq.items(), key=lambda x: -x[1])
    return [kw for kw, _ in sorted_kw[:top_n]]


def oov_ratio(text: str) -> float:
    """Calcula la proporción de tokens fuera de vocabulario.

    Usa spaCy para identificar palabras que no están en su diccionario.
    Un ratio bajo (< 0.10) es normal. Entre 0.10 y 0.25 puede indicar
    uso de neologismos o tecnicismos. > 0.25 sugiere problemas.

    NOTA: Este filtro es más permisivo que el de gibberish.py.
    Aquí solo medimos para el score, no para determinar si es gibberish.

    Returns:
        float entre 0.0 y 1.0.
    """
    from analyzers.nlp import get_nlp

    nlp = get_nlp()
    doc = nlp(text)

    tokens = [t for t in doc if t.is_alpha]
    if not tokens:
        return 0.0

    oov_count = sum(1 for t in tokens if not t.has_vector and t.is_oov)
    return round(oov_count / len(tokens), 4)
