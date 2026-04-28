package api

import (
	"fmt"
	"math/rand"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	originURL = "https://www.douyin.com"
	referURL  = "https://www.douyin.com/"
)

const (
	Accept          = "Accept"
	AcceptEncoding  = "Accept-Encoding"
	AcceptLanguage  = "Accept-Language"
	Origin          = "Origin"
	Cookie          = "Cookie"
	Referer         = "Referer"
	Priority        = "Priority"
	SecCHUA         = "Sec-CH-UA"
	SecCHUAMobile   = "Sec-CH-UA-Mobile"
	SecCHUAPlatform = "Sec-CH-UA-Platform"
	SecChFetchDest  = "Sec-Fetch-Dest"
	SecChFetchMode  = "Sec-Fetch-Mode"
	SecChFetchSite  = "Sec-Fetch-Site"
	UserAgent       = "User-Agent"
	CacheControl    = "Cache-Control"
)

// 默认返回Windows的
func userAgent() string {
	switch runtime.GOOS {
	case "darwin":
		return "Mozilla/5.0 (Macintosh; Intel Mac  OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
	case "linux":
		return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
	case "windows":
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"
	default:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"
	}
}

func osName() string {
	goOS := runtime.GOOS

	return strings.ToUpper(goOS[:1]) + goOS[1:]
}

func (d *Douyin) publicQueryParams() (map[string]any, error) {
	webId, err := d.getWebId()
	if err != nil {
		return nil, fmt.Errorf("获取 webId 失败: %w", err)
	}
	msToken, err := d.getMsToken()
	if err != nil {
		return nil, fmt.Errorf("获取 msToken 失败: %w", err)
	}
	p, err := d.cookieQuery()
	if err != nil {
		return nil, fmt.Errorf("获取 cookie 参数失败: %w", err)
	}
	return map[string]any{
		"device_platform":     "webapp",
		"aid":                 6383,
		"channel":             "channel_pc_web",
		"pc_client_type":      1,
		"pc_libra_divert":     osName(),
		"update_version_code": 170400,
		"support_h265":        1,
		"support_dash":        1,
		"version_code":        "170400",
		"version_name":        "17.4.0",
		"cookie_enabled":      true,
		"screen_width":        2560,
		"screen_height":       1440,
		"browser_language":    "zh-CN",
		"browser_platform":    browserPlatform(),
		"browser_name":        "Chrome",
		"browser_version":     "147.0.0.0",
		"os_name":             osName(),
		"os_version":          "x86_64",
		"platform":            "PC",
		"downlink":            1.4,
		"device_memory":       32,
		"cpu_core_num":        runtime.NumCPU(),
		"browser_online":      true,
		"engine_name":         "Blink",
		"engine_version":      "147.0.0.0",
		"effective_type":      "4g",
		"round_trip_time":     100,
		"webid":               webId,
		"uifid":               p.uifid,
		"verifyFp":            p.verifyFp,
		"fp":                  p.fp,
		"msToken":             msToken,
	}, nil
}

func browserPlatform() string {
	switch runtime.GOOS {
	case "linux":
		return "Linux x86_64"
	case "windows":
		return "Win32"
	default:
		return "Windows NT 10.0"
	}
}

func (d *Douyin) publicHeaders() (map[string]string, error) {
	cookie, err := d.GetCookie()
	if err != nil {
		return nil, fmt.Errorf("获取 cookie 失败: %w", err)
	}
	return map[string]string{
		"Accept":             "application/json, text/plain, */*",
		"Accept-Language":    "zh-CN,zh;q=0.9",
		"Accept-Encoding":    "gzip, deflate, br, zstd",
		"Origin":             originURL,
		"Pragma":             "no-cache",
		"Priority":           "u=1, i",
		"Cache-Control":      "no-cache",
		"Referer":            referURL,
		"Sec-CH-UA":          `"Google Chrome";v="147", "Not:A-Brand";v="8", "Chromium";v="147"`,
		"Sec-CH-UA-Mobile":   "?0",
		"Sec-CH-UA-Platform": fmt.Sprintf(`"%s"`, osName()),
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-site",
		"User-Agent":         userAgent(),
		Cookie:               cookie,
	}, nil

}

// 下面的代码由Codex + https://github.com:JoeanAmier/TikTokDownloader 生成

