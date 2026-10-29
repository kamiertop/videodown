//go:build linux

package gui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/kamiertop/videodown/internal/storage"
)

// Native lifecycle checks are opt-in: they need a desktop session and run in
// subprocesses because MyGo owns the main OS thread and application loop.
func TestMain(m *testing.M) {
	if action := os.Getenv("VIDEODOWN_GUI_LIFECYCLE_CHILD"); action != "" {
		store, err := storage.OpenMemory()
		if err != nil {
			panic(err)
		}
		if err := store.SetCloseToTray(false); err != nil {
			panic(err)
		}
		s := NewShellWithOptions(Options{Preferences: store})
		mygo.App.WhenReady(func() {
			s.OpenWindow()
			time.AfterFunc(500*time.Millisecond, func() {
				mygo.RunOnMain(func() {
					switch action {
					case "quit":
						mygo.App.Quit()
					case "close":
						s.win.Close() // Use a saved exit preference, as the title bar does.
					case "hide":
						if s.tray == nil {
							fmt.Fprintln(os.Stderr, "tray unavailable")
							os.Exit(1)
						}
						s.closing.CloseToTray = true
						s.win.Close()
						if s.win.IsDestroyed() || s.win.IsVisible() {
							fmt.Fprintln(os.Stderr, "close-to-tray did not hide the live window")
							os.Exit(1)
						}
						mygo.App.Quit() // Explicit Quit must still exit after hiding.
					default:
						fmt.Fprintln(os.Stderr, "unknown lifecycle action", action)
						os.Exit(1)
					}
				})
			})
		})
		if err := mygo.App.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := store.Close(); err != nil {
			panic(err)
		}
		fmt.Fprintln(os.Stdout, "native event loop returned")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestNativeQuit(t *testing.T) {
	if os.Getenv("VIDEODOWN_GUI_LIFECYCLE") != "1" {
		t.Skip("set VIDEODOWN_GUI_LIFECYCLE=1 to test with a real desktop")
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"quit", "close", "hide"} {
		t.Run(action, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, executable)
			cmd.Env = append(os.Environ(), "VIDEODOWN_GUI_LIFECYCLE_CHILD="+action)
			// Include Go stacks in failures instead of leaving a hidden test process.
			cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGQUIT) }
			cmd.WaitDelay = 2 * time.Second
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("native %s did not exit cleanly: %v\n%s", action, err, output)
			}
			if !strings.Contains(string(output), "native event loop returned") {
				t.Fatalf("event loop did not finish: %s", output)
			}
		})
	}
}
