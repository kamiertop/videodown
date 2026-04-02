package api

import (
	"net/url"
	"testing"

	"github.com/kamiertop/videodown/utils"
)

func Test_QRCode(t *testing.T) {
	data, err := NewTest().QRCode()
	utils.NoError(t, err)

	u, err := url.Parse(data.Url)

	utils.NoError(t, err)
	utils.Equal(t, "https", u.Scheme)
	utils.Equal(t, "account.bilibili.com", u.Host)
	utils.Equal(t, "/h5/account-h5/auth/scan-web", u.Path)
}
