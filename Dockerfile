# Use the official Golang image as a base image
FROM golang:1.22.6-alpine

# Set the Current Working Directory inside the container
WORKDIR /app

ADD https://github.com/pressly/goose/releases/download/v3.7.0/goose_linux_x86_64 /bin/goose
RUN chmod +x /bin/goose

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main ./cmd/main.go

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]