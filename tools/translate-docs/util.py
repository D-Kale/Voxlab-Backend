import os
import subprocess
import sys

_SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
PROJECT_ROOT = os.path.normpath(os.path.join(_SCRIPT_DIR, "..", ".."))

SWAGGER_PATH = os.path.join(PROJECT_ROOT, "docs", "swagger.json")
EXISTING_PATH = os.path.join(PROJECT_ROOT, "docs", "es", "openapi-es.json")
TMP_PATH = os.path.join(PROJECT_ROOT, "docs", "es", "tmp_swagger_es.json")

NVIDIA_API_URL = "https://integrate.api.nvidia.com/v1/chat/completions"


def load_env():
    candidates = [
        os.path.join(PROJECT_ROOT, ".env"),
        os.path.join(PROJECT_ROOT, ".env.example"),
        os.path.join(_SCRIPT_DIR, ".env"),
    ]
    for path in candidates:
        try:
            with open(path) as f:
                for line in f:
                    line = line.strip()
                    if not line or line.startswith("#") or "=" not in line:
                        continue
                    key, _, val = line.partition("=")
                    key, val = key.strip(), val.strip()
                    if key not in os.environ:
                        os.environ[key] = val
        except FileNotFoundError:
            continue


def merge_translations(existing, api):
    merged = dict(existing)
    merged.update({k: v for k, v in api.items() if k not in merged})
    return merged


def show_diff():
    rel_path = os.path.relpath(EXISTING_PATH, PROJECT_ROOT)
    print(f"\n--- Git diff of {rel_path} ---")
    result = subprocess.run(
        ["git", "diff", rel_path],
        capture_output=True, text=True, cwd=PROJECT_ROOT,
    )
    if result.stdout:
        print(result.stdout)
    if result.stderr:
        print(result.stderr, file=sys.stderr)
    if result.returncode != 0:
        print("[WARN] git diff failed", file=sys.stderr)
