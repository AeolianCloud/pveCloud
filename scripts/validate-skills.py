#!/usr/bin/env python3

from __future__ import annotations

import re
import sys
from pathlib import Path

try:
    import yaml
except ImportError:
    print("error: PyYAML is required to validate SKILL.md front matter", file=sys.stderr)
    sys.exit(2)


REPO_ROOT = Path(__file__).resolve().parents[1]
SKILLS_ROOT = REPO_ROOT / ".codex" / "skills"
REQUIRED_FIELDS = ("name", "description")


def main() -> int:
    skill_files = sorted(SKILLS_ROOT.glob("*/SKILL.md"))
    if not skill_files:
        print("error: no SKILL.md files found under .codex/skills", file=sys.stderr)
        return 1

    failures: list[str] = []
    for path in skill_files:
        failures.extend(validate_skill(path))

    if failures:
        for failure in failures:
            print(failure, file=sys.stderr)
        return 1

    print(f"validated {len(skill_files)} skill file(s)")
    return 0


def validate_skill(path: Path) -> list[str]:
    failures: list[str] = []
    relative = path.relative_to(REPO_ROOT)
    text = path.read_text(encoding="utf-8")

    match = re.match(r"\A---\n(.*?)\n---\n", text, re.DOTALL)
    if not match:
        return [f"{relative}: missing YAML front matter"]

    front_matter = match.group(1)
    try:
        metadata = yaml.safe_load(front_matter)
    except yaml.YAMLError as error:
        failures.append(f"{relative}: invalid YAML front matter: {error}")
        failures.extend(find_probable_unquoted_colons(relative, front_matter))
        return failures

    if not isinstance(metadata, dict):
        return [f"{relative}: front matter must be a mapping"]

    for field in REQUIRED_FIELDS:
        value = metadata.get(field)
        if not isinstance(value, str) or not value.strip():
            failures.append(f"{relative}: `{field}` must be a non-empty string")

    expected_name = path.parent.name
    actual_name = metadata.get("name")
    if isinstance(actual_name, str) and actual_name != expected_name:
        failures.append(f"{relative}: `name` should match directory name `{expected_name}`")

    return failures


def find_probable_unquoted_colons(relative: Path, front_matter: str) -> list[str]:
    failures: list[str] = []
    scalar_pattern = re.compile(r"^\s*(name|description):\s+[^'\"|>].*:\s+")
    for line_number, line in enumerate(front_matter.splitlines(), start=2):
        if scalar_pattern.search(line):
            failures.append(
                f"{relative}:{line_number}: value contains `: ` and probably needs quotes"
            )
    return failures


if __name__ == "__main__":
    sys.exit(main())
