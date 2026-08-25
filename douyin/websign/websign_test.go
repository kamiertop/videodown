package websign

import (
	"net/url"
	"regexp"
	"testing"
)

const testUIFID = "474fc21d711fc1a0e6b55424939cc65669a09f1ec16e06f98afa0e254e71783d4061c07006b18b4bbf2f5608dc9387f3ab141319b89581392d61a56ddc983ef47490031ac453c0b341ca028da62761e6cb27ab1263dda39905a040d4b2e2077a1d52943f8d50af9aaddd8638f14592dbecb209f37abc972f5a37a3fb83bc1238f722846a3822dc02010003c5d9b34da8a637e626e4e3c72f6ec053c4ddb33dea"

func TestSign(t *testing.T) {
	result, err := Sign("https://www.douyin.com/aweme/v1/web/collects/list/?aid=6383", testUIFID)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(result.URL)
	if err != nil {
		t.Fatal(err)
	}
	signature := parsed.Query().Get("x-secsdk-web-signature")
	if !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(signature) {
		t.Fatalf("unexpected signature %q", signature)
	}
	if signature != result.Headers["x-secsdk-web-signature"] {
		t.Fatal("URL and header signatures differ")
	}
}
