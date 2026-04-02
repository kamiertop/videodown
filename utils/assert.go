package utils

import (
	"testing"
)

func Equal[T comparable](t *testing.T, expected, actual T) {
	if expected != actual {
		t.Helper()
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func NoError(t *testing.T, err error) {
	if err != nil {
		t.Helper()
		t.Errorf("Expected no error, got %v", err)
	}
}
