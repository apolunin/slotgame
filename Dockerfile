# Use official Golang image as a builder stage
FROM golang:1.23.2 AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /slotgame ./cmd/slotgame

# Use a minimal image for the runtime stage
FROM alpine:latest

# Set work directory and copy built binary
WORKDIR /app
COPY --from=builder /slotgame /app/slotgame

# Set environment variables with ARG and ENV as default values
ARG DB_DATABASE
ENV DB_DATABASE=${DB_DATABASE}
ARG DB_HOST
ENV DB_HOST=${DB_HOST}
ARG DB_PORT
ENV DB_PORT=${DB_PORT}
ARG DB_PWD
ENV DB_PWD=${DB_PWD}
ARG DB_USR
ENV DB_USR=${DB_USR}
ARG JWT_SECRET
ENV JWT_SECRET=${JWT_SECRET}

# Expose the port your application runs on
EXPOSE 8080

# Run the application
CMD ["/app/slotgame"]
