package extractor

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestSIDAuthorizationAndClientHeaders(t *testing.T) {
	now := time.Unix(123, 0)
	hash := sha1.Sum([]byte("123 fixture https://www.youtube.com"))
	if got := sidAuthorization("__Secure-3PAPISID=fixture", "https://www.youtube.com", now); !strings.Contains(got, fmt.Sprintf("SAPISIDHASH 123_%x", hash)) || !strings.Contains(got, "SAPISID3PHASH") {
		t.Fatal("incorrect session signature")
	}
	called := false
	c := &Client{Cookie: "SAPISID=fixture; LOGIN_INFO=fixture", HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		data := `<script src="/s/player/test/player_ias/base.js"></script>`
		if strings.HasSuffix(r.URL.Path, "base.js") {
			data = "signatureTimestamp:12345"
		}
		if strings.Contains(r.URL.Path, "/youtubei/") {
			called = true
			if !strings.HasPrefix(r.Header.Get("Authorization"), "SAPISIDHASH ") || r.Header.Get("X-Youtube-Client-Name") != "7" || r.Header.Get("X-Origin") != "https://www.youtube.com" {
				t.Error("missing authenticated client headers")
			}
			var request innertubeRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Error(err)
			}
			if request.Context.Client.DeviceMake != "Samsung" || request.Context.Client.DeviceModel != "UKS9800" || request.Context.Client.OSName != "Tizen" || request.Context.Client.OSVersion != "2.4.0" {
				t.Error("missing television device context")
			}
			if request.Params != "" {
				t.Error("unexpected player params")
			}
			if request.Context.Client.ClientName != "TVHTML5" || request.PlaybackContext.ContentPlaybackContext.SignatureTimestamp != 12345 {
				t.Error("invalid authenticated player context")
			}
			if request.Context.Client.VisitorData != r.Header.Get("X-Goog-Visitor-Id") {
				t.Error("visitor mismatch")
			}
			data = `{"playabilityStatus":{"status":"OK"},"streamingData":{"formats":[{"itag":18,"url":"https://fixture.googlevideo.com/media","mimeType":"video/mp4"}]}}`
		} else if r.Header.Get("Authorization") != "" {
			t.Error("authorization leaked outside API")
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(data)), Request: r}, nil
	})}}
	if _, err := c.GetVideoContext(context.Background(), "KWGhxyu5osA"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("no player request")
	}
}

func TestSessionConfigAndAccountBoundSignature(t *testing.T) {
	c := &Client{}
	c.parseSessionConfig(`<script>ytcfg.set({"IGNORED":true});ytcfg.set({"VISITOR_DATA":"visitor-fixture","DATASYNC_ID":"channel-fixture||user-fixture","SESSION_INDEX":2,"LOGGED_IN":true});</script>`)
	if loggedIn, known := c.AccountState(); !known || !loggedIn {
		t.Fatal("login state not parsed")
	}
	if c.visitorId.value != "visitor-fixture" || c.userSessionID != "user-fixture" || c.delegatedSessionID != "channel-fixture" || c.sessionIndex != "2" {
		t.Fatal("lost session context")
	}
	hash := sha1.Sum([]byte("user-fixture 123 fixture https://www.youtube.com"))
	want := fmt.Sprintf("SAPISIDHASH 123_%x_u", hash)
	if got := sidAuthorization("SAPISID=fixture", "https://www.youtube.com", time.Unix(123, 0), c.userSessionID); got != want {
		t.Fatal("incorrect account-bound signature")
	}
}
