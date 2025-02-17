ARG BASE
FROM ghcr.io/wndhydrnt/saturn-bot:${BASE}
USER root
RUN apt-get update && \
    apt-get install --no-install-recommends -y python3=3.11.2-1+b1 python3-dev=3.11.2-1+b1 python3-venv=3.11.2-1+b1 python-is-python3=3.11.2-1+deb12u1 python3-pip=23.0.1+dfsg-1 default-jre=2:1.17-74 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
USER saturn-bot
RUN python3 -m venv --prompt saturn-bot /home/saturn-bot/venv
ENV SATURN_BOT_PYTHONPATH=/home/saturn-bot/venv/bin/python