const (
	uaKey          = "\x00\x01\x0e"
	endString      = "cus"
	defaultBrowser = "1536|742|1536|864|0|0|0|0|1536|864|1536|864|1536|742|24|24|Win32"
)

var (
	initReg = []uint32{
		1937774191,
		1226093241,
		388252375,
		3666478592,
		2842636476,
		372324522,
		3817729613,
		2969243214,
	}
	customAlphabet = map[string]string{
		"s0": "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=",
		"s1": "Dkdpgh4ZKsQB80/Mfvw36XI1R25+WUAlEi7NLboqYTOPuzmFjJnryx9HVGcaStCe=",
		"s2": "Dkdpgh4ZKsQB80/Mfvw36XI1R25-WUAlEi7NLboqYTOPuzmFjJnryx9HVGcaStCe=",
		"s3": "ckdp1h4ZKsUB80/Mfvw36XIgR25+WQAlEi7NLboqYTOPuzmFjJnryx9HVGDaStCe",
		"s4": "Dkdpgh2ZmsQB80/MfvV36XI1R45-WUAlEixNLwoqYTOPuzKFjJnry79HbGcaStCe",
	}
)

type ABogus struct {
	chunk       []byte
	size        int
	reg         []uint32
	uaCode      []int
	browser     string
	browserLen  int
	browserCode []int
	rng         *rand.Rand
}

func NewABogus(platform string) *ABogus {
	a := &ABogus{
		reg: initRegCopy(),
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	a.uaCode = a.generateUACode(userAgent())
	if platform == "" {
		a.browser = defaultBrowser
	} else {
		a.browser = generateBrowserInfo(platform, a.rng)
	}
	a.browserLen = len(a.browser)
	a.browserCode = charCodeAt(a.browser)
	return a
}

func initRegCopy() []uint32 {
	r := make([]uint32, len(initReg))
	copy(r, initReg)
	return r
}

func charCodeAt(s string) []int {
	out := make([]int, len(s))
	for i := range s {
		out[i] = int(s[i])
	}
	return out
}

func fromCharCode(values ...int) string {
	b := make([]byte, len(values))
	for i, v := range values {
		b[i] = byte(v)
	}
	return string(b)
}

func randomList(rng *rand.Rand, randomNum *float64, b, c, d, e, f, g int) []int {
	r := rng.Float64() * 10000
	if randomNum != nil {
		r = *randomNum
	}
	v1 := int(r) & 255
	v2 := int(r) >> 8
	return []int{
		v1&b | d,
		v1&c | e,
		v2&b | f,
		v2&c | g,
	}
}

func list1(rng *rand.Rand, randomNum *float64) []int {
	return randomList(rng, randomNum, 170, 85, 1, 2, 5, 45&170)
}

func list2(rng *rand.Rand, randomNum *float64) []int {
	return randomList(rng, randomNum, 170, 85, 1, 0, 0, 0)
}

func list3(rng *rand.Rand, randomNum *float64) []int {
	return randomList(rng, randomNum, 170, 85, 1, 0, 5, 0)
}

func (a *ABogus) generateString1(randomNum1, randomNum2, randomNum3 *float64) string {
	v := append(list1(a.rng, randomNum1), list2(a.rng, randomNum2)...)
	v = append(v, list3(a.rng, randomNum3)...)
	return fromCharCode(v...)
}

func (a *ABogus) generateString2(urlParams, method string, startTime, endTime int64) string {
	values := a.generateString2List(urlParams, method, startTime, endTime)
	check := endCheckNum(values)
	values = append(values, a.browserCode...)
	values = append(values, check)
	return rc4Encrypt(fromCharCode(values...), "y")
}

func (a *ABogus) generateString2List(urlParams, method string, startTime, endTime int64) []int {
	if startTime == 0 {
		startTime = time.Now().UnixMilli()
	}
	if endTime == 0 {
		endTime = startTime + int64(a.rng.Intn(5)+4)
	}
	paramsArray := sm3ToArray(sm3ToArrayString(urlParams + endString))
	methodArray := sm3ToArray(sm3ToArrayString(method + endString))
	return list4(
		int((endTime>>24)&255),
		paramsArray[21],
		a.uaCode[23],
		int((endTime>>16)&255),
		paramsArray[22],
		a.uaCode[24],
		int((endTime>>8)&255),
		int(endTime&255),
		int((startTime>>24)&255),
		int((startTime>>16)&255),
		int((startTime>>8)&255),
		int(startTime&255),
		methodArray[21],
		methodArray[22],
		int(endTime/256/256/256/256),
		int(startTime/256/256/256/256),
		a.browserLen,
	)
}

func list4(a, b, c, d, e, f, g, h, i, j, k, m, n, o, p, q, r int) []int {
	return []int{
		44, a, 0, 0, 0, 0, 24, b, n, 0, c, d, 0, 0, 0, 1, 0, 239, e, o, f, g,
		0, 0, 0, 0, h, 0, 0, 14, i, j, 0, k, m, 3, p, 1, q, 1, r, 0, 0, 0,
	}
}

func endCheckNum(values []int) int {
	n := 0
	for _, v := range values {
		n ^= v
	}
	return n
}

func (a *ABogus) generateUACode(userAgent string) []int {
	u := rc4Encrypt(userAgent, uaKey)
	u = generateResult(u, "s3")
	return a.sumString(u, 60)
}

func (a *ABogus) reset() {
	a.chunk = nil
	a.size = 0
	a.reg = initRegCopy()
}

func (a *ABogus) writeString(s string) {
	a.size = len(s)
	data := percentDecodeBytes(s)
	if len(data) <= 64 {
		a.chunk = data
		return
	}
	for len(data) > 64 {
		a.compress(data[:64])
		data = data[64:]
	}
	a.chunk = data
}

func percentDecodeBytes(s string) []byte {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) && isUpperHex(s[i+1]) && isUpperHex(s[i+2]) {
			v, err := strconv.ParseUint(s[i+1:i+3], 16, 8)
			if err == nil {
				out = append(out, byte(v))
				i += 2
				continue
			}
		}
		out = append(out, s[i])
	}
	return out
}

