package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var packagePath = flag.String("package", "", "Path to downloaded update package.")
var targetPath = flag.String("target", "", "Path of the application to update.")

func main() {
	flag.Parse()
	target := *targetPath
	if target == "" {
		var err error
		target, err = os.Executable()
		if err != nil {
			fatal(err)
		}
	}
	archivePath := *packagePath
	var err error
	if archivePath == "" {
		fatal(errors.New("package is required"))
	}
	// The downloaded package is no longer needed after extraction.
	defer os.Remove(archivePath)
	// Wait for the main process to release its database and executable files.
	time.Sleep(2 * time.Second)
	filePath, err := extract(archivePath, filepath.Dir(target))
	if err != nil {
		fatal(err)
	}
	defer os.RemoveAll(filepath.Dir(filePath))
	if err := replace(target, filePath); err != nil {
		fatal(err)
	}
	cmd := exec.Command(target)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := cmd.Start(); err != nil {
		fatal(fmt.Errorf("restart application: %w", err))
	}
}

func fatal(err error) { fmt.Fprintln(os.Stderr, "updater:", err); os.Exit(1) }

func extract(archivePath, baseDir string) (string, error) {
	dir, err := os.MkdirTemp(baseDir, ".videodown-update-extract-")
	if err != nil {
		return "", err
	}
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(archivePath, dir)
	}
	return extractTarGz(archivePath, dir)
}

func extractZip(path, dir string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()
	for _, f := range r.File {
		name := filepath.Base(f.Name)
		if name != "videodown" && name != "videodown.exe" {
			continue
		}
		in, err := f.Open()
		if err != nil {
			return "", err
		}
		outPath := filepath.Join(dir, name)
		out, err := os.Create(outPath)
		if err != nil {
			in.Close()
			return "", err
		}
		_, copyErr := io.Copy(out, in)
		in.Close()
		out.Close()
		if copyErr != nil {
			return "", copyErr
		}
		if runtime.GOOS != "windows" {
			_ = os.Chmod(outPath, 0755)
		}
		return outPath, nil
	}
	return "", errors.New("application binary not found in archive")
}

func extractTarGz(path, dir string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		if h.Typeflag != tar.TypeReg {
			continue
		}
		name := filepath.Base(h.Name)
		if name != "videodown" && name != "videodown.exe" {
			continue
		}
		outPath := filepath.Join(dir, name)
		out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return "", err
		}
		_, copyErr := io.Copy(out, tr)
		out.Close()
		if copyErr != nil {
			return "", copyErr
		}
		return outPath, nil
	}
	return "", errors.New("application binary not found in archive")
}

func replace(target, source string) error {
	backup := target + ".old"
	_ = os.Remove(backup)
	if err := os.Rename(target, backup); err != nil {
		return fmt.Errorf("backup current application: %w", err)
	}
	if err := os.Rename(source, target); err != nil {
		if errors.Is(err, syscall.EXDEV) {
			if copyErr := copyFile(source, target); copyErr == nil {
				_ = os.Remove(backup)
				return nil
			} else {
				_ = os.Rename(backup, target)
				return fmt.Errorf("install update across filesystems: %w", copyErr)
			}
		}
		_ = os.Rename(backup, target)
		return fmt.Errorf("install update: %w", err)
	}
	_ = os.Remove(backup)
	return nil
}

func copyFile(source, target string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
