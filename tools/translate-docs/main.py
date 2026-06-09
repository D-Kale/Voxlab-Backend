import argparse
import json
import os
import shutil
import sys

from swagger import (
    extract_items,
    load_existing,
    apply_translations,
    validate_translation,
)
from translate import call_nvidia
from util import (
    SWAGGER_PATH,
    EXISTING_PATH,
    TMP_PATH,
    load_env,
    merge_translations,
    show_diff,
)


def main():
    parser = argparse.ArgumentParser(description="Translate Swagger spec to Spanish via NVIDIA API")
    parser.add_argument("--show-diff", action="store_true", default=False)
    parser.add_argument("--force", action="store_true", default=False)
    parser.add_argument("--validate-only", action="store_true", default=False)
    parser.add_argument("--verbose", action="store_true", default=False)
    parser.add_argument("--timeout", type=int, default=300)
    args = parser.parse_args()

    load_env()

    if args.verbose:
        print("[INFO] Translation script starting")

    if args.validate_only:
        if args.verbose:
            print("[INFO] Validating existing translation...")
        try:
            with open(EXISTING_PATH) as f:
                data = f.read()
        except FileNotFoundError:
            print(f"[ERR] Cannot read {EXISTING_PATH}", file=sys.stderr)
            sys.exit(1)
        if validate_translation(data):
            print(f"[OK] {EXISTING_PATH} is valid")
            return
        sys.exit(1)

    try:
        with open(SWAGGER_PATH) as f:
            doc = json.load(f)
    except (FileNotFoundError, json.JSONDecodeError) as e:
        print(f"[ERR] Cannot read {SWAGGER_PATH}: {e}", file=sys.stderr)
        sys.exit(1)

    items = extract_items(doc)
    if args.verbose:
        print(f"[INFO] Found {len(items)} translatable strings")

    if not items:
        print("[WARN] No translatable strings found")
        return

    existing = {}
    has_existing = False
    try:
        with open(EXISTING_PATH) as f:
            existing_doc = json.load(f)
        existing = load_existing(existing_doc)
        has_existing = bool(existing)
        if args.verbose:
            print(f"[INFO] Found {len(existing)} existing translations")
    except (FileNotFoundError, json.JSONDecodeError):
        pass

    to_translate = []
    keep = {}
    for item in items:
        path = item["path"]
        if path in existing and not args.force:
            keep[path] = existing[path]
        else:
            to_translate.append(item)

    if not to_translate and has_existing:
        if args.verbose:
            print("[INFO] All strings already translated, nothing to do")
        print("[OK] No new strings to translate (use --force to re-translate all)")
        return

    api_translations = {}
    if to_translate:
        try:
            api_translations = call_nvidia(to_translate, args.timeout, args.verbose)
        except Exception as e:
            if has_existing:
                print(f"[ERR] Translation failed: {e}", file=sys.stderr)
                print("[INFO] Keeping existing translation")
                return
            print(f"[ERR] Translation failed and no fallback available: {e}", file=sys.stderr)
            sys.exit(1)

    merged = merge_translations(keep, api_translations)
    if args.verbose:
        print(f"[INFO] Merged {len(keep)} kept + {len(api_translations)} new = {len(merged)} translations")

    translated = apply_translations(doc, merged)

    os.makedirs(os.path.dirname(TMP_PATH), exist_ok=True)
    tmp_data = json.dumps(translated, indent=4, ensure_ascii=False)

    with open(TMP_PATH, "w") as f:
        f.write(tmp_data)

    if not validate_translation(tmp_data):
        os.remove(TMP_PATH)
        sys.exit(1)

    shutil.move(TMP_PATH, EXISTING_PATH)
    print(f"[OK] Translation written to {EXISTING_PATH}")
    print(f"[OK] You can review the changes at {EXISTING_PATH}")

    if args.show_diff:
        show_diff()


if __name__ == "__main__":
    main()