func isUpperHex(b byte) bool {
	return ('0' <= b && b <= '9') || ('A' <= b && b <= 'F')
}

func (a *ABogus) sumString(s string, length int) []int {
	a.reset()
	a.writeString(s)
	a.fill(length)
	a.compress(a.chunk)
	return regToArray(a.reg)
}

func (a *ABogus) fill(length int) {
	size := 8 * a.size
	a.chunk = append(a.chunk, 128)
	for len(a.chunk) < length {
		a.chunk = append(a.chunk, 0)
	}
	for i := 0; i < 4; i++ {
		a.chunk = append(a.chunk, byte((size>>(8*(3-i)))&255))
	}
}

func regToArray(reg []uint32) []int {
	out := make([]int, 32)
	for i := 0; i < 8; i++ {
		c := reg[i]
		out[4*i+3] = int(255 & c)
		c >>= 8
		out[4*i+2] = int(255 & c)
		c >>= 8
		out[4*i+1] = int(255 & c)
		c >>= 8
		out[4*i] = int(255 & c)
	}
	return out
}

func (a *ABogus) compress(block []byte) {
	f := generateF(block)
	i := initRegCopy()
	copy(i, a.reg)
	for o := 0; o < 64; o++ {
		c := de(i[0], 12) + i[4] + de(pe(o), uint(o))
		c &= 0xFFFFFFFF
		c = de(c, 7)
		s := (c ^ de(i[0], 12)) & 0xFFFFFFFF
		u := he(o, i[0], i[1], i[2])
		u = (u + i[3] + s + f[o+68]) & 0xFFFFFFFF
		b := ve(o, i[4], i[5], i[6])
		b = (b + i[7] + c + f[o]) & 0xFFFFFFFF
		i[3] = i[2]
		i[2] = de(i[1], 9)
		i[1] = i[0]
		i[0] = u
		i[7] = i[6]
		i[6] = de(i[5], 19)
		i[5] = i[4]
		i[4] = (b ^ de(b, 9) ^ de(b, 17)) & 0xFFFFFFFF
	}
	for l := 0; l < 8; l++ {
		a.reg[l] = (a.reg[l] ^ i[l]) & 0xFFFFFFFF
	}
}

func generateF(e []byte) []uint32 {
	r := make([]uint32, 132)
	for t := 0; t < 16; t++ {
		r[t] = uint32(e[4*t])<<24 | uint32(e[4*t+1])<<16 | uint32(e[4*t+2])<<8 | uint32(e[4*t+3])
		r[t] &= 0xFFFFFFFF
	}
	for n := 16; n < 68; n++ {
		a := r[n-16] ^ r[n-9] ^ de(r[n-3], 15)
		a = a ^ de(a, 15) ^ de(a, 23)
		r[n] = (a ^ de(r[n-13], 7) ^ r[n-6]) & 0xFFFFFFFF
	}
	for n := 68; n < 132; n++ {
		r[n] = (r[n-68] ^ r[n-64]) & 0xFFFFFFFF
	}
	return r
}

