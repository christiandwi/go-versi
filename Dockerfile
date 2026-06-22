FROM golang:1.25-alpine AS dev

WORKDIR /workspace

COPY go.mod ./
RUN go mod download

COPY . .

ENV GOCACHE=/tmp/go-build

CMD ["go", "test", "./..."]
