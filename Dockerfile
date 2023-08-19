FROM golang:1.20.4 as builder

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nostrinbot ./cmd/nostrinbot

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /project/

COPY --from=builder /app/nostrinbot .

CMD ["./nostrinbot"]
