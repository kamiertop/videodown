package youtube

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/kamiertop/videodown/internal/storage"
)

func TestPlaylistUsesSameAuthenticationModeForContinuation(t *testing.T) {
	const input = "https://www.youtube.com/watch?v=jNQXAC9IVRw&list=PLfixture&index=2"
	for _, tc := range []struct {
		name       string
		guestAlert string
		want       []bool
	}{
		{"public", "", []bool{false, false, false}},
		{"private", "This playlist is private", []bool{false, true, true, true}},
		{"private hidden as missing", "The playlist does not exist.", []bool{false, true, true, true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store, err := storage.OpenMemory()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = store.Close() })
			service := New(store)
			if err := service.SetCookie("SID=fixture"); err != nil {
				t.Fatal(err)
			}
			var calls []bool
			service.transport = transportFunc(func(r *http.Request) (*http.Response, error) {
				authenticated := strings.Contains(r.Header.Get("Cookie"), "SID=fixture")
				calls = append(calls, authenticated)
				var body string
				switch {
				case r.Method == http.MethodPost:
					var request struct{ Continuation string }
					if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Continuation != "page2" {
						t.Errorf("continuation sent to YouTube = %q, %v; want page2", request.Continuation, err)
					}
					body = `{"onResponseReceivedActions":[{"appendContinuationItemsAction":{"continuationItems":[{"playlistVideoRenderer":{"videoId":"KWGhxyu5osA"}}]}}]}`
				case tc.guestAlert != "" && !authenticated:
					body = `var ytInitialData = {"alerts":[{"alertRenderer":{"type":"ERROR","text":{"simpleText":"` + tc.guestAlert + `"}}}]};`
				default:
					body = `ytcfg.set({"VISITOR_DATA":"playlist-visitor"}); var ytInitialData = {"contents":{"playlistVideoListRenderer":{"contents":[{"playlistVideoRenderer":{"videoId":"jNQXAC9IVRw"}},{"continuationItemRenderer":{"continuationEndpoint":{"continuationCommand":{"token":"page2"}}}}]}}};`
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: r}, nil
			})
			page, err := service.Playlist(context.Background(), input, "")
			if err != nil || page.Continuation == "" || len(page.Videos) != 1 {
				t.Fatalf("first page = %+v, %v", page, err)
			}
			page, err = service.Playlist(context.Background(), input, page.Continuation)
			if err != nil || len(page.Videos) != 1 || page.Videos[0].ID != "KWGhxyu5osA" {
				t.Fatalf("next page = %+v, %v", page, err)
			}
			if !slices.Equal(calls, tc.want) {
				t.Errorf("account-cookie use per request = %v; want %v", calls, tc.want)
			}
		})
	}
}

func TestPlaylistStillMissingAfterAccountRetry(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	service := New(store)
	if err := service.SetCookie("SID=fixture"); err != nil {
		t.Fatal(err)
	}
	const body = `var ytInitialData = {"alerts":[{"alertRenderer":{"type":"ERROR","text":{"simpleText":"The playlist does not exist."}}}]};`
	var calls []bool
	service.transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		calls = append(calls, strings.Contains(r.Header.Get("Cookie"), "SID=fixture"))
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: r}, nil
	})
	_, err = service.Playlist(t.Context(), "https://www.youtube.com/playlist?list=PLfixture", "")
	if err == nil || !strings.Contains(err.Error(), "请确认当前账号能打开此播放列表") {
		t.Errorf("Playlist error = %v; want account access guidance", err)
	}
	if !slices.Equal(calls, []bool{false, true}) {
		t.Errorf("account-cookie use per request = %v; want [false true]", calls)
	}
}
