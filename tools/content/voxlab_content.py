#!/usr/bin/env python3
"""Cliente de contenido de Voxlab para agentes y scripts (solo stdlib).

Lee y publica el contenido del curso (tracks, módulos, lecciones,
ejercicios) a través de la API. Ver docs/md/CONTENT_AGENT.md.

Variables de entorno:
    VOXLAB_API_URL   (default https://tvoxlab.tech/api/v1)
    VOXLAB_EMAIL     cuenta con role=admin
    VOXLAB_PASSWORD

Comandos:
    tree                         Árbol completo: tracks > módulos > lecciones > ejercicios
    validate PLAN.json           Revisa un plan sin tocar la API
    apply PLAN.json [--dry-run]  Crea lo que falte del plan (idempotente por título/nombre)
    publish TRACK_ID             is_active=true (la app lo muestra)
    unpublish TRACK_ID           is_active=false (borrador, la app lo oculta)
"""

from __future__ import annotations

import json
import os
import sys
import urllib.error
import urllib.request

API = os.getenv("VOXLAB_API_URL", "https://tvoxlab.tech/api/v1").rstrip("/")

# Tipos que la app móvil sabe mostrar. audio y video existen en la API
# pero la app muestra "no disponible".
APP_TYPES = {"quiz", "reading", "writing", "oratory_minigame"}


class ApiError(Exception):
    pass


class Client:
    def __init__(self, dry_run: bool = False):
        self.dry_run = dry_run
        self.token: str | None = None
        self._fake_id = 0

    def login(self) -> None:
        email, password = os.getenv("VOXLAB_EMAIL"), os.getenv("VOXLAB_PASSWORD")
        if not email or not password:
            raise ApiError("Definí VOXLAB_EMAIL y VOXLAB_PASSWORD (cuenta admin).")
        data = self._send("POST", "/auth/login", {"email": email, "password": password}, auth=False)
        self.token = data.get("token") or data.get("access_token")
        user = data.get("user") or {}
        if user.get("role") != "admin":
            raise ApiError(f"La cuenta {email} no es admin; la API rechazará las escrituras.")

    def get(self, path: str):
        return self._send("GET", path, None, auth=False)

    def write(self, method: str, path: str, body: dict):
        if self.dry_run:
            print(f"[dry-run] {method} {path} {json.dumps(body, ensure_ascii=False)[:200]}")
            self._fake_id += 1
            return {"id": f"dry-{self._fake_id}"}
        return self._send(method, path, body, auth=True)

    def _send(self, method: str, path: str, body, auth: bool):
        headers = {"Content-Type": "application/json", "Accept": "application/json"}
        if auth:
            headers["Authorization"] = f"Bearer {self.token}"
        req = urllib.request.Request(
            API + path,
            data=None if body is None else json.dumps(body).encode(),
            headers=headers,
            method=method,
        )
        try:
            with urllib.request.urlopen(req, timeout=30) as resp:
                payload = json.loads(resp.read() or b"{}")
        except urllib.error.HTTPError as e:
            raise ApiError(f"{method} {path} -> {e.code}: {e.read().decode(errors='replace')[:300]}")
        if isinstance(payload, dict) and "data" in payload:
            return payload["data"]
        return payload


# ── Lectura ───────────────────────────────────────────────────────────────

def tree(client: Client) -> None:
    for t in client.get("/tracks") or []:
        state = "" if t.get("is_active", True) else "  [BORRADOR]"
        print(f"TRACK {t['id']} · {t['title']} (order {t.get('order_index')}){state}")
        for m in t.get("modules") or []:
            print(f"  MÓDULO {m['id']} · {m['title']} (order {m.get('order_index')})")
            for ml in m.get("lessons") or []:
                lesson = ml.get("lesson") or {}
                print(f"    LECCIÓN {lesson.get('id')} · {lesson.get('title')} (order {ml.get('order_index')})")
                for le in lesson.get("exercises") or []:
                    ex = le.get("exercise") or {}
                    print(f"      {ex.get('type'):<17} {ex.get('id')} · {ex.get('name')} (order {le.get('order_index')})")


# ── Validación ────────────────────────────────────────────────────────────

def _need(cond: bool, errors: list, msg: str) -> None:
    if not cond:
        errors.append(msg)


