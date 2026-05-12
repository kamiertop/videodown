package utils

import (
	"fmt"
	"testing"
)

func TestFileName(t *testing.T) {
	rawName := `
4K音乐 尽在/   
BiliBili\ #4k #music
`
	fmt.Println(FileName(rawName))
}
