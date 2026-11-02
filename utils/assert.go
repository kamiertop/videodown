package utils

import (
	"testing"
)

func Equal[T comparable](t testing.TB, expected, actual T) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected %#v, got %#v", expected, actual)
	}
}

func NoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, got %#v", err)
	}
}
