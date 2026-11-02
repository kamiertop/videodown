// Command package builds a portable application folder from the repository root.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	out := flag.String("out", "dist/videodown", "便携程序输出文件夹")
	targetOS := flag.String("os", runtime.GOOS, "目标操作系统")
	targetArch := flag.String("arch", runtime.GOARCH, "目标架构")
	version := flag.String("version", "dev", "应用版本号")
	flag.Parse()
	if err := buildFolder(*out, *targetOS, *targetArch, *version); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func buildFolder(out, targetOS, targetArch, version string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	icon, err := os.ReadFile(filepath.Join(root, "internal", "gui", "assets", "appicon.png"))
	if err != nil {
		return fmt.Errorf("请从仓库根目录执行打包命令: %w", err)
	}
	out, err = filepath.Abs(out)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(out, 0755); err != nil {
		return err
	}
	ext := ""
	if targetOS == "windows" {
		ext = ".exe"
	}
	if targetOS == "windows" {
		resource, err := windowsIconObject(icon, targetArch)
		if err != nil {
			return err
		}
		path := filepath.Join(root, "zz_appicon_windows_"+targetArch+".syso")
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer os.Remove(path)
		_, writeErr := f.Write(resource)
		closeErr := f.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	ldflags := "-s -w -X github.com/kamiertop/videodown/internal/updater.Version=" + version + " -X github.com/kamiertop/videodown/logger.Mode=prod"
	if targetOS == "windows" {
		ldflags += " -H=windowsgui"
	}
	for _, target := range []struct {
		name, path string
		flags      []string
	}{
		{"videodown" + ext, ".", []string{"-ldflags", ldflags}},
		{"videodown-updater" + ext, "./cmd/updater", nil},
	} {
		args := []string{"build", "-trimpath", "-buildvcs=false", "-o", filepath.Join(out, target.name)}
		args = append(args, target.flags...)
		args = append(args, target.path)
		cmd := exec.Command("go", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+targetOS, "GOARCH="+targetArch)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("构建 %s: %w", target.name, err)
		}
	}
	fmt.Println("便携程序已生成：" + out)
	return nil
}
