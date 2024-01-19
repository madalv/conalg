# Builder stage
FROM golang:1.21 AS builder

WORKDIR /app

# Copy go mod and sum files 
COPY go.mod go.sum ./
RUN go mod download

# Copy the Go application source code into the container
COPY . .

# Build the Go application
RUN go build -o main

# Final stage
FROM scratch

# Set the environment variable, which determines the application mode
ENV APP_MODE=prod

# Copy the executable binary from the builder stage into the final image
COPY --from=builder /app/main /main
COPY --from=builder /app/.prod.env ./

# Run the application
CMD ["/main"]