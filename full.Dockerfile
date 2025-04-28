ARG BASE
FROM ghcr.io/wndhydrnt/saturn-bot:${BASE}
ENV VIRTUALENV_DIR=/var/virtualenvs/saturn-bot
USER root
RUN apt-get update && \
    apt-get install --no-install-recommends -y python3=3.11.2-1+b1 python3-dev=3.11.2-1+b1 python3-venv=3.11.2-1+b1 python-is-python3=3.11.2-1+deb12u1 python3-pip=23.0.1+dfsg-1 default-jre=2:1.17-74 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    mkdir -p "${VIRTUALENV_DIR}" && \
    chown 1001:1001 "${VIRTUALENV_DIR}"
USER saturn-bot
RUN python3 -m venv --prompt saturn-bot "${VIRTUALENV_DIR}"
ENV SATURN_BOT_PYTHONPATH=${VIRTUALENV_DIR}/bin/python
