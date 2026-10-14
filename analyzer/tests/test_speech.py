"""Tests de las métricas de oratoria (sin cargar el modelo de Whisper)."""

from analyzers.speech import (
    WHISPER_PROMPT,
    analyze_words,
    count_fillers,
    drop_prompt_echo,
    words_from,
)


def steady_speech(n_words, start=0.5, word=0.3, gap=0.1, pause_every=0, pause=1.0):
    """Palabras regulares; cada `pause_every` palabras mete una pausa."""
    items, t = [], start
    for i in range(n_words):
        items.append((f"palabra{i}", t, t + word))
        t += word + gap
        if pause_every and (i + 1) % pause_every == 0:
            t += pause
    return words_from(items)


def test_no_speech_scores_zero():
    result = analyze_words([], duration=30)
    assert result["score"] == 0
    assert result["word_count"] == 0
    assert "No detectamos tu voz" in result["feedback"][0]


def test_hesitations_always_count():
    words = words_from([("Eh,", 0, 0.3), ("hola", 0.4, 0.7), ("mmm", 0.8, 1.2)])
    assert count_fillers(words) == {"eh": 1, "mmm": 1}


def test_contextual_filler_only_when_loose():
    demonstrative = words_from([("este", 0, 0.2), ("proyecto", 0.25, 0.7)])
    assert count_fillers(demonstrative) == {}
    filler = words_from([("este,", 0, 0.2), ("bueno", 0.3, 0.5)])
    assert count_fillers(filler)["este"] == 1


def test_multiword_filler():
    words = words_from([("o", 0, 0.1), ("sea,", 0.15, 0.4), ("todos", 0.5, 0.8)])
    assert count_fillers(words) == {"o sea": 1}


def test_good_delivery_scores_high():
    # ~150 palabras por minuto con una pausa cada 15 palabras (~9 por minuto).
    words = steady_speech(150, pause_every=15, pause=0.9)
    duration = words[-1].end + 0.5
    result = analyze_words(words, duration, min_seconds=30, target_seconds=90)
    breakdown = result["score_breakdown"]
    assert 110 <= result["words_per_minute"] <= 160
    assert breakdown["ritmo"]["score"] == 100
    assert breakdown["pausas"]["score"] == 100
    assert breakdown["muletillas"]["score"] == 100
    assert result["score"] >= 90


def test_long_pauses_and_slow_pace_are_penalized():
    words = steady_speech(20, word=0.4, gap=0.3, pause_every=4, pause=4.5)
    result = analyze_words(words, words[-1].end, min_seconds=10)
    assert result["long_pause_count"] >= 3
    assert result["score_breakdown"]["pausas"]["score"] < 100
    assert result["words_per_minute"] < 110
    assert result["score_breakdown"]["ritmo"]["score"] < 100


def test_short_take_penalizes_duration():
    words = steady_speech(10)
    result = analyze_words(words, duration=5, min_seconds=30)
    assert result["score_breakdown"]["duracion"]["score"] < 20


def test_prompt_echo_is_dropped():
    echo = words_from([(t, 0, 0.1) for t in WHISPER_PROMPT.split()])
    assert drop_prompt_echo(echo) == []
    real = words_from([("Hola", 0, 0.3), ("a", 0.4, 0.5), ("todos", 0.5, 0.9)])
    assert drop_prompt_echo(real) == real
