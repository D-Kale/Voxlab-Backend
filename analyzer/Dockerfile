FROM python:3.12-slim

WORKDIR /app

RUN pip install uv --no-cache-dir

COPY pyproject.toml .python-version ./
RUN uv sync --no-dev

COPY . .

RUN uv run python -m spacy download es_core_news_md

EXPOSE 8000

CMD ["uv", "run", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
