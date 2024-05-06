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
        json_path = matches[0][2]
        after = matches[0][3]
        schema_path = os.path.abspath(os.path.join(os.path.dirname(page.file.abs_src_path), path))
        with open(schema_path) as f:
            data = json.load(f)

        content = JSONPath(json_path).parse(data)[0]
        result.append(before + content + after)

    return "\n".join(result)
