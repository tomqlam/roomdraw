FROM golang:1.25-alpine

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install Air for live reloading (moved from cosmtrek to air-verse)
RUN go install github.com/air-verse/air@latest

# Copy the rest of the code
COPY . .

# Use Air for live reloading
CMD ["air", "-c", ".air.toml"]
