FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o reminder-service .

FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/reminder-service .

ENV PORT=4010
EXPOSE 4010

CMD ["./reminder-service"]