def validate_exercise(ex: dict, where: str) -> list:
    errors: list = []
    t, c, name = ex.get("type"), ex.get("content") or {}, ex.get("name")
    _need(bool(name), errors, f"{where}: falta name")
    _need(t in APP_TYPES, errors, f"{where}: type '{t}' no lo muestra la app ({sorted(APP_TYPES)})")
    if t == "quiz":
        qs = c.get("questions") or []
        _need(len(qs) >= 1, errors, f"{where}: quiz sin preguntas")
        _need(isinstance(c.get("points_per_question"), int) and c["points_per_question"] > 0,
              errors, f"{where}: points_per_question entero > 0")
        for i, q in enumerate(qs):
            opts = q.get("options") or []
            _need(bool(q.get("question")), errors, f"{where}: pregunta {i} vacía")
            _need(2 <= len(opts) <= 4, errors, f"{where}: pregunta {i} necesita 2 a 4 opciones")
            ci = q.get("correct_index")
            _need(isinstance(ci, int) and 0 <= ci < len(opts), errors,
                  f"{where}: pregunta {i} correct_index fuera de rango (0-based)")
    elif t == "reading":
        _need(bool(c.get("title")) and bool(c.get("content")), errors, f"{where}: reading necesita title y content")
        _need(isinstance(c.get("reading_time_seconds"), int), errors, f"{where}: reading_time_seconds entero")
        _need(isinstance(c.get("points"), int), errors, f"{where}: points entero")
    elif t == "writing":
        mn, mx = c.get("min_words"), c.get("max_words")
        _need(bool(c.get("prompt")), errors, f"{where}: writing necesita prompt")
        _need(isinstance(mn, int) and isinstance(mx, int) and 0 < mn < mx, errors,
              f"{where}: 0 < min_words < max_words")
        _need(isinstance(c.get("requirements"), list), errors, f"{where}: requirements lista (puede ser vacía)")
        _need(isinstance(c.get("points"), int), errors, f"{where}: points entero")
    elif t == "oratory_minigame":
        d, md = c.get("duration_seconds"), c.get("min_duration_seconds")
        _need(bool(c.get("prompt")) and bool(c.get("topic")), errors, f"{where}: oratoria necesita prompt y topic")
        _need(isinstance(d, int) and isinstance(md, int) and 0 < md <= d <= 180, errors,
              f"{where}: 0 < min_duration_seconds <= duration_seconds <= 180")
        _need(isinstance(c.get("requirements"), list), errors, f"{where}: requirements lista")
        _need(isinstance(c.get("points"), int), errors, f"{where}: points entero")
    return errors


def validate_plan(plan: dict) -> list:
    errors: list = []
    track = plan.get("track") or {}
    _need(bool(track.get("title")) or bool(track.get("id")), errors, "track: falta title o id")
    for mi, m in enumerate(plan.get("modules") or []):
        _need(bool(m.get("title")), errors, f"módulo {mi}: falta title")
        for li, lesson in enumerate(m.get("lessons") or []):
            where = f"módulo {mi} lección {li}"
            _need(bool(lesson.get("title")), errors, f"{where}: falta title")
            _need(isinstance(lesson.get("estimated_time_seconds"), int), errors,
                  f"{where}: estimated_time_seconds entero")
            exs = lesson.get("exercises") or []
            _need(len(exs) >= 1, errors, f"{where}: sin ejercicios")
            for ei, ex in enumerate(exs):
                errors += validate_exercise(ex, f"{where} ejercicio {ei}")
    return errors


# ── Aplicar un plan ───────────────────────────────────────────────────────

