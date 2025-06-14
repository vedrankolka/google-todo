# Build the CLI binary
build:
    go build -o bin/todo *.go

# Build and run the CLI
run: build
    ./bin/todo

# Run tests
test:
    go test ./...

# Clean build artifacts
clean:
    rm -rf bin

# Install binary to $GOBIN (or $HOME/go/bin)
install:
    go install
