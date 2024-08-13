FROM golang:1.22-bookworm AS gobuilder
ADD . /src
WORKDIR /src
RUN go build .

# final image
FROM debian:bookworm-slim

RUN apt update && \
    apt -y dist-upgrade && \
    apt install -y --no-install-recommends \
        curl jq \
        tzdata \
        bc \
        ca-certificates && \
    echo "deb http://deb.debian.org/debian testing main" >> /etc/apt/sources.list && \
    apt update && \
    apt install -y --no-install-recommends -t testing \
      zlib1g \
      libgnutls30 \
      perl-base && \
    rm -rf /var/cache/apt/*

# Detect the architecture and download the appropriate binary
ARG TARGETARCH="amd64"
RUN mkdir -p /usr/local/bin/previous/v2 && \
    curl -L https://github.com/allora-network/allora-chain/releases/download/v0.2.14/allorad_linux_${TARGETARCH} -o /usr/local/bin/previous/v2/allorad; \
    curl -L https://github.com/allora-network/allora-chain/releases/download/v0.3.0/allorad_linux_${TARGETARCH} -o /usr/local/bin/allorad; \
    chmod -R 777 /usr/local/bin/allorad && \
    chmod -R 777 /usr/local/bin/previous/v2/allorad

COPY --from=gobuilder /src/allora-cosmos-pump /usr/local/bin/allora-cosmos-pump
# EXPOSE 8080
ENTRYPOINT ["allora-cosmos-pump"]