func de(e uint32, r uint) uint32 {
	r %= 32
	return ((e << r) & 0xFFFFFFFF) | (e >> (32 - r))
}

func pe(e int) uint32 {
	if e >= 0 && e < 16 {
		return 2043430169
	}
	return 2055708042
}

func he(e int, r, t, n uint32) uint32 {
	if e >= 0 && e < 16 {
		return (r ^ t ^ n) & 0xFFFFFFFF
	}
	return (r&t | r&n | t&n) & 0xFFFFFFFF
}

func ve(e int, r, t, n uint32) uint32 {
	if e >= 0 && e < 16 {
		return (r ^ t ^ n) & 0xFFFFFFFF
	}
	return (r&t | ^r&n) & 0xFFFFFFFF
}

func sm3ToArrayString(s string) []int {
	return sm3Digest([]byte(s))
}

func sm3ToArray(values []int) []int {
	data := make([]byte, len(values))
	for i, v := range values {
		data[i] = byte(v)
	}
	return sm3Digest(data)
}

func sm3Digest(data []byte) []int {
	msg := append([]byte{}, data...)
	bitLen := uint64(len(msg) * 8)
	msg = append(msg, 0x80)
	for len(msg)%64 != 56 {
		msg = append(msg, 0)
	}
	for i := 7; i >= 0; i-- {
		msg = append(msg, byte(bitLen>>(uint(i)*8)))
	}

	reg := initRegCopy()
	for i := 0; i < len(msg); i += 64 {
		compressSM3(reg, msg[i:i+64])
	}
	return regToArray(reg)
}

func compressSM3(reg []uint32, block []byte) {
	w := make([]uint32, 68)
	w1 := make([]uint32, 64)
	for i := 0; i < 16; i++ {
		w[i] = uint32(block[4*i])<<24 | uint32(block[4*i+1])<<16 | uint32(block[4*i+2])<<8 | uint32(block[4*i+3])
	}
	for i := 16; i < 68; i++ {
		x := w[i-16] ^ w[i-9] ^ de(w[i-3], 15)
		x = x ^ de(x, 15) ^ de(x, 23)
		w[i] = (x ^ de(w[i-13], 7) ^ w[i-6]) & 0xFFFFFFFF
	}
	for i := 0; i < 64; i++ {
		w1[i] = w[i] ^ w[i+4]
	}

	a, b, c, d := reg[0], reg[1], reg[2], reg[3]
	e, f, g, h := reg[4], reg[5], reg[6], reg[7]
	for j := 0; j < 64; j++ {
		ss1 := de((de(a, 12)+e+de(pe(j), uint(j)))&0xFFFFFFFF, 7)
		ss2 := ss1 ^ de(a, 12)
		tt1 := (he(j, a, b, c) + d + ss2 + w1[j]) & 0xFFFFFFFF
		tt2 := (ve(j, e, f, g) + h + ss1 + w[j]) & 0xFFFFFFFF
		d = c
		c = de(b, 9)
		b = a
		a = tt1
		h = g
		g = de(f, 19)
		f = e
		e = (tt2 ^ de(tt2, 9) ^ de(tt2, 17)) & 0xFFFFFFFF
	}
	reg[0] ^= a
	reg[1] ^= b
	reg[2] ^= c
	reg[3] ^= d
	reg[4] ^= e
	reg[5] ^= f
	reg[6] ^= g
	reg[7] ^= h
}

func generateBrowserInfo(platform string, rng *rand.Rand) string {
	innerWidth := rng.Intn(641) + 1280
	innerHeight := rng.Intn(361) + 720
	outerWidth := rng.Intn(1920-innerWidth+1) + innerWidth
	outerHeight := rng.Intn(1080-innerHeight+1) + innerHeight
	screenY := 0
	if rng.Intn(2) == 1 {
		screenY = 30
	}
	values := []string{
		strconv.Itoa(innerWidth),
		strconv.Itoa(innerHeight),
		strconv.Itoa(outerWidth),
		strconv.Itoa(outerHeight),
		"0",
		strconv.Itoa(screenY),
		"0",
		"0",
		strconv.Itoa(outerWidth),
		strconv.Itoa(outerHeight),
		strconv.Itoa(outerWidth),
		strconv.Itoa(outerHeight),
		strconv.Itoa(innerWidth),
		strconv.Itoa(innerHeight),
		"24",
		"24",
		platform,
	}
	return strings.Join(values, "|")
}

