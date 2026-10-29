package preferences

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func CheckFFmpeg(custom string) (path, version string, err error) {
	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	var candidates []string
	if custom != "" {
		candidates = append(candidates, custom)
	} else {
		if executable, e := os.Executable(); e == nil {
			dir := filepath.Dir(executable)
			candidates = append(candidates, filepath.Join(dir, name), filepath.Join(dir, "bin", name))
		}
		if found, e := exec.LookPath(name); e == nil {
			candidates = append(candidates, found)
		}
	}
	for _, candidate := range candidates {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		output, e := exec.CommandContext(ctx, candidate, "-version").CombinedOutput()
		cancel()
		if e == nil {
			return candidate, strings.SplitN(string(output), "\n", 2)[0], nil
		}
	}
	return "", "", fmt.Errorf("未检测到可用的 FFmpeg，请安装或手动指定路径")
}
