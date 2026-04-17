package utils

import "testing"

func TestParseMid(t *testing.T) {
	mid, err := ParseMid("https://space.bilibili.com/325864133/lists")
	NoError(t, err)
	Equal(t, "325864133", mid)

	mid, err = ParseMid("https://space.bilibili.com/325864133")
	NoError(t, err)
	Equal(t, "325864133", mid)

	mid, err = ParseMid("325864133")
	NoError(t, err)
	Equal(t, "325864133", mid)
}
