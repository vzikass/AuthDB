# Stage 1: Build
FROM golang:1.22 as builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download 

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

# Stage 2: Test
FROM golang:1.22 as tester
WORKDIR /app

COPY --from=builder /app /app
RUN go test ./...

# Stage 3: Runtime
FROM alpine:latest
WORKDIR /app

COPY --from=builder /server /server

COPY public /app/public
COPY db.env /app/db.env


EXPOSE 4444
CMD ["/server"]