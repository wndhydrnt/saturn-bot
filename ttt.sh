#!/usr/bin/env bash

make mdox
git diff --color --exit-code .
