package api

import "testing"

func TestParseMid(t *testing.T) {
	const want = "325864133"
	for _, raw := range []string{
		"https://space.bilibili.com/325864133/lists",
		"https://space.bilibili.com/325864133",
		want,
	} {
		if got, err := parseMid(raw); err != nil || got != want {
			t.Errorf("parseMid(%q) = %q, %v; want %q, nil", raw, got, err, want)
		}
	}
}

func TestParseMidMissingUID(t *testing.T) {
	for _, raw := range []string{"space.bilibili.com", "https://space.bilibili.com", "https://space.bilibili.com/", "https://space.bilibili.com/not-a-mid"} {
		if _, err := parseMid(raw); err == nil {
			t.Errorf("expected invalid UID for %q", raw)
		}
	}
}
