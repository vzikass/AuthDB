# Stage 1: Build
FROM golang:1.23 as builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main AuthDB/cmd

# Stage 2: PostgreSQL with custom configuration
FROM postgres:latest as postgres-config
WORKDIR /app

ENV POSTGRES_HOST_AUTH_METHOD=md5

# Stage 3: Runtime
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app /app
COPY public /app/public
COPY ./configs/db.env /app/configs/db.env

RUN apk add --no-cache postgresql-client

EXPOSE 4444

CMD ["/app/main", "-port", "4444"]
