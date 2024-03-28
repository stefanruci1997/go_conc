# Start from a golang base image
FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o app .

# Command to run the executable
CMD ["./app"]
