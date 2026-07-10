FROM python:3.12-slim AS builder

WORKDIR /app

ENV HF_HOME=/app/.cache/huggingface
ENV UV_COMPILE_BYTECODE=1 UV_LINK_MODE=copy

COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

COPY pyproject.toml uv.lock .python-version ./
RUN uv sync --frozen --no-dev --no-install-project

COPY . .
RUN uv sync --frozen --no-dev
RUN uv run download-model

RUN find /app -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null; \
    rm -rf /root/.cache/uv /root/.cache/pip /tmp/*

FROM python:3.12-slim

WORKDIR /app

ENV PATH="/app/.venv/bin:$PATH"
ENV HF_HOME=/app/.cache/huggingface
ENV PYTHONDONTWRITEBYTECODE=1

COPY --from=builder /app/.venv /app/.venv
COPY --from=builder /app/analyzers /app/analyzers
COPY --from=builder /app/main.py /app/main.py
COPY --from=builder /app/.cache/huggingface /app/.cache/huggingface

RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8000

CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
