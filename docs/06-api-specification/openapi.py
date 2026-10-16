#!/usr/bin/env python3
"""
Convert OpenAPI YAML specs to JSON and Postman collections.

Reads each `<service>/<service>-openapi.yaml` and writes:
- `<service>/<service>-openapi.json` — the OpenAPI spec as JSON
- `<service>/<service>-openapi.postman_collection.json` — Postman Collection v2.1

Usage:
    uv run openapi.py
"""

import json
from pathlib import Path
from urllib.parse import urlparse

root = Path(__file__).parent


def load_yaml(path: Path) -> dict:
    import yaml

    with open(path) as f:
        return yaml.safe_load(f)


def write_json(path: Path, data: dict) -> None:
    with open(path, "w") as f:
        json.dump(data, f, indent=2)
        f.write("\n")


def generate_example(schema: dict, visited: set | None = None) -> object:
    if visited is None:
        visited = set()

    if schema is None:
        return None

    # Handle refs
    if "$ref" in schema:
        # Can't resolve refs fully here; return basic type
        ref = schema["$ref"]
        if ref not in visited:
            visited.add(ref)

    if "example" in schema:
        return schema["example"]

    schema_type = schema.get("type")

    if schema_type == "object" or "properties" in schema:
        result = {}
        for prop_name, prop_schema in schema.get("properties", {}).items():
            result[prop_name] = generate_example(prop_schema, visited)
        if not result and schema.get("additionalProperties") is True:
            return {}
        return result

    if schema_type == "array":
        items = schema.get("items", {})
        return [generate_example(items, visited)]

    if schema_type == "string":
        fmt = schema.get("format")
        if fmt == "uuid":
            return "e7f8a9b0-1c2d-4e5f-8a9b-0c1d2e3f4a5b"
        if fmt == "date-time":
            return "2026-01-01T00:00:00Z"
        if fmt == "date":
            return "2026-01-01"
        if fmt == "email":
            return "user@example.com"
        return schema.get("default", "string")

    if schema_type == "integer" or schema_type == "number":
        return schema.get("default", schema.get("minimum", 1000))

    if schema_type == "boolean":
        return schema.get("default", True)

    return None


def to_postman(spec: dict) -> dict:
    info = spec.get("info", {})
    paths = spec.get("paths", {})
    servers = spec.get("servers", [])
    base_url = servers[0]["url"] if servers else "http://localhost:8080"

    collection = {
        "info": {
            "name": info.get("title", "NEO Bank API"),
            "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
            "description": info.get("description", ""),
            "version": info.get("version", "1.0.0"),
        },
        "variable": [{"key": "baseUrl", "value": base_url}],
        "item": [],
    }

    for path, methods in paths.items():
        for method, detail in methods.items():
            method_upper = str(method).upper()
            path_vars = path.strip("/").split("/") if path.strip("/") else []
            url_obj = {
                "raw": "{{baseUrl}}" + path,
                "host": ["{{baseUrl}}"],
                "path": path_vars,
            }

            request_item = {
                "method": method_upper,
                "header": [],
                "url": url_obj,
                "description": detail.get("summary", ""),
            }

            # Path parameters - mark with colon for Postman
            for param in detail.get("parameters", []):
                if param.get("in") == "path":
                    name = param.get("name")
                    # Replace path segment with :name if exact match
                    url_obj["path"] = [
                        ":" + name if seg == "{" + name + "}" or seg == name else seg
                        for seg in path_vars
                    ]
                    # Add description if missing - but Postman uses path variables differently
                    # Better to handle common case
                    url_obj["variable"] = [
                        {
                            "key": name,
                            "value": str(generate_example({"type": "string", "format": "uuid"}, set())),
                        }
                    ]

            # Query parameters
            query_params = [
                p for p in detail.get("parameters", []) if p.get("in") == "query"
            ]
            if query_params:
                url_obj["query"] = []
                for p in query_params:
                    url_obj["query"].append(
                        {
                            "key": p.get("name"),
                            "value": str(generate_example(p.get("schema", {}), set()) or ""),
                            "description": p.get("description", ""),
                        }
                    )

            # Request body
            body = detail.get("requestBody")
            if body:
                content = body.get("content", {})
                if "application/json" in content:
                    schema = content["application/json"].get("schema", {})
                    example = generate_example(schema, set())
                    request_item["body"] = {
                        "mode": "raw",
                        "raw": json.dumps(example if example is not None else {}, indent=2),
                        "options": {"raw": {"language": "json"}},
                    }

            # Auth
            security = detail.get("security", [])
            if security:
                request_item["auth"] = {
                    "type": "bearer",
                    "bearer": [
                        {
                            "key": "token",
                            "value": "{{accessToken}}",
                            "type": "string",
                        }
                    ],
                }

            # Responses
            responses = detail.get("responses", {})
            response_list = []
            for status_code, resp in responses.items():
                try:
                    code_int = int(status_code)
                except ValueError:
                    continue
                content = resp.get("content", {})
                body_example = ""
                if "application/json" in content:
                    schema = content["application/json"].get("schema", {})
                    example = generate_example(schema, set())
                    body_example = json.dumps(example if example is not None else {}, indent=2)
                response_list.append(
                    {
                        "name": status_code,
                        "status": (
                            "OK"
                            if code_int < 400
                            else "Client Error"
                            if code_int < 500
                            else "Server Error"
                        ),
                        "code": code_int,
                        "_postman_previewlanguage": "json",
                        "header": [
                            {"key": "X-Request-ID", "value": "e7f8a9b0-1c2d-4e5f-8a9b-0c1d2e3f4a5b"}
                        ],
                        "body": body_example,
                    }
                )

            collection["item"].append(
                {
                    "name": f"{method_upper} {path}",
                    "request": request_item,
                    "response": response_list,
                }
            )

    return collection


def convert_service(service_dir: Path) -> tuple[Path, Path]:
    service = service_dir.name
    yaml_file = service_dir / f"{service}-openapi.yaml"
    json_file = service_dir / f"{service}-openapi.json"
    postman_file = service_dir / f"{service}-openapi.postman_collection.json"

    spec = load_yaml(yaml_file)
    write_json(json_file, spec)

    collection = to_postman(spec)
    write_json(postman_file, collection)

    return json_file, postman_file


def main() -> None:
    service_dirs = sorted(
        [d for d in root.iterdir() if d.is_dir() and d.name.endswith("-service") or d.name == "api-gateway"],
        key=lambda d: d.name,
    )

    if not service_dirs:
        print(f"No service directories found in {root}")
        return

    print(f"Found {len(service_dirs)} services to convert")

    for service_dir in service_dirs:
        json_file, postman_file = convert_service(service_dir)
        print(f"Converted {service_dir.name} -> {json_file.name} and {postman_file.name}")

    print("All YAML files converted successfully")


if __name__ == "__main__":
    main()