def apply_plan(client: Client, plan: dict) -> None:
    errors = validate_plan(plan)
    if errors:
        raise ApiError("Plan inválido:\n  - " + "\n  - ".join(errors))

    tracks = client.get("/tracks") or []
    spec = plan["track"]
    track = next((t for t in tracks if t["id"] == spec.get("id") or t["title"] == spec.get("title")), None)
    if track is None:
        track = client.write("POST", "/tracks", {
            "title": spec["title"],
            "description": spec.get("description", ""),
            "icon_url": spec.get("icon_url", ""),
        })
        print(f"+ track {track['id']} {spec['title']}")
        # Los tracks nuevos nacen como borrador; `publish` los muestra.
        client.write("PUT", f"/tracks/{track['id']}", {"is_active": False})
        track.setdefault("modules", [])
    else:
        print(f"= track {track['id']} {track['title']}")

    existing_lessons = {l["title"]: l for l in (client.get("/lessons") or [])}
    existing_exercises = {e["name"]: e for e in (client.get("/exercises") or [])}
    modules = {m["title"]: m for m in track.get("modules") or []}
    next_order = max([m.get("order_index", 0) for m in modules.values()] + [0]) + 1

    for m_spec in plan.get("modules") or []:
        module = modules.get(m_spec["title"])
        if module is None:
            module = client.write("POST", "/modules", {
                "track_id": track["id"],
                "title": m_spec["title"],
                "description": m_spec.get("description", ""),
                "order_index": next_order,
            })
            module["lessons"] = []
            next_order += 1
            print(f"  + módulo {module['id']} {m_spec['title']}")
        else:
            print(f"  = módulo {module['id']} {module['title']}")
        linked = {(ml.get("lesson") or {}).get("title") for ml in module.get("lessons") or []}

        lesson_order = []
        for l_spec in m_spec.get("lessons") or []:
            lesson = existing_lessons.get(l_spec["title"])
            if lesson is None:
                lesson = client.write("POST", "/lessons", {
                    "title": l_spec["title"],
                    "description": l_spec.get("description", ""),
                    "estimated_time_seconds": l_spec["estimated_time_seconds"],
                })
                existing_lessons[l_spec["title"]] = lesson
                print(f"    + lección {lesson['id']} {l_spec['title']}")
            if l_spec["title"] not in linked:
                client.write("POST", f"/modules/{module['id']}/lessons", {"lesson_id": lesson["id"]})
            lesson_order.append(lesson["id"])

            linked_ex = set()
            if not client.dry_run and not str(lesson["id"]).startswith("dry-"):
                for le in client.get(f"/lessons/{lesson['id']}/exercises") or []:
                    ex = le.get("exercise") or le
                    linked_ex.add(ex.get("name"))
            ex_order = []
            for e_spec in l_spec["exercises"]:
                ex = existing_exercises.get(e_spec["name"])
                if ex is None:
                    ex = client.write("POST", "/exercises", {
                        "name": e_spec["name"],
                        "type": e_spec["type"],
                        "content": e_spec["content"],
                    })
                    existing_exercises[e_spec["name"]] = ex
                    print(f"      + {e_spec['type']} {ex['id']} {e_spec['name']}")
                if e_spec["name"] not in linked_ex:
                    client.write("POST", f"/lessons/{lesson['id']}/exercises", {"exercise_id": ex["id"]})
                ex_order.append(ex["id"])
            client.write("PUT", f"/lessons/{lesson['id']}/exercises/reorder", {
                "items": [{"exercise_id": i, "order_index": n} for n, i in enumerate(ex_order, 1)],
            })
        # Vincular una lección le pone order_index 0: siempre reordenar.
        client.write("PUT", f"/modules/{module['id']}/lessons/reorder", {
            "items": [{"lesson_id": i, "order_index": n} for n, i in enumerate(lesson_order, 1)],
        })
    print(f"Listo. Revisá con `tree` y publicá con `publish {track['id']}`.")


def main(argv: list) -> int:
    if len(argv) < 2 or argv[1] in {"-h", "--help"}:
        print(__doc__)
        return 0
    cmd, args = argv[1], argv[2:]
    dry = "--dry-run" in args
    args = [a for a in args if a != "--dry-run"]
    client = Client(dry_run=dry)
    try:
        if cmd == "tree":
            tree(client)
        elif cmd == "validate":
            errors = validate_plan(json.load(open(args[0], encoding="utf-8")))
            print("Plan válido." if not errors else "\n".join(errors))
            return 1 if errors else 0
        elif cmd == "apply":
            plan = json.load(open(args[0], encoding="utf-8"))
            if not dry:
                client.login()
            apply_plan(client, plan)
        elif cmd in {"publish", "unpublish"}:
            client.login()
            client.write("PUT", f"/tracks/{int(args[0])}", {"is_active": cmd == "publish"})
            print(f"Track {args[0]} {'publicado' if cmd == 'publish' else 'oculto'}.")
        else:
            print(f"Comando desconocido: {cmd}\n{__doc__}")
            return 2
    except ApiError as e:
        print(f"ERROR: {e}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
