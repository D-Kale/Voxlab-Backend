from typing import List, Optional


def generate_feedback(metrics: dict, requirements: Optional[List[str]] = None) -> List[str]:
    feedback = []

    wc = metrics.get("word_count", 0)
    if wc < 50:
        feedback.append("El texto es muy corto. Intentá desarrollar más tus ideas.")
    elif wc < 100:
        feedback.append("El texto es breve. Podrías profundizar en algunos puntos.")

    avg_sent = metrics.get("sentence_length", {}).get("avg", 0)
    if avg_sent > 30:
        feedback.append("Tus oraciones son muy largas. Intentá dividirlas para mejorar la claridad.")
    elif avg_sent < 8 and wc > 50:
        feedback.append("Tus oraciones son muy cortas. Podrías combinarlas para dar fluidez al texto.")

    vocab = metrics.get("vocabulary_richness", 0)
    if vocab < 0.4 and wc > 50:
        feedback.append("Repetís muchas palabras. Intentá usar sinónimos para enriquecer el texto.")
    elif vocab > 0.7:
        feedback.append("Buena variedad de vocabulario.")

    filler = metrics.get("filler_words", 0)
    if filler > 0:
        feedback.append(f"Encontramos {filler} palabra(s) de relleno. Evitá 'este', 'eh', 'o sea' para mayor claridad.")

    pars = metrics.get("paragraphs", {})
    if pars.get("paragraph_count", 0) <= 1 and wc > 100:
        feedback.append("El texto no tiene párrafos. Separar en párrafos mejora la lectura.")
    if not pars.get("has_conclusion") and wc > 100:
        feedback.append("Podrías agregar una conclusión para cerrar tus ideas.")

    readability = metrics.get("readability", {})
    label = readability.get("label", "")
    if label in ("Bastante difícil", "Muy difícil"):
        feedback.append("El texto es difícil de leer. Intentá usar oraciones más cortas y vocabulario más simple.")
    elif label in ("Bastante fácil", "Muy fácil") and wc > 50:
        feedback.append("El texto es muy fácil de leer. Podrías incorporar vocabulario más específico.")

    for req in metrics.get("requirements", []):
        if not req.get("matched"):
            feedback.append(f'Requisito no cumplido: "{req.get("requirement")}" — intentá incluir este tema en tu texto.')
        else:
            feedback.append(f'Requisito cumplido: "{req.get("requirement")}"')

    return feedback


def calculate_score(metrics: dict) -> int:
    score = 0

    wc = metrics.get("word_count", 0)
    if wc >= 100:
        score += 20
    elif wc >= 50:
        score += 10

    vocab = metrics.get("vocabulary_richness", 0)
    if vocab >= 0.6:
        score += 20
    elif vocab >= 0.4:
        score += 10

    avg_sent = metrics.get("sentence_length", {}).get("avg", 0)
    if 10 <= avg_sent <= 25:
        score += 15

    pars = metrics.get("paragraphs", {})
    if pars.get("paragraph_count", 0) >= 3:
        score += 15
    elif pars.get("paragraph_count", 0) >= 2:
        score += 10

    filler = metrics.get("filler_words", 0)
    if filler == 0:
        score += 10
    elif filler <= 2:
        score += 5

    readability = metrics.get("readability", {})
    label = readability.get("label", "")
    if label == "Normal":
        score += 10
    elif label in ("Bastante fácil", "Bastante difícil"):
        score += 5

    reqs = metrics.get("requirements", [])
    req_matched = sum(1 for r in reqs if r.get("matched"))
    req_total = len(reqs)
    if req_total > 0:
        score += int((req_matched / req_total) * 10)

    return min(score, 100)
