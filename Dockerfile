# Use Go 1.20 base image with build tools
FROM golang:1.20

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for better cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Download dependencies and build the app
RUN go mod tidy
RUN go build -o league-sim main.go

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./league-sim"] 