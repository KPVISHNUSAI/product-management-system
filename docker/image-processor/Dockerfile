# docker/image-processor/Dockerfile
FROM golang:1.21-alpine

WORKDIR /app

# Install build and runtime dependencies
RUN apk add --no-cache gcc musl-dev vips-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o processor ./image-processor/main.go

CMD ["./processor"]