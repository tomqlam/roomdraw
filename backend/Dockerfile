FROM golang:1.21

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install Air for live reloading
RUN go install github.com/cosmtrek/air@v1.49.0

# Copy the rest of the code
COPY . .

# Air configuration file
COPY .air.toml ./

# Use Air for live reloading
CMD ["air"]
