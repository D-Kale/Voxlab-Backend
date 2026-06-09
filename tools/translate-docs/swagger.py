import json
import sys


def walk_fields(doc, fn):
    def collect(path, obj, key):
        if isinstance(obj, dict) and key in obj and isinstance(obj[key], str):
            val = obj[key]
            if val:
                fn(path, val, key)

    info = doc.get("info")
    if isinstance(info, dict):
        collect("info.title", info, "title")
        collect("info.description", info, "description")
        collect("info.termsOfService", info, "termsOfService")
        contact = info.get("contact")
        if isinstance(contact, dict):
            collect("info.contact.name", contact, "name")
        license_ = info.get("license")
        if isinstance(license_, dict):
            collect("info.license.name", license_, "name")

    top_tags = doc.get("tags")
    if isinstance(top_tags, list):
        for i, tag_obj in enumerate(top_tags):
            if isinstance(tag_obj, dict):
                collect(f"tags[{i}].name", tag_obj, "name")
                collect(f"tags[{i}].description", tag_obj, "description")

    paths = doc.get("paths")
    if isinstance(paths, dict):
        for path_key, path_obj in paths.items():
            if not isinstance(path_obj, dict):
                continue
            for method, method_obj in path_obj.items():
                if not isinstance(method_obj, dict):
                    continue
                base = f"paths.{path_key}.{method}"
                collect(f"{base}.summary", method_obj, "summary")
                collect(f"{base}.description", method_obj, "description")

                tags = method_obj.get("tags")
                if isinstance(tags, list):
                    for i, tag in enumerate(tags):
                        if isinstance(tag, str) and tag:
                            fn(f"{base}.tags[{i}]", tag, "tag")

                params = method_obj.get("parameters")
                if isinstance(params, list):
                    for i, param in enumerate(params):
                        if isinstance(param, dict):
                            collect(f"{base}.parameters[{i}].description", param, "description")

                responses = method_obj.get("responses")
                if isinstance(responses, dict):
                    for code, resp_obj in responses.items():
                        if isinstance(resp_obj, dict):
                            collect(f"{base}.responses.{code}.description", resp_obj, "description")

    defs = doc.get("definitions")
    if isinstance(defs, dict):
        for def_key, def_obj in defs.items():
            if isinstance(def_obj, dict):
                collect(f"definitions.{def_key}.description", def_obj, "description")

    sec_defs = doc.get("securityDefinitions")
    if isinstance(sec_defs, dict):
        for sec_key, sec_obj in sec_defs.items():
            if isinstance(sec_obj, dict):
                collect(f"securityDefinitions.{sec_key}.description", sec_obj, "description")


def extract_items(doc):
    items = []
    walk_fields(doc, lambda path, val, key: items.append({
        "path": path, "original": val, "context": key
    }))
    return items


def load_existing(doc):
    result = {}
    walk_fields(doc, lambda path, val, key: result.setdefault(path, val))
    return result


def apply_translations(doc, translations):
    result = json.loads(json.dumps(doc))

    def apply_str(path, obj, key):
        if path in translations:
            obj[key] = translations[path]

    info = result.get("info")
    if isinstance(info, dict):
        apply_str("info.title", info, "title")
        apply_str("info.description", info, "description")
        apply_str("info.termsOfService", info, "termsOfService")
        contact = info.get("contact")
        if isinstance(contact, dict):
            apply_str("info.contact.name", contact, "name")
        license_ = info.get("license")
        if isinstance(license_, dict):
            apply_str("info.license.name", license_, "name")

    top_tags = result.get("tags")
    if isinstance(top_tags, list):
        for i, tag_obj in enumerate(top_tags):
            if isinstance(tag_obj, dict):
                apply_str(f"tags[{i}].name", tag_obj, "name")
                apply_str(f"tags[{i}].description", tag_obj, "description")

    paths = result.get("paths")
    if isinstance(paths, dict):
        for path_key, path_obj in paths.items():
            if not isinstance(path_obj, dict):
                continue
            for method, method_obj in path_obj.items():
                if not isinstance(method_obj, dict):
                    continue
                base = f"paths.{path_key}.{method}"
                apply_str(f"{base}.summary", method_obj, "summary")
                apply_str(f"{base}.description", method_obj, "description")

                tags = method_obj.get("tags")
                if isinstance(tags, list):
                    for i in range(len(tags)):
                        p = f"{base}.tags[{i}]"
                        if p in translations:
                            tags[i] = translations[p]

                params = method_obj.get("parameters")
                if isinstance(params, list):
                    for i, param in enumerate(params):
                        if isinstance(param, dict):
                            apply_str(f"{base}.parameters[{i}].description", param, "description")

                responses = method_obj.get("responses")
                if isinstance(responses, dict):
                    for code, resp_obj in responses.items():
                        if isinstance(resp_obj, dict):
                            apply_str(f"{base}.responses.{code}.description", resp_obj, "description")

    defs = result.get("definitions")
    if isinstance(defs, dict):
        for def_key, def_obj in defs.items():
            if isinstance(def_obj, dict):
                apply_str(f"definitions.{def_key}.description", def_obj, "description")

    sec_defs = result.get("securityDefinitions")
    if isinstance(sec_defs, dict):
        for sec_key, sec_obj in sec_defs.items():
            if isinstance(sec_obj, dict):
                apply_str(f"securityDefinitions.{sec_key}.description", sec_obj, "description")

    return result


def validate_translation(data):
    try:
        doc = json.loads(data)
    except json.JSONDecodeError as e:
        print(f"[ERR] Invalid JSON: {e}", file=sys.stderr)
        return False

    if doc.get("swagger") != "2.0":
        print(f'[ERR] Missing or invalid swagger version: {doc.get("swagger")!r}', file=sys.stderr)
        return False

    info = doc.get("info")
    if not isinstance(info, dict) or not info.get("title") or not info.get("version"):
        print("[ERR] Missing info.title or info.version", file=sys.stderr)
        return False

    if "paths" not in doc or not isinstance(doc["paths"], dict):
        print("[ERR] Missing paths object", file=sys.stderr)
        return False

    return True
