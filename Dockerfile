# node build
from golang:1.21.7-bookworm as gobuilder
WORKDIR /
COPY . .
RUN go build .

# final image
from debian:bookworm-slim
WORKDIR /
COPY --from=gobuilder allora-cosmos-pump allora-cosmos-pump

EXPOSE 8080 8080
ENTRYPOINT ["./allora-cosmos-pump"]