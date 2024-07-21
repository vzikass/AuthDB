# syntax=docker/dockerfile:1

FROM golang:1.22.5 as builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

FROM alpine:latest
WORKDIR /app

COPY --from=builder /server /server

COPY public /app/public
COPY db.env /app/db.env

EXPOSE 4444
CMD ["/server"]