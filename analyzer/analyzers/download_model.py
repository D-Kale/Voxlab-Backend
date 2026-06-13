import subprocess
import sys


def main():
    subprocess.run(
        [sys.executable, "-m", "spacy", "download", "es_core_news_md"],
        check=True,
    )


if __name__ == "__main__":
    main()
