FROM golang:alpine

# Install git (needed for go mod tidy if plugins require new deps)
RUN apk add --no-cache git

WORKDIR /app

# Copy necessary files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Make the start script executable
RUN chmod +x start.sh

# Use the script as the entrypoint, NOT the binary directly
CMD ["./start.sh"]