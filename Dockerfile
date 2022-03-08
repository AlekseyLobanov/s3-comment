FROM golang:1.17-buster AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./src ./src

RUN cd /app/src && go build -o ../s3-somment


FROM debian:buster-slim

WORKDIR /app

COPY ./static /app/static
COPY --from=builder /app/s3-somment /app/s3-somment

EXPOSE 8123

# USER nonroot:nonroot

ENTRYPOINT ["/app/s3-somment"]
