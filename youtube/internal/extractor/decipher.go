// Adapted from github.com/kkdai/youtube, commit 3876742b95fd1cd96f836f396c2d54be8912a297.
// Copyright (c) 2015 Evan Lin. MIT license: see LICENSE and THIRD_PARTY_NOTICES.md.
package extractor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/dop251/goja"
)

func (c *Client) decipherURL(ctx context.Context, videoID string, cipher string) (string, error) {
	params, err := url.ParseQuery(cipher)
	if err != nil {
		return "", err
	}

	uri, err := url.Parse(params.Get("url"))
	if err != nil {
		return "", err
	}
	query := uri.Query()

	config, err := c.getPlayerConfig(ctx, videoID)
	if err != nil {
		return "", err
	}

	// decrypt s-parameter
	bs, err := config.decrypt(ctx, []byte(params.Get("s")))
	if err != nil {
		return "", err
	}
	query.Add(params.Get("sp"), string(bs))

	query, err = c.decryptNParam(ctx, config, query)
	if err != nil {
		return "", err
	}

	uri.RawQuery = query.Encode()

	return uri.String(), nil
}

// see https://github.com/kkdai/youtube/pull/244
func (c *Client) unThrottle(ctx context.Context, videoID string, urlString string) (string, error) {
	uri, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	if uri.Query().Get("n") == "" {
		return uri.String(), nil
	}
	config, err := c.getPlayerConfig(ctx, videoID)
	if err != nil {
		return "", err
	}
	query, err := c.decryptNParam(ctx, config, uri.Query())
	if err != nil {
		return "", err
	}

	uri.RawQuery = query.Encode()
	return uri.String(), nil
}

func (c *Client) decryptNParam(ctx context.Context, config playerConfig, query url.Values) (url.Values, error) {
	// decrypt n-parameter
	nSig := query.Get("n")
	log := slog.Default().With("n", nSig)

	if nSig != "" {
		nDecoded, err := config.decodeNsig(ctx, nSig)
		if err != nil {
			return nil, fmt.Errorf("unable to decode nSig: %w", err)
		}
		query.Set("n", nDecoded)
		log = log.With("decoded", nDecoded)
	}

	log.Debug("nParam")

	return query, nil
}

const (
	jsvarStr   = "[a-zA-Z_\\$][a-zA-Z_0-9]*"
	reverseStr = ":function\\(a\\)\\{" +
		"(?:return )?a\\.reverse\\(\\)" +
		"\\}"
	spliceStr = ":function\\(a,b\\)\\{" +
		"a\\.splice\\(0,b\\)" +
		"\\}"
	swapStr = ":function\\(a,b\\)\\{" +
		"var c=a\\[0\\];a\\[0\\]=a\\[b(?:%a\\.length)?\\];a\\[b(?:%a\\.length)?\\]=c(?:;return a)?" +
		"\\}"
)

var (
	nFunctionNameRegexp = regexp.MustCompile("\\.get\\(\"n\"\\)\\)&&\\(b=([a-zA-Z0-9$]{0,3})\\[(\\d+)\\](.+)\\|\\|([a-zA-Z0-9]{0,3})")
	actionsObjRegexp    = regexp.MustCompile(fmt.Sprintf(
		"var (%s)=\\{((?:(?:%s%s|%s%s|%s%s),?\\n?)+)\\};", jsvarStr, jsvarStr, swapStr, jsvarStr, spliceStr, jsvarStr, reverseStr))

	actionsFuncRegexp = regexp.MustCompile(fmt.Sprintf(
		"function(?: %s)?\\(a\\)\\{"+
			"a=a\\.split\\(\"\"\\);\\s*"+
			"((?:(?:a=)?%s\\.%s\\(a,\\d+\\);)+)"+
			"return a\\.join\\(\"\"\\)"+
			"\\}", jsvarStr, jsvarStr, jsvarStr))

	reverseRegexp = regexp.MustCompile(fmt.Sprintf("(?m)(?:^|,)(%s)%s", jsvarStr, reverseStr))
	spliceRegexp  = regexp.MustCompile(fmt.Sprintf("(?m)(?:^|,)(%s)%s", jsvarStr, spliceStr))
	swapRegexp    = regexp.MustCompile(fmt.Sprintf("(?m)(?:^|,)(%s)%s", jsvarStr, swapStr))
)

func (config playerConfig) decodeNsig(ctx context.Context, encoded string) (string, error) {
	return solvePlayerChallenge(ctx, config, "n", encoded)
}

