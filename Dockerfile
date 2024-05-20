# node build
from golang:1.21.7-bookworm as gobuilder
WORKDIR /
RUN apt-get update && apt-get install -y curl

# Detect the architecture and download the appropriate binary
ARG TARGETARCH
RUN if [ "$TARGETARCH" = "arm64" ]; then \
        curl -L https://github.com/allora-network/allora-chain/releases/download/v0.0.10/allorad_linux_arm64 -o /usr/local/bin/allorad; \
    else \
        curl -L https://github.com/allora-network/allora-chain/releases/download/v0.0.10/allorad_linux_amd64 -o /usr/local/bin/allorad; \
    fi

COPY . .
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

WORKDIR /
<<<<<<< Updated upstream
COPY --from=gobuilder allora-cosmos-pump allora-cosmos-pump
COPY --from=gobuilder /usr/local/bin/allorad /usr/local/bin/allorad
RUN chmod +x /usr/local/bin/allorad

EXPOSE 8080
ENTRYPOINT ["./allora-cosmos-pump"]
=======
# Detect the architecture and download the appropriate binary
# RUN curl -L https://github.com/allora-network/allora-chain/releases/download/untagged-897b45d53401bc762977/allorad_linux_amd64 -o /usr/local/bin/allorad; \
#     chmod a+x /usr/local/bin/allorad

COPY --from=gobuilder allora-cosmos-pump /usr/local/bin/allora-cosmos-pump
COPY allorad /usr/local/bin/allorad
# EXPOSE 8080
ENTRYPOINT ["allora-cosmos-pump"]
>>>>>>> Stashed changes
