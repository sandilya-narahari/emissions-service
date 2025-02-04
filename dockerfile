FROM golang:1.21-alpine

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Ensure dependencies are downloaded
RUN go mod tidy && go mod download

# Copy the entire source code
COPY . .

# Set correct working directory before building
WORKDIR /app/cmd/emissions-service

# Build the binary
RUN go build -o /bin/emissions-service main.go

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/bin/emissions-service"]