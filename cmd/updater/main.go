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
	defer func() {
		if cleanupErr := os.Remove(archivePath); cleanupErr != nil && !errors.Is(cleanupErr, os.ErrNotExist) {
			_, _ = fmt.Fprintln(os.Stderr, "updater: remove package:", cleanupErr)
		}
	}()
	// Wait for the main process to release its database and executable files.
	time.Sleep(2 * time.Second)
	filePath, err := extract(archivePath, filepath.Dir(target))
	if err != nil {
		fatal(err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(filepath.Dir(filePath)); cleanupErr != nil {
			_, _ = fmt.Fprintln(os.Stderr, "updater: remove extract directory:", cleanupErr)
		}
	}()
	if err := replace(target, filePath); err != nil {
		fatal(err)
	}
	cmd := exec.Command(target)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := cmd.Start(); err != nil {
		fatal(fmt.Errorf("restart application: %w", err))
	}
}

func fatal(err error) { _, _ = fmt.Fprintln(os.Stderr, "updater:", err); os.Exit(1) }

func extract(archivePath, baseDir string) (filePath string, err error) {
	dir, err := os.MkdirTemp(baseDir, ".videodown-update-extract-")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.RemoveAll(dir))
		}
	}()
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(archivePath, dir)
	}
	return extractTarGz(archivePath, dir)
}

func extractZip(path, dir string) (filePath string, err error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, r.Close()) }()
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
			_ = in.Close()
			return "", err
		}
		copyErr := func() (err error) {
			_, copyErr := io.Copy(out, in)
			err = errors.Join(copyErr, in.Close(), out.Close())
			return err
		}()
		if copyErr != nil {
			return "", fmt.Errorf("write extracted file: %w", copyErr)
		}
		if runtime.GOOS != "windows" {
			if err := os.Chmod(outPath, 0755); err != nil {
				return "", fmt.Errorf("set executable permission: %w", err)
			}
		}
		return outPath, nil
	}
	return "", errors.New("application binary not found in archive")
}

func extractTarGz(path, dir string) (outPath string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer func() {
		err = errors.Join(err, f.Close())
	}()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := gz.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close gzip reader: %w", closeErr)
		}
	}()
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
		outPath = filepath.Join(dir, name)
		out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return "", err
		}
		_, copyErr := io.Copy(out, tr)
		if closeErr := out.Close(); closeErr != nil {
			return "", fmt.Errorf("close extracted file: %w", closeErr)
		}
		if copyErr != nil {
			return "", copyErr
		}
		return outPath, nil
	}

	return "", errors.New("application binary not found in archive")
}

func replace(target, source string) error {
	backup := target + ".old"
	if err := os.Remove(backup); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove old backup: %w", err)
	}
	if err := os.Rename(target, backup); err != nil {
		return fmt.Errorf("backup current application: %w", err)
	}
	if err := os.Rename(source, target); err != nil {
		if errors.Is(err, syscall.EXDEV) {
			if copyErr := copyFile(source, target); copyErr == nil {
				if removeErr := os.Remove(backup); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
					return fmt.Errorf("remove old backup: %w", removeErr)
				}
				return nil
			} else {
				if restoreErr := os.Rename(backup, target); restoreErr != nil {
					return fmt.Errorf("install update across filesystems: %w (restore backup: %v)", copyErr, restoreErr)
				}
				return fmt.Errorf("install update across filesystems: %w", copyErr)
			}
		}
		if restoreErr := os.Rename(backup, target); restoreErr != nil {
			return fmt.Errorf("install update: %w (restore backup: %v)", err, restoreErr)
		}
		return fmt.Errorf("install update: %w", err)
	}
	if err := os.Remove(backup); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove old backup: %w", err)
	}
	return nil
}

func copyFile(source, target string) (err error) {
	var in *os.File
	in, err = os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, in.Close())
	}()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		return errors.Join(err, out.Close())
	}

	return out.Close()
}