func rc4Encrypt(plaintext, key string) string {
	s := make([]int, 256)
	for i := 0; i < 256; i++ {
		s[i] = i
	}
	j := 0
	for i := 0; i < 256; i++ {
		j = (j + s[i] + int(key[i%len(key)])) % 256
		s[i], s[j] = s[j], s[i]
	}
	out := make([]byte, len(plaintext))
	i := 0
	j = 0
	for k := 0; k < len(plaintext); k++ {
		i = (i + 1) % 256
		j = (j + s[i]) % 256
		s[i], s[j] = s[j], s[i]
		t := (s[i] + s[j]) % 256
		out[k] = byte(s[t] ^ int(plaintext[k]))
	}
	return string(out)
}

func generateResult(s, alphabetName string) string {
	alphabet := customAlphabet[alphabetName]
	result := strings.Builder{}
	for i := 0; i < len(s); i += 3 {
		var n int
		switch {
		case i+2 < len(s):
			n = int(s[i])<<16 | int(s[i+1])<<8 | int(s[i+2])
		case i+1 < len(s):
			n = int(s[i])<<16 | int(s[i+1])<<8
		default:
			n = int(s[i]) << 16
		}
		shifts := []int{18, 12, 6, 0}
		masks := []int{0xFC0000, 0x03F000, 0x0FC0, 0x3F}
		for j := 0; j < 4; j++ {
			if shifts[j] == 6 && i+1 >= len(s) {
				break
			}
			if shifts[j] == 0 && i+2 >= len(s) {
				break
			}
			result.WriteByte(alphabet[(n&masks[j])>>shifts[j]])
		}
	}
	for result.Len()%4 != 0 {
		result.WriteByte('=')
	}
	return result.String()
}

func (a *ABogus) GetValue(urlParams any, method string, startTime, endTime int64, randomNum1, randomNum2, randomNum3 *float64) string {
	if method == "" {
		method = "GET"
	}
	var params string
	switch v := urlParams.(type) {
	case url.Values:
		params = v.Encode()
	case string:
		params = v
	default:
		params = fmt.Sprint(v)
	}
	string1 := a.generateString1(randomNum1, randomNum2, randomNum3)
	string2 := a.generateString2(params, method, startTime, endTime)
	return generateResult(string1+string2, "s4")
}

func GenerateABogus(urlParams string) string {
	return NewABogus("").GetValue(urlParams, "GET", 0, 0, nil, nil, nil)
}

func detailParams(detailID string) string {
	return url.Values{
		"device_platform":     {"webapp"},
		"aid":                 {"6383"},
		"channel":             {"channel_pc_web"},
		"update_version_code": {"170400"},
		"pc_client_type":      {"1"},
		"pc_libra_divert":     {osName()},
		"support_h265":        {"1"},
		"support_dash":        {"1"},
		"version_code":        {"190500"},
		"version_name":        {"19.5.0"},
		"cookie_enabled":      {"true"},
		"screen_width":        {"2560"},
		"screen_height":       {"1440"},
		"browser_language":    {"zh-CN"},
		"browser_platform":    {browserPlatform()},
		"browser_name":        {"Chrome"},
		"browser_version":     {"139.0.0.0"},
		"browser_online":      {"true"},
		"engine_name":         {"Blink"},
		"engine_version":      {"139.0.0.0"},
		"os_name":             {osName()},
		"os_version":          {osVersion()},
		"cpu_core_num":        {strconv.Itoa(runtime.NumCPU())},
		"device_memory":       {"8"},
		"platform":            {"PC"},
		"downlink":            {"10"},
		"effective_type":      {"4g"},
		"round_trip_time":     {"200"},
		"uifid":               {""},
		"msToken":             {""},
		"aweme_id":            {detailID},
	}.Encode()
}

func osVersion() string {
	switch runtime.GOOS {
	case "windows":
		return "10"
	case "linux":
		return "x86_64"
	default:
		return "10"
	}
}
