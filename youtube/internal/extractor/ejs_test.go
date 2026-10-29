package extractor

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestEmbeddedSolverContract(t *testing.T) {
	for _, kind := range []string{"n", "sig"} {
		request := map[string]any{"type": "preprocessed", "preprocessed_player": "_result.n=(value)=>value.split('').reverse().join('');_result.sig=(value)=>value.slice(1);", "requests": []any{map[string]any{"type": kind, "challenges": []string{"abcd"}}}}
		got, err := runChallenge(request, "abcd")
		want := "dcba"
		if kind == "sig" {
			want = "bcd"
		}
		if err != nil || got != want {
			t.Fatalf("%s: %s %v", kind, got, err)
		}
	}
	if _, err := runChallenge(map[string]any{"type": "preprocessed", "preprocessed_player": "", "requests": []any{map[string]any{"type": "n", "challenges": []string{"abcd"}}}}, "abcd"); err == nil {
		t.Fatal("missing solver accepted")
	}
}

func TestSolverHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := runChallengeContext(ctx, nil, "x"); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation ignored")
	}
}

func TestSolverInterruptsRunningScript(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	request := map[string]any{"type": "preprocessed", "preprocessed_player": "while(true) {}", "requests": []any{map[string]any{"type": "n", "challenges": []string{"x"}}}}
	if _, err := runChallengeContext(ctx, request, "x"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("running script was not canceled: %v", err)
	}
}
