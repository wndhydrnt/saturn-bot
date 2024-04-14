ARG BASE
FROM ghcr.io/wndhydrnt/saturn-sync:${BASE}
USER root
RUN apt-get update && \
    apt-get install --no-install-recommends -y python3=3.11.2-1+b1 python3-dev=3.11.2-1+b1 python-is-python3=3.11.1-3 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
USER saturn-sync
