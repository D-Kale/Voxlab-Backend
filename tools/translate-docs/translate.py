import json
import os

import requests

from util import NVIDIA_API_URL

DEFAULT_MODEL = "moonshotai/kimi-k2.6"


def build_prompt(items):
    prompt = """You are a professional translator specializing in API documentation in Spanish.
Translate the following English strings from a Swagger/OpenAPI specification to natural Latin American Spanish.

RULES:
- Keep standard technical terms in English (API, token, JWT, UUID, JSON, endpoint, HTTP, REST, JSONB)
- Do NOT translate: field names, code examples, URLs, markdown code blocks, JSON schema values
- Use natural, friendly Spanish suitable for Spanish-speaking developers
- For tags (used as category/section names), provide short Spanish equivalents
- For descriptions (which may contain markdown and code blocks), translate the surrounding text but preserve any markdown/code formatting exactly
- Ensure consistency: the same English string should always translate the same way

Return ONLY a JSON object where:
- Keys are exactly the "path" values from each item (unchanged)
- Values are the Spanish translations

Do NOT include any explanation, commentary, or markdown formatting around the JSON.
The response must be parseable as raw JSON.

Items to translate:
"""
    prompt += json.dumps(items, ensure_ascii=False, indent=2)
    return prompt


def call_nvidia(items, timeout, verbose):
    api_key = os.environ.get("NVIDIA_API_KEY")
    if not api_key:
        raise RuntimeError("NVIDIA_API_KEY not set")

    prompt = build_prompt(items)

    if verbose:
        print(f"[INFO] Sending {len(items)} strings to NVIDIA API...")

    headers = {
        "Authorization": f"Bearer {api_key}",
        "Accept": "application/json",
    }

    payload = {
        "model": DEFAULT_MODEL,
        "messages": [
            {"role": "system", "content": "You are a precise JSON generator. Always respond with valid JSON only, no other text."},
            {"role": "user", "content": prompt},
        ],
        "max_tokens": 16384,
        "temperature": 0.1,
    }

    resp = requests.post(NVIDIA_API_URL, headers=headers, json=payload, timeout=timeout)
    resp.raise_for_status()
    result = resp.json()

    if "error" in result:
        raise RuntimeError(f"API error: {result['error']}")

    choices = result.get("choices", [])
    if not choices:
        raise RuntimeError("API returned no choices")

    content = choices[0]["message"]["content"].strip()
    if content.startswith("```"):
        lines = content.split("\n")
        cleaned = []
        for i, line in enumerate(lines):
            if i == 0 and line.startswith("```"):
                continue
            if i == len(lines) - 1 and line.startswith("```"):
                continue
            cleaned.append(line)
        content = "\n".join(cleaned).strip()

    try:
        translations = json.loads(content)
    except json.JSONDecodeError as e:
        raise RuntimeError(f"Failed to parse translation JSON: {e}\nRaw:\n{content}")

    if verbose:
        print(f"[INFO] Received {len(translations)} translations from API")

    return translations
