"""
Descarga los modelos necesarios para el analyzer.

Uso:
    uv run download-model

Descarga:
    - spaCy: es_core_news_md (modelo de lenguaje español)
    - fastembed: paraphrase-multilingual-MiniLM-L12-v2 (embeddings ONNX)
"""

import subprocess
import sys


def download_spacy_model():
    print("Descargando modelo spaCy es_core_news_md...")
    subprocess.run(
        [sys.executable, "-m", "spacy", "download", "es_core_news_md"],
        check=True,
    )
    print("spaCy: OK")


def download_embedding_model():
    print("Descargando modelo de embeddings MiniLM multilingüe (ONNX)...")
    from fastembed import TextEmbedding
    TextEmbedding(
        model_name="sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2",
    )
    print("Embeddings: OK")


def main():
    download_spacy_model()
    download_embedding_model()
    print("Todos los modelos descargados correctamente.")


if __name__ == "__main__":
    main()
