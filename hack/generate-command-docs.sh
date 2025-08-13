#!/usr/bin/env bash

rm ./docs/reference/commands/*.txt
find ./docs/reference/commands/ -type f -name "*.md" -print0 | xargs -0 basename -s .md | xargs -I % sh -c 'saturn-bot % --help > ./docs/reference/commands/%.txt'
