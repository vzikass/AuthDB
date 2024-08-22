# Stage 1: Build
FROM golang:1.22 as builder
WORKDIR /app

COPY go.mod go.sum ./
# RUN export GOPROXY=direct && go mod download
RUN go mod download
# RUN GOPROXY=direct go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main .
# RUN go build -o main .

# Stage 2: Test
FROM golang:1.22 as tester
WORKDIR /app
COPY . ./
RUN go test -v ./...
# COPY --from=builder /app/main /app/
# WORKDIR /app/cmd/repository
# RUN go test -v ./cmd/app/repository...

# FROM golang:1.22 as tester
# WORKDIR /app

# COPY . ./
# RUN go test -v ./cmd/app/repository/...


# Stage 3: PostgreSQL with custom configuration
FROM postgres:latest as postgres-config
WORKDIR /app

ENV POSTGRES_HOST_AUTH_METHOD=md5

# Stage 4: Runtime
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app /app
COPY public /app/public
COPY db.env /app/db.env
# COPY --from=builder /app/main /app/
# COPY public /app/public
# COPY db.env /app/db.env


RUN apk add --no-cache postgresql14-client


EXPOSE 4444
CMD ["/app/main", "-port", "4444"]
