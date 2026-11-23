package components

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/kamiertop/videodown/internal/gui/style"
)

func TestSettingsSwitchPointerAndKeyboard(t *testing.T) {
	on, disabled := false, false
	test := ui.NewTester(func(c *ui.Context) {
		style.UseTheme(c)
		SettingsSwitch(c, &on, "测试开关", SwitchOptions{Disabled: disabled})
	}, 200, 100)
	if err := test.Click("测试开关"); err != nil {
		t.Fatal(err)
	}
	test.Advance(150 * time.Millisecond)
	if !on {
		t.Fatal("pointer did not turn switch on")
	}
	test.Key(0, ui.KeySpace)
	test.Advance(150 * time.Millisecond)
	if on {
		t.Fatal("Space did not turn switch off")
	}
	disabled = true
	test.Frame()
	if err := test.Click("测试开关"); err != nil {
		t.Fatal(err)
	}
	test.Key(0, ui.KeySpace)
	test.Advance(150 * time.Millisecond)
	if on {
		t.Fatal("disabled switch accepted input")
	}
}
