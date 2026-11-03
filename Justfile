# Native builds need neither Node.js nor CGO.
# Pre-commit hooks and IDE-spawned git run with a minimal environment where a
# user-local Go toolchain is not on PATH; prepend its conventional location.
export PATH := env_var('HOME') + '/.local/bin/go/bin:' + env_var('PATH')

default: test

# Fast feedback: tests and vet, no compiling.
test:
    CGO_ENABLED=0 go test ./...
    CGO_ENABLED=0 go tool mygo vet .

# Full check also compiles every release platform.
check: test
    CGO_ENABLED=0 go build -o /tmp/videodown-check .
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /tmp/videodown-check.exe .
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o /tmp/videodown-check-darwin .

# Development runs only the app; keep its data in the project directory.
run:
    CGO_ENABLED=0 go run . -data .
