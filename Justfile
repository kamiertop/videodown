# Native builds need neither Node.js nor CGO.
check:
    CGO_ENABLED=0 go test ./...
    CGO_ENABLED=0 go tool mygo vet .
    CGO_ENABLED=0 go build -o /tmp/videodown-check .
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /tmp/videodown-check.exe .
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o /tmp/videodown-check-darwin .

# Development runs only the app; keep its data in the project directory.
run:
    CGO_ENABLED=0 go run . -data .