func evalJavascript(jsFunction, arg string) (string, error) {
	const myName = "myFunction"

	vm := goja.New()
	timer := time.AfterFunc(5*time.Second, func() { vm.Interrupt("player script timed out") })
	defer timer.Stop()
	_, err := vm.RunString(myName + "=" + jsFunction)
	if err != nil {
		return "", err
	}

	callable, ok := goja.AssertFunction(vm.Get(myName))
	if !ok {
		return "", errors.New("invalid player function")
	}
	result, err := callable(goja.Undefined(), vm.ToValue(arg))
	if err != nil {
		return "", err
	}
	return result.String(), nil
}

func (config playerConfig) getNFunction() (string, error) {
	nameResult := nFunctionNameRegexp.FindSubmatch(config)
	if len(nameResult) == 0 {
		return "", errors.New("unable to extract n-function name")
	}

	var name string
	if idx, _ := strconv.Atoi(string(nameResult[2])); idx == 0 {
		name = string(nameResult[4])
	} else {
		name = string(nameResult[1])
	}

	return config.extraFunction(name)

}

func (config playerConfig) extraFunction(name string) (string, error) {
	// find the beginning of the function
	def := []byte(name + "=function(")
	start := bytes.Index(config, def)
	if start < 1 {
		return "", fmt.Errorf("unable to extract n-function body: looking for '%s'", def)
	}

	// start after the first curly bracket
	pos := start + bytes.IndexByte(config[start:], '{') + 1

	var strChar byte

	// find the bracket closing the function
	for brackets := 1; brackets > 0; pos++ {
		b := config[pos]
		switch b {
		case '{':
			if strChar == 0 {
				brackets++
			}
		case '}':
			if strChar == 0 {
				brackets--
			}
		case '`', '"', '\'':
			if config[pos-1] == '\\' && config[pos-2] != '\\' {
				continue
			}
			if strChar == 0 {
				strChar = b
			} else if strChar == b {
				strChar = 0
			}
		}
	}

	return string(config[start:pos]), nil
}

func (config playerConfig) decrypt(ctx context.Context, cyphertext []byte) ([]byte, error) {
	decoded, solveErr := solvePlayerChallenge(ctx, config, "sig", string(cyphertext))
	if solveErr == nil {
		return []byte(decoded), nil
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	operations, err := config.parseDecipherOps()
	if err != nil {
		return nil, err
	}

	// apply operations
	bs := []byte(cyphertext)
	for _, op := range operations {
		bs = op(bs)
	}

	return bs, nil
}

/*
parses decipher operations from https://youtube.com/s/player/4fbb4d5b/player_ias.vflset/en_US/base.js

var Mt={
splice:function(a,b){a.splice(0,b)},
reverse:function(a){a.reverse()},
EQ:function(a,b){var c=a[0];a[0]=a[b%a.length];a[b%a.length]=c}};

a=a.split("");
Mt.splice(a,3);
Mt.EQ(a,39);
Mt.splice(a,2);
Mt.EQ(a,1);
Mt.splice(a,1);
Mt.EQ(a,35);
Mt.EQ(a,51);
Mt.splice(a,2);
Mt.reverse(a,52);
return a.join("")
*/
func (config playerConfig) parseDecipherOps() (operations []DecipherOperation, err error) {
	objResult := actionsObjRegexp.FindSubmatch(config)
	funcResult := actionsFuncRegexp.FindSubmatch(config)
	if len(objResult) < 3 || len(funcResult) < 2 {
		return nil, fmt.Errorf("error parsing signature tokens (#obj=%d, #func=%d)", len(objResult), len(funcResult))
	}

	obj := objResult[1]
	objBody := objResult[2]
	funcBody := funcResult[1]

	var reverseKey, spliceKey, swapKey string

	if result := reverseRegexp.FindSubmatch(objBody); len(result) > 1 {
		reverseKey = string(result[1])
	}
	if result := spliceRegexp.FindSubmatch(objBody); len(result) > 1 {
		spliceKey = string(result[1])
	}
	if result := swapRegexp.FindSubmatch(objBody); len(result) > 1 {
		swapKey = string(result[1])
	}

	regex, err := regexp.Compile(fmt.Sprintf("(?:a=)?%s\\.(%s|%s|%s)\\(a,(\\d+)\\)", regexp.QuoteMeta(string(obj)), regexp.QuoteMeta(reverseKey), regexp.QuoteMeta(spliceKey), regexp.QuoteMeta(swapKey)))
	if err != nil {
		return nil, err
	}

	var ops []DecipherOperation
	for _, s := range regex.FindAllSubmatch(funcBody, -1) {
		switch string(s[1]) {
		case reverseKey:
			ops = append(ops, reverseFunc)
		case swapKey:
			arg, _ := strconv.Atoi(string(s[2]))
			ops = append(ops, newSwapFunc(arg))
		case spliceKey:
			arg, _ := strconv.Atoi(string(s[2]))
			ops = append(ops, newSpliceFunc(arg))
		}
	}
	return ops, nil
}
