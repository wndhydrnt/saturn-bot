import json
import os.path
import re
from jsonpath import JSONPath

from mkdocs.config.defaults import MkDocsConfig
from mkdocs.structure.files import Files
from mkdocs.structure.pages import Page


PATTERN = re.compile(r"(.*)\[json-path:(.+):(.+)](.*)")


def on_page_markdown(
    markdown: str, /, *, page: Page, config: MkDocsConfig, files: Files
) -> str | None:
    result: list[str] = []
    for line in markdown.split("\n"):
        matches = PATTERN.findall(line)
        if len(matches) == 0:
            result.append(line)
            continue

        before = matches[0][0]
        path = matches[0][1]
        json_path: str = matches[0][2]
        after = matches[0][3]
        schema_path = os.path.abspath(os.path.join(os.path.dirname(page.file.abs_src_path), path))
        with open(schema_path) as f:
            data = json.load(f)

        json_path_parts = json_path.split("|")
        selector = json_path_parts[0]
        if len(json_path_parts) == 2:
            default_value = json_path_parts[1]
        else:
            default_value = None

        json_path_result = JSONPath(selector).parse(data)
        if len(json_path_result) == 0:
            if default_value is None:
                raise RuntimeError(f"JSON path '{selector}' not found in document")
            else:
                content = default_value
        else:
            content = json_path_result[0]

        result.append(before + content + after)

    return "\n".join(result)
