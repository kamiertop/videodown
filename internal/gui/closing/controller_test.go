package closing

import (
	"errors"
	"testing"
)

type memoryPreferences struct {
	enabled, configured bool
	readErr, writeErr   error
}

func (p *memoryPreferences) CloseToTray() (bool, bool, error) {
	return p.enabled, p.configured, p.readErr
}
func (p *memoryPreferences) SetCloseToTray(value bool) error {
	if p.writeErr != nil {
		return p.writeErr
	}
	p.enabled, p.configured = value, true
	return nil
}
func TestCloseDecisions(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		configured, enabled, tray bool
		active                    int
		want                      Action
	}{
		{"first close", false, true, true, 0, Choose}, {"first close with tasks", false, true, true, 2, Choose},
		{"hide", true, true, true, 2, Hide}, {"exit", true, false, true, 0, Exit},
		{"active tasks", true, false, true, 2, ConfirmTasks}, {"no tray", true, true, false, 0, Exit},
		{"no tray with tasks", true, true, false, 2, ConfirmTasks},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := New(&memoryPreferences{configured: tc.configured, enabled: tc.enabled})
			if got := c.OnClose(tc.tray, tc.active); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
func TestQuitAndPreferenceFailure(t *testing.T) {
	repo := &memoryPreferences{enabled: true, configured: true}
	c := New(repo)
	if c.OnQuit(0, false) != Exit || c.OnQuit(2, false) != ConfirmTasks || c.OnQuit(2, true) != Exit {
		t.Fatal("explicit quit must ignore close-to-tray and protect tasks")
	}
	c.Saving = true
	if c.OnClose(true, 0) != Wait || c.OnQuit(0, true) != Wait {
		t.Fatal("must wait for pending preference writes")
	}
	repo.writeErr = errors.New("disk full")
	c.Saved(false, c.WritePreference(false))
	if !c.CloseToTray || !c.Configured || c.Error == "" || c.Saving {
		t.Fatal("failed writes must preserve last saved choice")
	}
	repo.writeErr = nil
	c.Saved(false, c.WritePreference(false))
	if c.CloseToTray || !repo.configured || repo.enabled || c.Error != "" {
		t.Fatal("saved choice not committed")
	}
	repo.readErr = errors.New("database unavailable")
	c = New(repo)
	if c.Configured || c.Error == "" || c.OnClose(true, 0) != Choose {
		t.Fatal("read failures must prompt instead of silently exiting")
	}
}
