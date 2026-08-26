package utils

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"time"

	"github.com/imroc/req/v3"
	"golang.org/x/mod/semver"
)

const autoUpdateKey = "auto_update"

func (s *Settings) IsAutoUpdate() (bool, error) {
	key, err := s.GetKey(autoUpdateKey)
	if err != nil {
		return false, err
	}
	if key == "true" {
		return true, nil
	}

	return false, nil
}

func (s *Settings) SetAutoUpdate(enable bool) error {
	var value string
	if enable {
		value = "true"
	} else {
		value = "false"
	}

	return s.SetKey(autoUpdateKey, value)
}

// Version is set by build flags, e.g. go build -ldflags="-X utils.Version=1.0.0"
var Version string

// GetVersion returns the current version of the application.
func (s *Settings) GetVersion() string {
	return Version
}

type githubResponse struct {
	Url       string `json:"url"`
	AssetsUrl string `json:"assets_url"`
	UploadUrl string `json:"upload_url"`
	HtmlUrl   string `json:"html_url"`
	Id        int    `json:"id"`
	Author    struct {
		Login             string `json:"login"`
		Id                int    `json:"id"`
		NodeId            string `json:"node_id"`
		AvatarUrl         string `json:"avatar_url"`
		GravatarId        string `json:"gravatar_id"`
		Url               string `json:"url"`
		HtmlUrl           string `json:"html_url"`
		FollowersUrl      string `json:"followers_url"`
		FollowingUrl      string `json:"following_url"`
		GistsUrl          string `json:"gists_url"`
		StarredUrl        string `json:"starred_url"`
		SubscriptionsUrl  string `json:"subscriptions_url"`
		OrganizationsUrl  string `json:"organizations_url"`
		ReposUrl          string `json:"repos_url"`
		EventsUrl         string `json:"events_url"`
		ReceivedEventsUrl string `json:"received_events_url"`
		Type              string `json:"type"`
		UserViewType      string `json:"user_view_type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"author"`
	NodeId          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Immutable       bool      `json:"immutable"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []struct {
		Url      string `json:"url"`
		Id       int    `json:"id"`
		NodeId   string `json:"node_id"`
		Name     string `json:"name"`
		Label    string `json:"label"`
		Uploader struct {
			Login             string `json:"login"`
			Id                int    `json:"id"`
			NodeId            string `json:"node_id"`
			AvatarUrl         string `json:"avatar_url"`
			GravatarId        string `json:"gravatar_id"`
			Url               string `json:"url"`
			HtmlUrl           string `json:"html_url"`
			FollowersUrl      string `json:"followers_url"`
			FollowingUrl      string `json:"following_url"`
			GistsUrl          string `json:"gists_url"`
			StarredUrl        string `json:"starred_url"`
			SubscriptionsUrl  string `json:"subscriptions_url"`
			OrganizationsUrl  string `json:"organizations_url"`
			ReposUrl          string `json:"repos_url"`
			EventsUrl         string `json:"events_url"`
			ReceivedEventsUrl string `json:"received_events_url"`
			Type              string `json:"type"`
			UserViewType      string `json:"user_view_type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"uploader"`
		ContentType        string    `json:"content_type"`
		State              string    `json:"state"`
		Size               int       `json:"size"`
		Digest             string    `json:"digest"`
		DownloadCount      int       `json:"download_count"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		BrowserDownloadUrl string    `json:"browser_download_url"`
	} `json:"assets"`
	TarballUrl string `json:"tarball_url"`
	ZipballUrl string `json:"zipball_url"`
	Body       string `json:"body"`
}

func (s *Settings) NeedUpdate() (bool, error) {
	resp, err := req.Get("https://api.github.com/repos/kamiertop/videodown/releases/latest")
	if err != nil {
		return false, err
	}
	if resp.StatusCode == 403 {
		if err = resp.Into(
			&struct {
				Message          string `json:"message"`
				DocumentationUrl string `json:"documentation_url"`
			}{},
		); err != nil {
			return false, err
		}
		return false, errors.New("GitHub API rate limit exceeded, 考虑关闭/更换代理")
	}
	if resp.StatusCode != 200 {
		return false, errors.New("请求 GitHub API 失败")
	}
	var res githubResponse
	if err = resp.Into(&res); err != nil {
		return false, errors.New("解析 GitHub API 响应失败")
	}

	return semver.Compare(res.TagName, s.GetVersion()) > 0, nil
}

func ParseVersionByRedirect() (string, error) {
	resp, err := req.
		NewClient().
		SetRedirectPolicy(func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}).
		R().
		Get("https://github.com/kamiertop/videodown/releases/latest")
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 302 {
		return "", errors.New("访问GitHub最新版本页面失败，可能是网络问题或GitHub访问限制")
	}

	u, err := url.Parse(resp.GetHeader("Location"))
	if err != nil {
		return "", err
	}

	return path.Base(u.Path), nil
}

func ParseDownloadURL(version string) string {
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("https://github.com/kamiertop/videodown/releases/download/%s/videodown-windows-amd64.zip", version)
	case "darwin":
		return fmt.Sprintf("https://github.com/kamiertop/videodown/releases/download/%s/videodown-macos-universal.tar.gz", version)
	case "linux":
		return fmt.Sprintf("https://github.com/kamiertop/videodown/releases/download/%s/videodown-linux-amd64.tar.gz", version)
	default:
		return UnSupportedOS
	}
}

const UnSupportedOS = "不支持的操作系统或架构"
