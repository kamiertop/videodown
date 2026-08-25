// Package websign runs Douyin's original secsdk JavaScript VM inside goja.
// It is self-contained, performs no network requests and starts no subprocesses.
package websign

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
)

//go:embed environment.js
var environmentSource string

//go:embed vendor/runtime_bundler_34.js
var runtimeBundlerSource string

//go:embed vendor/webmssdk.es5.js
var webmssdkSource string

//go:embed vendor/sdk-glue.js
var sdkGlueSource string

var (
	uifidPattern = regexp.MustCompile(`^[a-fA-F0-9]{64,1000}$`)
	compileOnce  sync.Once
	compileErr   error
	programs     []*goja.Program
)

// Response contains the URL and headers produced by the original webSignUrl VM.
type Response struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

func compilePrograms() error {
	compileOnce.Do(func() {
		for _, source := range []struct {
			name string
			code string
		}{
			{"environment.js", environmentSource},
			{"runtime_bundler_34.js", runtimeBundlerSource},
			{"webmssdk.es5.js", webmssdkSource},
			{"sdk-glue.js", sdkGlueSource},
		} {
			program, err := goja.Compile(source.name, source.code, false)
			if err != nil {
				compileErr = fmt.Errorf("compile %s: %w", source.name, err)
				return
			}
			programs = append(programs, program)
		}
	})
	return compileErr
}

// Sign generates x-secsdk-web-signature using Douyin's embedded JavaScript VM.
// UIFID must come from the same account cookies used for the HTTP request.
func Sign(targetURL, uifid string) (Response, error) {
	var result Response
	if !uifidPattern.MatchString(uifid) {
		return result, errors.New("uifid must be a 64-1000 character hexadecimal string")
	}
	parsed, err := url.Parse(targetURL)
	if err != nil || parsed.Scheme != "https" || (parsed.Hostname() != "douyin.com" && !strings.HasSuffix(parsed.Hostname(), ".douyin.com")) {
		return result, errors.New("target URL must be an HTTPS URL under douyin.com")
	}
	if err := compilePrograms(); err != nil {
		return result, err
	}

	vm := goja.New()
	_ = vm.Set("__uifid", uifid)
	_ = vm.Set("__pageURL", "https://www.douyin.com/user/self")
	_ = vm.Set("atob", func(value string) string {
		decoded, _ := base64.StdEncoding.DecodeString(value)
		return string(decoded)
	})
	_ = vm.Set("btoa", func(value string) string {
		return base64.StdEncoding.EncodeToString([]byte(value))
	})
	_ = vm.Set("__unixMillis", func() int64 { return time.Now().UnixMilli() })

	for _, program := range programs {
		if _, err := vm.RunProgram(program); err != nil {
			return result, fmt.Errorf("initialize secsdk VM: %w", err)
		}
	}

	if _, err := vm.RunString(`window._SdkGlueInit({self:{aid:6383,pageId:6241},bdms:{aid:6383,pageId:6241,paths:["^/aweme/v1/","^/aweme/v2/"],boe:false,ddrt:8.5,ic:8.5}},{});`); err != nil {
		return result, fmt.Errorf("initialize sdk glue: %w", err)
	}
	value, err := vm.RunString(`window.use("webSignUrl")`)
	if err != nil {
		return result, fmt.Errorf("resolve webSignUrl: %w", err)
	}
	sign, ok := goja.AssertFunction(value)
	if !ok {
		return result, errors.New("secsdk did not expose webSignUrl")
	}
	signed, err := sign(goja.Undefined(), vm.ToValue(parsed.String()))
	if err != nil {
		return result, fmt.Errorf("execute webSignUrl: %w", err)
	}
	raw, err := json.Marshal(signed.Export())
	if err != nil {
		return result, fmt.Errorf("encode webSignUrl result: %w", err)
	}
	if err = json.Unmarshal(raw, &result); err != nil {
		return result, fmt.Errorf("decode webSignUrl result: %w", err)
	}
	if result.URL == "" || result.Headers["x-secsdk-web-signature"] == "" {
		return result, errors.New("secsdk policy did not sign this URL")
	}
	return result, nil
}

func XSecSdkWebSignature(targetURL, uifid string) (string, error) {
	var result Response

	if !uifidPattern.MatchString(uifid) {
		return result.URL, errors.New("uifid must be a 64-1000 character hexadecimal string")
	}
	parsed, err := url.Parse(targetURL)
	if err != nil || parsed.Scheme != "https" || (parsed.Hostname() != "douyin.com" && !strings.HasSuffix(parsed.Hostname(), ".douyin.com")) {
		return result.URL, errors.New("target URL must be an HTTPS URL under douyin.com")
	}
	if err := compilePrograms(); err != nil {
		return result.URL, err
	}

	vm := goja.New()
	_ = vm.Set("__uifid", uifid)
	_ = vm.Set("__pageURL", "https://www.douyin.com/user/self")
	_ = vm.Set("atob", func(value string) string {
		decoded, _ := base64.StdEncoding.DecodeString(value)
		return string(decoded)
	})
	_ = vm.Set("btoa", func(value string) string {
		return base64.StdEncoding.EncodeToString([]byte(value))
	})
	_ = vm.Set("__unixMillis", func() int64 { return time.Now().UnixMilli() })

	for _, program := range programs {
		if _, err := vm.RunProgram(program); err != nil {
			return result.URL, fmt.Errorf("initialize secsdk VM: %w", err)
		}
	}

	if _, err := vm.RunString(`window._SdkGlueInit({self:{aid:6383,pageId:6241},bdms:{aid:6383,pageId:6241,paths:["^/aweme/v1/","^/aweme/v2/"],boe:false,ddrt:8.5,ic:8.5}},{});`); err != nil {
		return result.URL, fmt.Errorf("initialize sdk glue: %w", err)
	}
	value, err := vm.RunString(`window.use("webSignUrl")`)
	if err != nil {
		return result.URL, fmt.Errorf("resolve webSignUrl: %w", err)
	}
	sign, ok := goja.AssertFunction(value)
	if !ok {
		return result.URL, errors.New("secsdk did not expose webSignUrl")
	}
	signed, err := sign(goja.Undefined(), vm.ToValue(parsed.String()))
	if err != nil {
		return result.URL, fmt.Errorf("execute webSignUrl: %w", err)
	}
	raw, err := json.Marshal(signed.Export())
	if err != nil {
		return result.URL, fmt.Errorf("encode webSignUrl result: %w", err)
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return result.URL, fmt.Errorf("decode webSignUrl result: %w", err)
	}

	if result.URL == "" || result.Headers["x-secsdk-web-signature"] == "" {
		return result.URL, errors.New("secsdk policy did not sign this URL")
	}

	return result.URL, nil
}
