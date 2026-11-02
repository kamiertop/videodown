package extractor

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dop251/goja"
)

//go:embed ejs/solver.js
var ejsSource string

var preparedPlayers = struct {
	sync.Mutex
	code map[[32]byte]string
}{code: map[[32]byte]string{}}

// solvePlayerChallenge runs the pinned EJS solver in an isolated Goja VM.
// No filesystem, process or network APIs are exposed to the player script.
func solvePlayerChallenge(ctx context.Context, config playerConfig, kind, challenge string) (string, error) {
	key := sha256.Sum256(config)
	preparedPlayers.Lock()
	prepared := preparedPlayers.code[key]
	preparedPlayers.Unlock()
	request := map[string]any{"type": "player", "player": string(config), "output_preprocessed": true, "requests": []any{map[string]any{"type": kind, "challenges": []string{challenge}}}}
	if prepared != "" {
		request["type"] = "preprocessed"
		request["preprocessed_player"] = prepared
		delete(request, "player")
	}

	return runChallengeContext(ctx, request, challenge)
}
func runChallenge(request any, challenge string) (string, error) {
	return runChallengeContext(context.Background(), request, challenge)
}

func runChallengeContext(ctx context.Context, request any, challenge string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	data, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	vm := goja.New()
	stop := context.AfterFunc(ctx, func() { vm.Interrupt(ctx.Err()) })
	defer stop()
	timer := time.AfterFunc(20*time.Second, func() { vm.Interrupt("player solver timed out") })
	defer timer.Stop()
	if _, err := vm.RunString(ejsSource); err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", fmt.Errorf("加载播放器求解器失败：%w", err)
	}
	if err := vm.Set("__requestJSON", string(data)); err != nil {
		return "", err
	}
	value, err := vm.RunString(`JSON.stringify(EJSSolver.default(JSON.parse(__requestJSON)))`)
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", fmt.Errorf("播放器签名计算失败：%w", err)
	}
	var result struct {
		Type      string `json:"type"`
		Prepared  string `json:"preprocessed_player"`
		Error     string `json:"error"`
		Responses []struct {
			Type  string            `json:"type"`
			Error string            `json:"error"`
			Data  map[string]string `json:"data"`
		} `json:"responses"`
	}
	if err := json.Unmarshal([]byte(value.String()), &result); err != nil {
		return "", err
	}
	if result.Type != "result" || len(result.Responses) != 1 {
		return "", errors.New("播放器求解器未返回结果")
	}
	if result.Prepared != "" {
		if input, ok := request.(map[string]any); ok {
			if player, ok := input["player"].(string); ok {
				key := sha256.Sum256([]byte(player))
				preparedPlayers.Lock()
				if len(preparedPlayers.code) >= 8 {
					for old := range preparedPlayers.code {
						delete(preparedPlayers.code, old)
						break
					}
				}
				preparedPlayers.code[key] = result.Prepared
				preparedPlayers.Unlock()
			}
		}
	}
	response := result.Responses[0]
	if response.Type != "result" {
		return "", fmt.Errorf("播放器挑战求解失败：%s", response.Error)
	}
	decoded := response.Data[challenge]
	if decoded == "" {
		return "", errors.New("播放器签名为空")
	}
	return decoded, nil
}
