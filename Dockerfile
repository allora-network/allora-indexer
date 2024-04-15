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
from debian:bookworm-slim
WORKDIR /
COPY --from=gobuilder allora-cosmos-pump allora-cosmos-pump
COPY --from=gobuilder /usr/local/bin/allorad /usr/local/bin/allorad
RUN chmod +x /usr/local/bin/allorad

EXPOSE 8080
ENTRYPOINT ["./allora-cosmos-pump"]
