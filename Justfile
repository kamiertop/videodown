# 跨平台编译 + vet 检查，在 pre-commit 中调用。
# Wails 的 macOS 实现包含原生 Objective-C/Cocoa 代码，不能在 Linux 上
# 交叉编译，因此 Darwin 仅在 macOS 主机上检查；完整平台验证由 CI 负责。
export PATH := env_var("HOME") + "/.local/bin/go/bin:" + env_var("PATH")

check:
	@host="$(uname -s)"; if [ "$host" = "Linux" ]; then echo "=== linux/amd64 (CGO) ==="; GOCACHE=/tmp/videodown-go-cache CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /dev/null .; else echo "=== linux/amd64: skipped (host=$host) ==="; fi
	@echo "=== windows/amd64 ==="
	GOCACHE=/tmp/videodown-go-cache CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /dev/null .
	@host="$(uname -s)"; if [ "$host" = "Darwin" ]; then echo "=== darwin/amd64 (native) ==="; GOCACHE=/tmp/videodown-go-cache CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o /dev/null .; else echo "=== darwin/amd64: skipped (requires macOS host) ==="; fi
	@echo "=== vet native ==="
	GOCACHE=/tmp/videodown-go-cache go vet ./...
	@echo "OK"
