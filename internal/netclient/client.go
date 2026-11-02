// Package netclient centralizes req clients and proxy validation.
package netclient

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

func NormalizeProxy(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", errors.New("代理地址格式无效")
	}
	u.Scheme = strings.ToLower(u.Scheme)
	switch u.Scheme {
	case "http", "https", "socks5", "socks5h":
	default:
		return "", errors.New("代理地址需以 http://、https:// 或 socks5:// 开头")
	}
	if u.Hostname() == "" || strings.HasSuffix(u.Host, ":") || u.Opaque != "" || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return "", errors.New("请输入完整的代理服务器地址，例如 http://127.0.0.1:7890")
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return "", errors.New("代理端口应在 1 到 65535 之间")
		}
	}
	if (u.Scheme == "socks5" || u.Scheme == "socks5h") && u.Port() == "" {
		return "", errors.New("SOCKS5 代理需要指定端口")
	}
	u.Path = ""
	return u.String(), nil
}

// New preserves streaming bodies. Callers own and close each response body.
// An empty proxy keeps req's environment proxy behavior. transport is a test seam.
func New(proxy string, timeout time.Duration, transport http.RoundTripper) *req.Client {
	c := req.C().SetTimeout(timeout).DisableAutoReadResponse().SetCookieJar(nil).DisableAutoDecode()
	c.SetMaxConnsPerHost(8)
	c.GetTransport().MaxIdleConnsPerHost = 8
	c.SetTLSHandshakeTimeout(15 * time.Second).SetResponseHeaderTimeout(30 * time.Second)
	if proxy != "" {
		c.SetProxyURL(proxy)
	}
	if transport != nil {
		c.GetTransport().WrapRoundTripFunc(func(_ http.RoundTripper) req.HttpRoundTripFunc { return transport.RoundTrip })
	}
	return c
}

const DefaultProxy = "http://127.0.0.1:7890"

// NewWithProxy uses an explicit switch: disabled means direct, including when
// environment proxy variables are present.
func NewWithProxy(proxy string, enabled bool, timeout time.Duration, transport http.RoundTripper) *req.Client {
	if enabled {
		if proxy == "" {
			proxy = DefaultProxy
		}
		return New(proxy, timeout, transport)
	}
	return New("", timeout, transport).SetProxy(nil)
}
