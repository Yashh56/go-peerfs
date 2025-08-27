FROM golang:1.23.10-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o ./out/go-peerfs ./cmd/go-peerfs

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/out/go-peerfs .

COPY ./shared ./shared

RUN mkdir ./downloads

EXPOSE 8000

ENTRYPOINT ["./go-peerfs"]
CMD ["start"]
