package extractor

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// YouTube's SID authorization protocol is described by yt-dlp's
// YoutubeBaseInfoExtractor._get_sid_authorization_header; see THIRD_PARTY_NOTICES.md.
func cookieValues(raw string) map[string]string {
	values := map[string]string{}
	for _, cookie := range (&http.Request{Header: http.Header{"Cookie": []string{raw}}}).Cookies() {
		values[cookie.Name] = cookie.Value
	}
	return values
}
func hasSID(raw string) bool {
	v := cookieValues(raw)
	return v["SAPISID"] != "" || v["__Secure-3PAPISID"] != "" || v["__Secure-1PAPISID"] != ""
}
func sidAuthorization(raw, origin string, now time.Time, userSession ...string) string {
	v := cookieValues(raw)
	sapisid := v["SAPISID"]
	if sapisid == "" {
		sapisid = v["__Secure-3PAPISID"]
	}
	var auth []string
	for _, item := range []struct{ scheme, value string }{{"SAPISIDHASH", sapisid}, {"SAPISID1PHASH", v["__Secure-1PAPISID"]}, {"SAPISID3PHASH", v["__Secure-3PAPISID"]}} {
		if item.value == "" {
			continue
		}
		timestamp := fmt.Sprint(now.Unix())
		message := timestamp + " " + item.value + " " + origin
		suffix := ""
		if len(userSession) > 0 && userSession[0] != "" {
			message = userSession[0] + " " + message
			suffix = "_u"
		}
		hash := sha1.Sum([]byte(message))
		auth = append(auth, fmt.Sprintf("%s %s_%x%s", item.scheme, timestamp, hash, suffix))
	}
	return strings.Join(auth, " ")
}
func (c *Client) UserAgent() string { c.assureClient(); return c.client.UserAgent }
