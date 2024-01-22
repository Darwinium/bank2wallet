# Use an official Golang runtime as a parent image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy only the necessary directories and files into the container
COPY certificates/ /app/certificates/
COPY template/ /app/template/
COPY *.go /app/
COPY go.mod /app/
COPY go.sum /app/
COPY .env /app/

# Create the working directories
RUN mkdir -p /app/b2wData
RUN mkdir -p /app/b2wData/tmp
RUN mkdir -p /app/b2wData/passes
# Define this derictory as a volume
VOLUME ["/app/b2wData"]

# Install OpenSSL
RUN apt-get update && apt-get install -y openssl zip

# Download and install any required third party dependencies into the container
RUN go mod download
RUN go mod verify

# Build the Go app
RUN go build -o bank2wallet .

# Expose port 8080 to the outside world
EXPOSE 8080

# Run the app
CMD ["/app/bank2wallet"]
