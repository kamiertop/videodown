// Adapted from github.com/kkdai/youtube, commit 3876742b95fd1cd96f836f396c2d54be8912a297.
// Copyright (c) 2015 Evan Lin. MIT license: see LICENSE and THIRD_PARTY_NOTICES.md.
package extractor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const ContentPlaybackNonceAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

var ErrNoFormat = errors.New("no video format provided")

// DefaultClient type to use. No reason to change but you could if you wanted to.
var DefaultClient = VisionOSClient

// Profile fields follow yt-dlp's tv_samsung fix (#17723, #17805).
var AuthenticatedClient = ClientInfo{Name: "TVHTML5", Version: "5.20260707", DeviceMake: "Samsung", DeviceModel: "UKS9800", OSName: "Tizen", OSVersion: "2.4.0", UserAgent: "Mozilla/5.0 (SMART-TV; Linux; Tizen 2.4.0) AppleWebKit/538.1 (KHTML, like Gecko) Version/2.4.0 TV Safari/538.1"}

var VisionOSClient = ClientInfo{Name: "VISIONOS", Version: "1.02", DeviceMake: "Apple", OSName: "visionOS", OSVersion: "26.5.23O471", DeviceModel: "RealityDevice17,1", UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 15_7_3) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/26.0 Safari/605.1.15"}

// Client offers methods to download video metadata and video streams.
type Client struct {
	loggedIn                                        *bool
	Cookie                                          string
	userSessionID, delegatedSessionID, sessionIndex string
	// HTTPClient can be used to set a custom HTTP client.
	// If not set, http.DefaultClient will be used
	HTTPClient *http.Client

	// playerCache caches the JavaScript code of a player response
	playerCache playerCache

	client *ClientInfo

	consentID string

	visitorId struct {
		value   string
		updated time.Time
	}
}

func (c *Client) assureClient() {
	if c.consentID == "" {
		c.consentID = strconv.Itoa(rand.Intn(899) + 100)
	}
	if c.client == nil {
		c.client = &DefaultClient
		if hasSID(c.Cookie) {
			c.client = &AuthenticatedClient
		}
	}
}

// GetVideo fetches video metadata
func (c *Client) GetVideo(url string) (*Video, error) {
	return c.GetVideoContext(context.Background(), url)
}

// GetVideoContext fetches video metadata with a context
func (c *Client) GetVideoContext(ctx context.Context, url string) (*Video, error) {
	id, err := ExtractVideoID(url)
	if err != nil {
		return nil, fmt.Errorf("extractVideoID failed: %w", err)
	}

	return c.videoFromID(ctx, id)
}

