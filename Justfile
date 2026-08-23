# 交叉编译 + vet 检查，在 pre-commit 中调用。
# 仅编译不链接，避免交叉 CGO 链接器问题 —— 真正的 CGO 链接
# 由 CI (release.yml) 在各原生平台上验证。
export PATH := env_var("HOME") + "/.local/bin/go/bin:" + env_var("PATH")

check:
    @echo "=== linux/amd64 (CGO) ==="
    CGO_ENABLED=1 GOOS=linux   GOARCH=amd64 go build -o /dev/null .
    @echo "=== windows/amd64 ==="
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /dev/null .
    @echo "=== darwin/amd64 ==="
    CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o /dev/null .
    @echo "=== vet all ==="
    go vet ./...
    GOOS=darwin GOARCH=amd64 go vet ./...
    GOOS=windows GOARCH=amd64 go vet ./...
    @echo "OK"
