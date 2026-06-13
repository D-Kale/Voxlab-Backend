from fastapi import FastAPI
from pydantic import BaseModel
from typing import List, Optional

from analyzers.text import analyze_text
from analyzers.scoring import generate_feedback, calculate_score

app = FastAPI(title="Voxlab Analyzer", version="0.1.0")


class AnalyzeRequest(BaseModel):
    text: str
    requirements: Optional[List[str]] = None


class AnalyzeResponse(BaseModel):
    word_count: int
    sentence_count: int
    sentence_length: dict
    vocabulary_richness: float
    paragraphs: dict
    readability: dict
    filler_words: int
    keywords: List[str]
    requirements: List[dict]
    score: int
    feedback: List[str]


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/analyze/text", response_model=AnalyzeResponse)
def analyze(req: AnalyzeRequest):
    metrics = analyze_text(req.text, req.requirements or [])
    score = calculate_score(metrics)
    feedback = generate_feedback(metrics, req.requirements)
    return AnalyzeResponse(**metrics, score=score, feedback=feedback)