func (c *Client) videoFromID(ctx context.Context, id string) (*Video, error) {
	c.assureClient()
	metaURL := "https://www.youtube.com/oembed?format=json&url=" + url.QueryEscape("https://www.youtube.com/watch?v="+id)
	var metadata chan []byte
	if c.client.Name == "TVHTML5" {
		metadata = make(chan []byte, 1)
		go func() { data, _ := c.httpGetBodyBytes(ctx, metaURL); metadata <- data }()
	}

	body, err := c.videoDataByInnertube(ctx, id)
	if err != nil {
		return nil, err
	}

	v := Video{
		ID: id,
	}

	if err := v.parseVideoInfo(body); err != nil {
		return nil, err
	}
	if v.Title == "" {
		var meta []byte
		var metaErr error
		if metadata != nil {
			select {
			case meta = <-metadata:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		} else {
			meta, metaErr = c.httpGetBodyBytes(ctx, metaURL)
		}
		if metaErr == nil {
			var info struct {
				Title  string `json:"title"`
				Author string `json:"author_name"`
				Cover  string `json:"thumbnail_url"`
			}
			if json.Unmarshal(meta, &info) == nil {
				v.Title = info.Title
				v.Author = info.Author
				if info.Cover != "" {
					v.Thumbnails = append(v.Thumbnails, Thumbnail{URL: info.Cover})
				}
			}
		}
		if v.Title == "" {
			v.Title = id
		}
	}
	return &v, nil
}

type innertubeRequest struct {
	VideoID         string            `json:"videoId,omitempty"`
	BrowseID        string            `json:"browseId,omitempty"`
	Continuation    string            `json:"continuation,omitempty"`
	Context         inntertubeContext `json:"context"`
	PlaybackContext *playbackContext  `json:"playbackContext,omitempty"`
	ContentCheckOK  bool              `json:"contentCheckOk,omitempty"`
	RacyCheckOk     bool              `json:"racyCheckOk,omitempty"`
	Params          string            `json:"params,omitempty"`
}

type playbackContext struct {
	ContentPlaybackContext contentPlaybackContext `json:"contentPlaybackContext"`
}

type contentPlaybackContext struct {
	SignatureTimestamp int    `json:"signatureTimestamp,omitempty"`
	HTML5Preference    string `json:"html5Preference"`
}

type inntertubeContext struct {
	Client innertubeClient `json:"client"`
}

type innertubeClient struct {
	DeviceMake        string `json:"deviceMake,omitempty"`
	OSName            string `json:"osName,omitempty"`
	OSVersion         string `json:"osVersion,omitempty"`
	HL                string `json:"hl"`
	GL                string `json:"gl"`
	ClientName        string `json:"clientName"`
	ClientVersion     string `json:"clientVersion"`
	AndroidSDKVersion int    `json:"androidSDKVersion,omitempty"`
	UserAgent         string `json:"userAgent,omitempty"`
	TimeZone          string `json:"timeZone"`
	UTCOffset         int    `json:"utcOffsetMinutes"`
	DeviceModel       string `json:"deviceModel,omitempty"`
	VisitorData       string `json:"visitorData,omitempty"`
}

// client info for the innertube API
type ClientInfo struct {
	DeviceMake, OSName, OSVersion string
	Name                          string
	Key                           string
	Version                       string
	UserAgent                     string
	AndroidVersion                int
	DeviceModel                   string
}

func (c *Client) videoDataByInnertube(ctx context.Context, id string) ([]byte, error) {
	data := innertubeRequest{
		VideoID:        id,
		Context:        prepareInnertubeContext(*c.client),
		ContentCheckOK: true,
		RacyCheckOk:    true,
		PlaybackContext: &playbackContext{
			ContentPlaybackContext: contentPlaybackContext{
				HTML5Preference: "HTML5_PREF_WANTS",
			},
		},
	}

	var visitor string
	var visitorErr, playerErr error
	var config playerConfig
	var workers sync.WaitGroup
	workers.Go(func() { visitor, visitorErr = c.getVisitorId(ctx) })
	if c.client.Name == "TVHTML5" {
		workers.Go(func() { config, playerErr = c.getPlayerConfig(ctx, id) })
	}
	workers.Wait()
	if visitorErr != nil {
		return nil, fmt.Errorf("获取账号会话失败：%w", visitorErr)
	}
	if playerErr != nil {
		return nil, fmt.Errorf("获取播放器脚本失败：%w", playerErr)
	}
	if c.client.Name == "TVHTML5" {
		match := regexp.MustCompile(`signatureTimestamp\s*[:=]\s*([0-9]+)`).FindSubmatch(config)
		if len(match) < 2 {
			return nil, ErrSignatureTimestampNotFound
		}
		data.PlaybackContext.ContentPlaybackContext.SignatureTimestamp, _ = strconv.Atoi(string(match[1]))
	}
	data.Context.Client.VisitorData = visitor
	if c.client.Name == "VISIONOS" {
		data.Context.Client.OSName = "visionOS"
		data.Context.Client.OSVersion = "26.5.23O471"
	}
	return c.httpPostBodyBytes(ctx, "https://www.youtube.com/youtubei/v1/player?key="+c.client.Key, data)
}

func randString(alphabet string, sz int) string {
	var buf strings.Builder
	buf.Grow(sz)
	for i := 0; i < sz; i++ {
		buf.WriteByte(alphabet[rand.Intn(len(alphabet))])
	}
	return buf.String()
}

func randomVisitorData(countryCode string) string {
	var pbE2 ProtoBuilder

	pbE2.String(2, "")
	pbE2.Varint(4, int64(rand.Intn(255)+1))

	var pbE ProtoBuilder
	pbE.String(1, countryCode)
	pbE.Bytes(2, pbE2.ToBytes())

	var pb ProtoBuilder
	pb.String(1, randString(ContentPlaybackNonceAlphabet, 11))
	pb.Varint(5, time.Now().Unix()-int64(rand.Intn(600000)))
	pb.Bytes(6, pbE.ToBytes())

	return pb.ToURLEncodedBase64()
}

func prepareInnertubeContext(clientInfo ClientInfo) inntertubeContext {
	return inntertubeContext{
		Client: innertubeClient{
			HL:         "en",
			GL:         "US",
			TimeZone:   "UTC",
			DeviceMake: clientInfo.DeviceMake, OSName: clientInfo.OSName, OSVersion: clientInfo.OSVersion,
			DeviceModel:       clientInfo.DeviceModel,
			ClientName:        clientInfo.Name,
			ClientVersion:     clientInfo.Version,
			AndroidSDKVersion: clientInfo.AndroidVersion,
			UserAgent:         clientInfo.UserAgent,
			VisitorData:       randomVisitorData("US"),
		},
	}
}

func (c *Client) GetStreamURL(video *Video, format *Format) (string, error) {
	return c.GetStreamURLContext(context.Background(), video, format)
}

// GetStreamURLContext returns the url for a specific format with a context
func (c *Client) GetStreamURLContext(ctx context.Context, video *Video, format *Format) (string, error) {
	if format == nil {
		return "", ErrNoFormat
	}

	c.assureClient()

	if format.URL != "" {
		if c.client.AndroidVersion > 0 {
			return format.URL, nil
		}

		return c.unThrottle(ctx, video.ID, format.URL)
	}

	// TODO: check rest of this function, is it redundant?

	cipher := format.Cipher
	if cipher == "" {
		return "", ErrCipherNotFound
	}

	uri, err := c.decipherURL(ctx, video.ID, cipher)
	if err != nil {
		return "", err
	}

	return uri, err
}

// httpDo sends an HTTP request and returns an HTTP response.
func (c *Client) httpDo(req *http.Request) (*http.Response, error) {
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	req.Header.Set("User-Agent", c.client.UserAgent)
	if c.Cookie != "" && (req.URL.Hostname() == "youtube.com" || strings.HasSuffix(req.URL.Hostname(), ".youtube.com")) {
		req.Header.Set("Cookie", c.Cookie)
		if req.URL.Scheme == "https" && strings.HasPrefix(req.URL.Path, "/youtubei/v1/") && c.client.Name == "TVHTML5" {
			if auth := sidAuthorization(c.Cookie, "https://www.youtube.com", time.Now(), c.userSessionID); auth != "" {
				req.Header.Set("Authorization", auth)
				req.Header.Set("X-Origin", "https://www.youtube.com")
				index := c.sessionIndex
				if index == "" {
					index = "0"
				}
				req.Header.Set("X-Goog-AuthUser", index)
				if c.delegatedSessionID != "" {
					req.Header.Set("X-Goog-PageId", c.delegatedSessionID)
				}
				req.Header.Set("X-Youtube-Bootstrap-Logged-In", "true")
			}
		}
	}
	req.Header.Set("Origin", "https://www.youtube.com")
	req.Header.Set("Sec-Fetch-Mode", "navigate")

	if len(c.consentID) == 0 {
		c.consentID = strconv.Itoa(rand.Intn(899) + 100) //nolint:gosec
	}

	req.AddCookie(&http.Cookie{
		Name:   "CONSENT",
		Value:  "YES+cb.20210328-17-p0.en+FX+" + c.consentID,
		Path:   "/",
		Domain: ".youtube.com",
	})

	res, err := client.Do(req)

	log := slog.With("method", req.Method, "url", req.URL)

	if err == nil && res.StatusCode != http.StatusOK {
		err = ErrUnexpectedStatusCode(res.StatusCode)
		res.Body.Close()
		res = nil
	}

	if err != nil {
		log.Debug("HTTP request failed", "error", err)
	} else {
		log.Debug("HTTP request succeeded", "status", res.Status)
	}

	return res, err
}

// httpGet does a HTTP GET request, checks the response to be a 200 OK and returns it
func (c *Client) httpGet(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpDo(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, ErrUnexpectedStatusCode(resp.StatusCode)
	}

	return resp, nil
}

// httpGetBodyBytes reads the whole HTTP body and returns it
func (c *Client) httpGetBodyBytes(ctx context.Context, url string) ([]byte, error) {
	resp, err := c.httpGet(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(io.LimitReader(resp.Body, 16<<20))
}

// httpPost does a HTTP POST request with a body, checks the response to be a 200 OK and returns it
func (c *Client) httpPost(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	clientID := "101"
	if c.client.Name == "TVHTML5" {
		clientID = "7"
	}
	req.Header.Set("X-Youtube-Client-Name", clientID)
	req.Header.Set("X-Youtube-Client-Version", c.client.Version)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	if xgoogvisitorid, err := c.getVisitorId(ctx); err != nil {
		return nil, err
	} else {
		req.Header.Set("x-goog-visitor-id", xgoogvisitorid)
	}

	resp, err := c.httpDo(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, ErrUnexpectedStatusCode(resp.StatusCode)
	}

	return resp, nil
}

var VisitorIdMaxAge = 10 * time.Hour

func (c *Client) getVisitorId(ctx context.Context) (string, error) {
	var err error
	if c.visitorId.value == "" || time.Since(c.visitorId.updated) > VisitorIdMaxAge {
		err = c.refreshVisitorId(ctx)
	}

	return c.visitorId.value, err
}

func (c *Client) refreshVisitorId(ctx context.Context) error {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.youtube.com", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return err
	}
	c.parseSessionConfig(string(data))
	if c.visitorId.value == "" {
		c.visitorId.value = randomVisitorData("US")
	}
	c.visitorId.updated = time.Now()
	return nil
}

func (c *Client) parseSessionConfig(page string) {
	for {
		_, rest, found := strings.Cut(page, "ytcfg.set(")
		if !found {
			return
		}
		page = rest
		var cfg struct {
			LoggedIn  *bool           `json:"LOGGED_IN"`
			Visitor   string          `json:"VISITOR_DATA"`
			User      string          `json:"USER_SESSION_ID"`
			Sync      string          `json:"DATASYNC_ID"`
			Delegated string          `json:"DELEGATED_SESSION_ID"`
			Index     json.RawMessage `json:"SESSION_INDEX"`
			Context   struct {
				Client struct {
					VisitorData string `json:"visitorData"`
				}
			} `json:"INNERTUBE_CONTEXT"`
		}
		if json.NewDecoder(strings.NewReader(rest)).Decode(&cfg) != nil {
			continue
		}
		if cfg.LoggedIn != nil {
			c.loggedIn = cfg.LoggedIn
		}
		visitor := cfg.Visitor
		if visitor == "" {
			visitor = cfg.Context.Client.VisitorData
		}
		if visitor != "" {
			c.visitorId.value, _ = url.PathUnescape(visitor)
		}
		if cfg.User != "" {
			c.userSessionID = cfg.User
		} else if cfg.Sync != "" {
			first, second, _ := strings.Cut(cfg.Sync, "||")
			if second != "" {
				c.userSessionID = second
				c.delegatedSessionID = first
			} else {
				c.userSessionID = first
			}
		}
		if cfg.Delegated != "" {
			c.delegatedSessionID = cfg.Delegated
		}
		if len(cfg.Index) > 0 {
			c.sessionIndex = strings.Trim(string(cfg.Index), "\"")
		}
	}
}

// httpPostBodyBytes reads the whole HTTP body and returns it
func (c *Client) httpPostBodyBytes(ctx context.Context, url string, body interface{}) ([]byte, error) {
	resp, err := c.httpPost(ctx, url, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(io.LimitReader(resp.Body, 16<<20))
}

func (c *Client) AccountState() (bool, bool) {
	if c.loggedIn == nil {
		return false, false
	}
	return *c.loggedIn, true
}

// Clone isolates mutable session and player caches between download jobs.
func (c *Client) Clone() *Client { copy := *c; return &copy }
