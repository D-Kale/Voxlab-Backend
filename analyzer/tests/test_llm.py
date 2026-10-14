"""Tests del coaching con LLM sin llamar a la red."""

import json

import pytest

from analyzers import llm


@pytest.fixture(autouse=True)
def _reset(monkeypatch):
    llm._cache.clear()
    monkeypatch.setattr(llm, "_budget_day", None)
    monkeypatch.setattr(llm, "_budget_used", 0)


def _fake_openrouter(monkeypatch, content, calls):
    class _Response:
        def __init__(self, body):
            self._body = body

        def read(self):
            return self._body

        def __enter__(self):
            return self

        def __exit__(self, *exc):
            return False

    def urlopen(request, timeout):
        calls.append(json.loads(request.data))
        body = {"model": "google/gemma-4-31b-it:free",
                "choices": [{"message": {"content": content}}]}
        return _Response(json.dumps(body).encode())

    monkeypatch.setattr(llm.urllib.request, "urlopen", urlopen)


LONG_TEXT = " ".join(["La comunicación clara ayuda a convencer al público."] * 4)


def test_disabled_without_key(monkeypatch):
    monkeypatch.delenv("OPENROUTER_API_KEY", raising=False)
    assert llm.review_text(LONG_TEXT, [], 50) is None
    assert llm.review_speech(LONG_TEXT, None, {}) is None


def test_review_text_only_asks_ambiguous_requirements(monkeypatch):
    monkeypatch.setenv("OPENROUTER_API_KEY", "sk-test")
    calls = []
    _fake_openrouter(
        monkeypatch,
        '```json\n{"requirements": {"hablar del público": true}, '
        '"feedback": "Sumá un ejemplo concreto."}\n```',
        calls,
    )
    reqs = [
        {"requirement": "hablar del público", "score": 0.4},
        {"requirement": "usar conectores", "score": 0.9},
    ]
    review = llm.review_text(LONG_TEXT, reqs, 60)
    assert review == {
        "feedback": "Sumá un ejemplo concreto.",
        "requirements": {"hablar del público": True},
        "model": "google/gemma-4-31b-it:free",
    }
    prompt = calls[0]["messages"][1]["content"]
    assert "hablar del público" in prompt
    assert "usar conectores" not in prompt
    assert calls[0]["models"] == llm.models()


def test_review_text_skips_short_or_excellent_text(monkeypatch):
    monkeypatch.setenv("OPENROUTER_API_KEY", "sk-test")
    calls = []
    _fake_openrouter(monkeypatch, '{"feedback": "x"}', calls)
    assert llm.review_text("muy corto", [], 10) is None
    assert llm.review_text(LONG_TEXT, [{"requirement": "a", "score": 0.9}], 92) is None
    assert calls == []


def test_cache_and_daily_budget(monkeypatch):
    monkeypatch.setenv("OPENROUTER_API_KEY", "sk-test")
    monkeypatch.setenv("LLM_DAILY_LIMIT", "1")
    calls = []
    _fake_openrouter(monkeypatch, '{"feedback": "Ordená tus ideas."}', calls)
    transcript = " ".join(["hoy les quiero contar algo"] * 6)
    first = llm.review_speech(transcript, "Presentación", {})
    again = llm.review_speech(transcript, "Presentación", {})
    other = llm.review_speech(transcript, "Otro tema", {})
    assert first == again
    assert other is None  # presupuesto agotado
    assert len(calls) == 1


def test_network_error_returns_none(monkeypatch):
    monkeypatch.setenv("OPENROUTER_API_KEY", "sk-test")

    def boom(request, timeout):
        raise llm.urllib.error.URLError("down")

    monkeypatch.setattr(llm.urllib.request, "urlopen", boom)
    assert llm.review_speech(" ".join(["palabra"] * 30), None, {}) is None


def test_invalid_json_returns_none(monkeypatch):
    monkeypatch.setenv("OPENROUTER_API_KEY", "sk-test")
    _fake_openrouter(monkeypatch, "Claro, acá va mi opinión", [])
    assert llm.review_speech(" ".join(["palabra"] * 30), None, {}) is None
