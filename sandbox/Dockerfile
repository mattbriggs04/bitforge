FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends gcc g++ libc6-dev python3 python3-numpy rustc ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /work
USER 65534:65534
