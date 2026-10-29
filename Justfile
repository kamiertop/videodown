# Native builds need neither Node.js nor CGO.
check:
    CGO_ENABLED=0 go test ./...
    CGO_ENABLED=0 go tool mygo vet .
    CGO_ENABLED=0 go build -o /tmp/videodown-check .
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /tmp/videodown-check.exe .
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o /tmp/videodown-check-darwin .

# Build one portable folder containing the app, updater and icon.
package version="dev":
    CGO_ENABLED=0 go run ./cmd/package -version {{quote(version)}}

run:
    CGO_ENABLED=0 go run .
