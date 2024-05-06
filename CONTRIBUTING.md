# Contributing

## Documentation

Documentation is written in [Markdown](https://www.markdownguide.org/) and rendered by [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/).

### Serve the documentation locally

1. Create a Python virtualenv (needs to be done only once):
   ```shell
   python -m venv --prompt saturn-bot-docs ./venv
   ```
2. Activate the virtualenv:
   ```shell
   source venv/bin/activate
   ```
3. Install dependencies:
   ```shell
   pip install -r mkdocs-requirements.txt
   ```
4. Start the server:
   ```shell
   mkdocs serve
   ```

Open [http://localhost:8000](http://localhost:8000) in a browser.

### Update dependencies

1. Activate the virtualenv:
   ```shell
   source venv/bin/activate
   ```
2. Install whatever dependencies are necessary:
   ```shell
   pip install ...
   ```
3. Update the requirements file:
   ```shell
   pip freeze -l > mkdocs-requirements.txt
   ```
