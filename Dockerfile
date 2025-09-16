# Use the official lightweight Go image
FROM golang:1.22-alpine

# Set environment vars
ENV GO111MODULE=on

# Create a directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first, then download dependencies
COPY go.mod ./
# COPY go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o service ./cmd

# Command to run your binary
CMD ["./service"]