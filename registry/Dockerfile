FROM golang:1.22.2-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./
COPY *.yaml ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /registry

EXPOSE 8081

ENTRYPOINT ["/registry"]