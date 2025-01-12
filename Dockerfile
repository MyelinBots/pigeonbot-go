FROM golang:1.22.10-alpine3.19 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod tidy

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Path: Dockerfile
# golang dockerfile
FROM gcr.io/distroless/static-debian10

EXPOSE 3000
# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder --chown=nonroot:nonroot /app/main .

ARG VERSION

ENV VERSION=$VERSION

USER nonroot

CMD ["./main", "serve"]

# Expose port 8080 to the outside world