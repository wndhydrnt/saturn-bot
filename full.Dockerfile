ARG BASE
FROM ${BASE}

RUN apt-get update && \
    apt-get install --no-install-recommends -y python3=3.11.2-1+b1 python3-dev=3.11.2-1+b1 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
