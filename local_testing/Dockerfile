FROM golang:1.24

# Install required packages
RUN apt-get update && apt-get install -y \
    # Required for make test
    make \
    # Required to run bash tests
    bash \
    # Required for fixtures
    python3 \
    # Required for testing
    git \
    # Required for mv operations
    sudo \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Starting from Go 1.20, the go standard library is no loger compiled.
# Setting GODEBUG to "installgoroot=all" restores the old behavior
RUN GODEBUG="installgoroot=all" go install std

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Default command
CMD ["/bin/bash"]