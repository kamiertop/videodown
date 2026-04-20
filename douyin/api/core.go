package api

import (
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"time"

	"github.com/imroc/req/v3"

	"github.com/kamiertop/videodown/logger"
	"github.com/kamiertop/videodown/utils"
)

const (
	douyinCookieKey = "douyin_cookie"
)

type Douyin struct {
	logger   *logger.Logger
	client   *req.Client
	settings *utils.Settings
	webId    struct {
		value       string
		lastUpdated time.Time
	}
	msToken struct {
		value       string    // msToken的值
		lastUpdated time.Time // 最后更新时间
	}
}

func New(logger *logger.Logger, settings *utils.Settings) *Douyin {
	logger = logger.WithName("Douyin")
	return &Douyin{
		logger:   logger.WithName("Douyin"),
		client:   req.C().SetLogger(logger).EnableDebugLog().EnableAutoDecompress(),
		settings: settings,
	}
}

//func (d *Douyin) getCookie() (string, error) {
//	return d.getKey(douyinCookieKey)
//}

func (d *Douyin) setCookie(cookie string) error {
	return d.settings.SetKey(douyinCookieKey, cookie)
}

// 从cookie中获取UIFID
func (d *Douyin) getUiFid() (string, error) {
	return "", nil
}

// getMsToken 保存至 Douyin 中，每次登录都重新请求，然后每10分钟刷新一次
func (d *Douyin) getMsToken() (string, error) {
	// TODO 校验是否需要刷新

	const (
		letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
		length  = 156
	)

	var b strings.Builder
	b.Grow(length)
	for i := 0; i < length; i++ {
		b.WriteByte(letters[rand.Intn(len(letters))])
	}

	resp, err := d.client.R().
		SetQueryParam("ms_appid", "6383").
		SetQueryParam("msToken", b.String()).
		SetHeaders(map[string]string{
			"User-Agent":               userAgent(),
			"Sec-Fetch-Site":           "cross-site",
			"Sec-Fetch-Storage-Access": "active",
			"Accept":                   "*/*",
			"Accept-Language":          "zh-CN,zh;q=0.9",
			"Content-Type":             "application/json; charset=UTF-8",
			"Origin":                   "https://www.douyin.com",
			"Referer":                  "https://www.douyin.com/",
			"Priority":                 "u=1, i",
			"Sec-CH-UA":                `"Google Chrome";v="146", "Not:A-Brand";v="24", "Chromium";v="146"`,
			"Sec-CH-UA-Mobile":         "?0",
			"Sec-CH-UA-Platform":       fmt.Sprintf(`"%s"`, runtime.GOOS),
			"Sec-Fetch-Dest":           "empty",
			"Sec-Fetch-Mode":           "cors",
		}).
		SetBodyJsonMarshal(map[string]any{
			"version":       1,
			"magic":         538969122,
			"dataType":      8,
			"url":           0,
			"tspFromClient": time.Now().UnixMilli(),
			"strData":       "fvqyKRSbuoeaKcFCUg9scuLu9c7K/pL1fJn52asvf0AAtCJurfz+szdgQXo8SUHM18FA00xGyGFmK5pJU0F4wjvnjYB2kb6Aaw6i5d3es2OaSW7Ht8gXa/IRB/Rv6cpzsTQCsHp78p7PkW5tGC6xGhdHaErWovl6SaDhDJBq4F5xJ2kAWnUYXI5eg5y9tl71me/0jj/h7XWQaUlehHuNaHYgPQLyPNef4fU4mXRC8sE8stZ4lDq+IhpMhtA2VnhQ1Yr0mXMS87ARTah5bNWhRV0qatxpVoLD3C7STvygNcgmE5djBijEhTPuSN5emP25tb4Lc41/fGqvBimJFzimXew4QN++jDuN/qrmQHxL9NeIEoi/8D16DCt75lmiQ6+I2WHfXtUSYyT4iLD+xXohC+qHnVDtUWdZ26ToSvRzSk8IXVkMmpNVP0mOmvU5WUeDu3gkVpIyqqd/KH6DGhM+TCcn3y5efSJDWfaBA8R49GJZVVakr5aim6hECpwE0CttUbjhFQJ6tW5h+hNp0dhCGz3pcUrkzrb09EC/Yr9nVtxd28KXyJskvpOV709pYFDl8TnKpvo669wFmFChfot0UYDcgNCWEXf7+Xv6+ZuZrIaGkt0XCPSC9Lm8xqG0NeFgO4ji2wXMiDPM1cYtgrfkkXK6OBMHLZAqM7Kmcevz130ueGxBEPSUI+WCTudU14IkEyE6UxfLykyOie9GKwRWoihPmi57+DUTXTeRPnsNEbLkcC+YCp1dxWt/g7cUdm7SVhFyLopySFUkNxQAoPgq+Jd3Q4vtAoW6R7R2moQ9WQkXJxPBK6kOeQpZ8nCOlECkv537Qh7R+cfOUIULy3NwqdeMsoQboWVmY/VHB+hw6syZ2pxYGG+cOpljo76GgEGsKWK5cp2o8Pm6/b6EY9otYvzrhHq3pSUkLOPYVbWmrXtzvW6Tk2tLo+yuALQAqMPaL+xWrIZOf+pbKMSXPGKSMbOTBGz2AKkpToCUMk5Cycq4X6u4mr5ZZXNa88xMEVt40TiLp13IeMNmQZW443PhssVIA+1k2ABzTpyYWqSMrKRQG+USrZkLT0Ad3NDEwNbO0VcUzKYF2IVU+vTMP/HZfaUwOyZNDhFm6e/NCQNnD+Sy/8TXtkFkNQxGujzeynDfCXptZgAHIvTGNmKkccDZf/ofWjeipT8st0/QQREA1cIoQ6dcIJerUa0+sbzIDTZKJ7/zspf25VmxBgWlBYT4VMNFV7bacnnRLo00if5uZso+4qG7HPSKlyRz/MCa2FN2ThoXc2UrHlOnRCMPPJc5DKg74+nA7f3+Fhg7cwldSceJMLktO1DGy3ujJVTdzC8xRN3swHMHGxlDjXhj2Vm2a1aBYScbVicgNeuLsUmMVgRHUk/aqy7fDrqBi/sCNM3oI8rVG+DgRlFrRTviwDei/ee9acu2vGXQCR28tQpjkmoaos+zmfws2UyVH5QCTBVRiksrjQGbIl9qEYagrPKNQvHi+HAU0S/qZNLAsXoqtIam/gu2xLl1zX1JHxpCeERMp4v1gH9K/JEJusu3/8oBYy22+qo+ISMsjDejEEzW6jpOb0V1AjXciF/yvs8mke9+2nmzvMu/2xC713sEgB97KwP9vjhPUb7sVNw2GsIon8MnLTwvEaGwxQ3sgntfLu6TiX4wdu4WAlA/aSezNVeScYOOaneJqCSKs5CKzkPk2RwoXtEbJe0VDDetycAHhW3rArLlH4lWK/tBSwPKDLo1+DTlcYeIhFx1sCHDkUpGP9zhImQysXvaKnbTblHTCzTaXDEbX2HDMRZqUuCjj/qdOC4KGohpR9QdC8RKNmFQXQA6S8WSq7tSvV8Xx5YOCuY52c0V5hpLm1Dey7BIpzUbl4oFU+S4dUIqmgTgL2QphM9GsSXaencM2LJRm/6NeiiYl7sC0NB9udlAYrvRvkCtoxV2YR+5VJU+gVo5EAY1YzDI2tP00cIRnQMhUzRUjj9FroS6oQoCbkLsMQ8b3/Ab54ftBHCSF3u1ZX4rbAl1L3WeZrfvsb+FY74JfRQPHzFfuLxnIZq6LPE5c0j5UppkVqsdfos25vAaOZ25IFSfjLLyj/eHkn0qmP3GXiUBXft6Z+YntVzqMgYxDGYJ5H++Ttx3sn9BfDYLZ0DubqKWaofSKmQiSZMwwu1Yna1M8rNuj4SpHBXNR3SghtVHBoY1klkCx34IBymVAghV5pyXbXEdlukcyOOVXFOyFqsg0wOYYaEWp0GG84kQV8CQ1b3JsMMefcUtnGJyt0+A+o3+Ej48KwoTuL5wgwWPZwRqo2CUbs0oUr/cNqi75la6+F5MEHtk7JCVJByDIZ8/HMIwQysf9gQBTmYnoFIEf4p3Keqb9gvFJGMqos1smI6QWWOT7f9q3lIA/ZeLnEqcm1jcc+DbR3/4vpPJKRJuTV5saXoRMdpFj8QlLGRTwhieUqHKbBYXMypfGAXncWWc/PIQKfYvBpv1xTzozPuHkehDX4CpYlv2ag8jz2k5ZnAtckstspJJdUGh8z0ItRFweTyNi2YOgoGrkb1r3Pj6bWt5CFYaM6d1n7T6GlKvmIX9CETKSK1gxQATaY2hUvm46qNNGkbZnrq1jG8qgz/MKZQITHVyq6xwvZZEzlF/3Djh4aCMkY4bKFuTjhjXYSPsCnNOdaMsUpVytAwKHdR0/xZ48IXwZ+IAzN6/P+eUpDVC8cSZJ5GOmF2LeaRk8dY5tLU3BjIEmk0jy626xbLgNts1fPfhbD2Kt0xsCApJKp7mdvTC43f/DP2UhCpLKrbB43gdE2k7dCLgCYE0ra5pWOsf8Aa/iqgHboT9HJbkKV7u0L2FSZeGdEa7YQvvJsvsh6AvzarsQe5Mkbi061pQr8+juCSutbCVJcivvQNh6+W3SOkrQLQlYvarEO3UUujsY4otRuPqOgGEGVuhYCfBIKLo+cKG4YwTG+8oqXiOL8xQ2xuorOh6iriDGaDSLMsQF24Vdh790oZvqOM7NBB8Uo6/wJ3e+r6RM8tzePznqEgAtfIKr6IfnW0D0/zVChwKVpFiNV+/mn8GwoyKK4IrbTGgmhuENjFfacCULAX3f+BnPUzXBNNJeyAVjx833gBKQT+YqfgTD2nlL6wGod+WDvGdZPMIs8d+qmlep4q/7HRGiNmbP3qX+llleg08KhxaVyywV8ngzMBPvdlwH3JWIukB2nOCrUC4JItVVi2Hp2JevfJ1IAz6DCApCkSWMvOFU2yUP2nnS/JhJjG9tjRZrEib/ZPh3nJFTmXw7or/L+cDLynXMAsaRVG4fYIUHLlIx8ky52UDXPmFdCQ0So8AvI7rz+rDZrvy0xjrItSL4bCHUxovToNLRwydWlWnDI1Wb0xwChw4SqN+ddnO75ByCP3meK8ZlY19GrQsyV6FawO5EEUX3XuoZeLsuyiCGLNAVj0oKcZBVu8PwimyNJr9bsU8ZOTwienOWrXDQsxhs2nmVoF9Dj1k1XSogc59IiYZ15Ebk3nHyK+2jGVmWYXg5s9SsSdG7Ub874qc5+smNykw5DloRhSugitRSImZdIHZlaA/FrZR5pti7cxXsxRrAoz+E1vSQOO2BdN5SjO/4nAV0vC3x6n5F8q6UyyjfIjweN/ehJyX6ZZB686x43VkpdUdVX13BzezAOvofAO+dC7t2yAN31JgC3nZ/ZLInWNMnLjHZuJruGy1cNarfj/kYcRJvnwJjo9WvvVvUI6RCh6if8IZd6sCSe0O/Zqq5VhhkN7/Ojch9w4tUXSp1NplqM6eKz0sX727JTjXMt1ev80Y6LU9slKDWu2sccjDZse8IDc6LCZRCXOLiwp4T9hHosza6HB5FM9KyDqePjTn+8i39hdeLBvwP+Ez8LUe6lFIeRDP/Hdyp3V4blz4J+qOvToM/3Z+XmBFQkaVFnnHUVjtUa9RBkqPbgaqPZk0VfuHRMIUsJMAL2mMCDL/2tq0gC4CX7aOf6p/lAHag3zW5FGPFLhnGoyTFZwptmh/evJ86p03zE3R3NoK8Z1OAvZDfejzFpJdeRDnGzt6M3vRve1rz28bibEOT72+8H+5wC7zV4msqXFkGd77XTqNY4uvGxsIpC/UqVnWqx8jxcwTIlBY77RaAI8eRpgnnJrVgt9ZhkIYYuf6lYOHCa9efjJp+9nASR2NYKcUItzDtSbETcs9XILQIJ5UVMY8kYcbyf+GsQrigfVMVsmj1OHe4eV/GZ55UCKkydYgJn6CY1SgsW4Y9I0u2F1EnthdCMsXxMVoIP3uuTjpWvAFkvGJSNFtlpp4O936OGYsg5KnQc0OYZDndePi/3BIV1PxDjV5Ijhtn72ayaOGjpwVLLKzPK6FkxZL/PjoMuF1ArhMxPeJVRojS5Lo5Xl3uOeVz32OXmXE3U12zEI7CJsVacTYIoP3hXqf9NROg5wHHCa8c8bb0F3XPfqSmGcL4+oaDGl/cJeXiqjH9r53Mw3+SRtAMXgxoBmV5YmowIdOz8YDJ2Q6nyGnRJ2xECSmcREpFwrulcW80I+MTHMu5It/HyhpJJwsntCv0QUJVWWVEUAq9e6IEL/e1fMJwUDbtj35UgN7Fge4I6vIppclA2vtqbXiDTjjx+zD0+T7YtsCb+wlGe+fDbNWIqxGlPCp4pSJC/9Cac7n6MWdZR7FbwrzB/LXR1rRXWQ9+MDwGNohxSZPsKRuPcvp3YtZdXqXyXZf1dM6Q4QnuRGSPk74ZI3xgbAe9+frLMNinSlHUEtBa1U+KmILZIofXjgFKeVFgrTIDlzvD8d6LVjgBmWZv1g8YFZxf4PmcoyRYMeW872LnOzhLD+2DtGEQhNw27nvTlubXuuwgVFK6AQfd62LD9a066BPJltMbN7qvkvVCnaE8XE7jhe1t+8KejAStYzosyaaGTR4AzeA/+EHh0rxgvwTXORQ79CUw0l9a4oOfgy22tFgrqToT/nyynjglmitnN/tXeDGSjdPS1ma1GjMwnkokRv3uH2u7aR+r9rJi2t3P7+wBL0+WSJECuw8rZbQeCp7kNlnNk5WguTX34rZV5lnLROm/3LioigBS71gCCjtrmC2b3BRL2vJDJDDZ628Gl+zpNuZ9o+NS6zs1dDJWtyI+HgT9IMJ6MV8pEcN/bhXohm+KDvueO7kMtTz1KjUTEFDV3Ks5pbQwjU2EMrpTqQq9QB+SZKBAEhgqwcQoFDvcwNUdtWQKBlisfJbg91r9DF6b06PVnTM0ToXqEpvmqw1HZF9oHSi7b/blbS3Ur8TD7T9Ysx0YVDzul8AE/3hDVfwQr8jqaIYXMTOMmrzmmHSasfNeosQYb11qiP6CfzSgNIKOmxqJ51XR17JhFAUpyqs8CSNVvDlyE7Ux+cmQf4JUk11CUjuWUFW4vp+cQ9f1c47X7FeJRxV0cIf8KrUICFoY4tjTpD3EcIcWLZjusACnS8SgmgCSi3g5fHEm4mSxGFRQJzDIpORiMHYBzxmtpa6QLFQuv7kELHk/FfBLX3bZb22Ym3ADHUY9pj0QeuI3sNibqWLx+5U+RXBbMsx9iABjslWwCL+L6Zqt2NYM6H65oNDISlKLoQYBbT5PiKDWyUXJKsy8AezIVaws6hL2tc8rfmyIsCum8ckuWVZXxb/bPrB0pswtaNbOeMIyHRAurXU0NbC/TT3SCuG2j/ToSjTtzWc4/HwwbCUO/nfJBzdRDwmyYj87uz/jwuHON2BcmqZ4LplMPEl8b/BaQ0uN3tJuEdKO8pi9JAP3Ip74KqV37oT7GClKw4T4xcHLsW/SaLKHF3d/Pi8DuxqP6nud1hZc+mH4zrTgnfH5/hMnCcPz66+MaTo075LT9eJYKDyKOkl4GFf8/NgQHO47fhpMGJRa7HNOrpqLfGPLKetLm1gyXRUJx946fkctxNRQsf1r2HpYuczzG0BLOpV87r4IFhvHJuxCC7Fk/mXRRM45CN6WmLSJNhfSPDN/7iGoKekNJaQb3gbauJOwPuHge/+NfxJiFJAf9ft709zG65B7teGv7eKLaeN7JvndNOdOIQWNyWcVnlRB3atHjMMgMDMad8cPdHACCEPCowbmFHiapiX6HXGauHAyIb1q2lKzPK4U7uez0ZIqyVHA2SC1/PZ2tE74cF0DlYSDqBSEbwz4a9/eWVbupnhuPIGzaMXtuzqxCxMc0qvGzZOo05Ei98Z13ZxdixIpGkDZCWvKt0CwQLgYUAjz04m8MWlA6oLUtCc87A3mgJe4lD6T0RBDzKr2Lr5XqWLCTMc7f+q7yLzDWwWmFpIY+0Q4mC3KkXDGSqZFZjVFgCmlMpqcDpg7qqrOo5oCLBT+duGdbf+KMKgEqbatbzAL20FEGBZhma9QpZ8ajMxMdCNcC1A0kDBWZxACGBXiQjIhVE7LUBOLLPsi13R5J6zgfQETBZjynJNSzB9bU4vyubGmcYqBS7aW40yCUZ08iycie148PtL6p1b6budGG2T3ghXr2IBgotocp9kC3NFjXwZX/tYobWGA/5fSIHdMSi/StNnLLgtd/RZKynWuIs5lHcTkn3XJk4b1Tqjb5eI+edRWTu3aFFyYBdmoNTCRVqHs0tkoiQE9mVO8Mw6FL4kx1bI8hC+ZQh3cnjtSFyBN8E37kztRNDQgo5g4jWE4FPB5vof/ivrvJ0+Py0inD/6YdyOEUVJBw0ioisLiGmZQSjGwngmrRzBBdG6EoViLAT5H7Gqpv+aFa5GxldoJ8mnBfiL4Q3h/xzT3TMBYNPS7MrxZdAjwfp87OnMQo/hvpUwqC89uvLzduICaasIUYcYBiiCN7B4SiAHbmx45hTFOWBw+ti3lmEotqhRxLmAzv092fePzn34uhgZmvoFkUq4BlikNm3HEWhVBZ9zPwmaOeMN9WUzkEOE1HXa48pSZMTt7jfjL9dUoCRZ7lcX8DfPnMKklEPmTIub9AsLjuITrtoeHa83sWwo3yHqtqTPHtuRGa9+nAmv0jOeGl6QZ+8bXCH/JgECz5iBfDq+WGTozdW0zAuUzQMalQSOHtRbJQiEVZfkOXbbMlRFStB+57z6EAzQIGwMSYn2TGyVn/PTCw73Rzl73G4tPHYVBGey8x06Ddq0vAzvIcSBD4t4SwtUde4rOVdJDprDy/1/pZeFpnRfLqy1qv4eiAvR+B7b/AqzD2L5MihiVck+edCwARXCNpnvmbd50glyxF3kSV/yDr62FEldZvuwQsjCPE4udQ8ZlD3dULr3QeKadFawGBA4zNbQWeydWFxWxpXohSYqhWFvEZfgsHWRkei33Uv8l1WSdNvYnxR/GGnr5QrEXy8CuJpSXNdiaAuj+IgpUkUc0iH6aBpbMkSQTW8XpRRdW78RedqjG/XTrScK9i9LlwVE9ZLrvoZwau10vLBAO5YHlsjsq30dM02aGmzgkTrqwYeCwwbUrKo1yWD2sP/3867Rrq2R5nQenjb9o70jZEX5oxNsELLr5ojGLEksxaXKXzpApH/c7jYwHGkHrRy6tV0RCADhItMW+w4PN3fX5Pxkj/43Ga0pwrSGIGjfqzFleXsZezHLjVp962UsxxJCM6/KyHZ4pqdvDIuDeNMOpfdrpRai55abiQASazpZ4CSkIpE+mVLotQt86L8gaa4Mu5ODbMRyqVFpVD/AgaeU/qU3qQ7loNJyoNd6SS0W1KgAZY7xTHqphiGeFWwtHMdSrzUI33N3uZw3zupqtfEp28cCkiLLIxK10807JBPkHY7IejM3X3f28Qf1MDRXT+CVUs0K0OeVx8pYJ1svztMGBonlzgjI73T/SFIdwqjLxDy54BknjYO9NjYt7DgC3lmweHENJzn2Nl8jvwjXVl5bdeMOp1/87nWz7qJnmBpxyvMCcriDU8qlNwkpFS/C1Cex2TGktHXuVIfP2i+/WI34S5FBgIFw0mSQ67kydGcnwMXV37C8vov/wqdtK0IDUzhXrwSvimpesyt/6TqLVl+fxe6s9CDqJ9KWgwbBvoBmKU9AjHUbOqtLUtVBO7cA3WFp3G/LlU+WTApyOjHANW8ljnCy+szyyrZjk0QQQAWpju3JSiyRYoeaolcDcv2LILptA8VVACp7WsOEbgr9+n5gDoM33ghwsMGgi1GCXfRcbTZAd0Z1ehkcIaVk8FcgYWU0oVVKm6alm5m0zl+eYdbN6KJI/UOKbf6827HFmTRjWjxnuwFwXmz3Ramv6jdvOJAQqhrLbisrnrCtGc1IaYCZhBhHcKrc5BYQVGboSPpmvlTOzV3pdAMzsez7Evd2CjNH8mYseisOKYHSnj8hn+JIQdeP+CZl08MOwxDLdfZpUJ441k24RyyON6ai5U0LgY/PVv+36p8r4oVfj72/NB6yi3Pa6pSP760RQxVUOL287vtZd/kTQHDGHieUfPOuAIYREO3rLmC51XCzrnxZ8q7jtQOyVlIPq7gso5fFMFx53slW4iqDaGhwKySUhlhvzSuZCl0Ry9muUKIF2gb0SnzhHC294A2BB/B6HHRoEiOMDm96TIYsVGTrV7SRvVg1vSnokOYBRHtx/bdlF67Jemf5mTV+fYlNxYprbFe10mwTQ6kQ3SWSpJH17Skat5NG9HgXloxR2EFQHmQn8nd7cWBrArBsP8MfaowIisJxAJSeHAeiec7gZy8H0cvg4VLNCZlPkwUdg3pJkcottXAhJepAWAx9jOuKrL64HP3UATpr8+4SGKPzHUZ5w6Em3ZBYELDoW5EFTXFQFrhBUsEHslURgbmTiOLVZNwdWnQK/Tz/1ShmPbFFsUlQyXx9CWHXyaVAlSm5z0elMJggO6EtYrP1/v1j0LVRgyMjE/RTiV/wNe/HXCB4ZtNELmLEmglDFQzLoOzHHyWonLxDcq/lATyExF2NN78wjaBKIR44SxkH0SxgE1QlC/JaFpLvK30kYuQOQkC4oP8qo6s7KQsaKpXZYzw6K8TPQ3Z9aCaExB8MfryDFpd0BHk+YFSHTuwXOayMO4pPLbsvXA+exfzV0BGOCodwB32RvtPYPRrKJAxO2xPxuUV7idF9NSc0w6jL+1Mps8EP6MCxdK20dbtaTNWZBwl0YNoodpBqvVM3ICKnJc6FExv7lUw2rClHNe2jCSn4wHRu68T/0yNsJa1KcDk9tuJtm2ZSN6bs3lXH4XxIN3yLgKA2UfgA6pc6x2g9zNXqXuxlBbttGaod3wFMswDofbIdomu8FjYSpG+ALuplBJPOuB4n3x7ZidqNUyhqf+l1+3eAUPpq08tlwL+wDYxNx8WeHaPRK7QdRZW+kPdOHyd3TrdCV89aeP3RBwcRCHmsflbSJffPhaWNziMN04ad3Vbqx0WNYTpJjzMt2GMa5Ql+vYlDd96C628GVqYwy3r7nFNkl/E3LeXME5T9N1St8cTJjEthhoa7C73XB4Sbm7nGav8GJovGGkCb7vdOb5DfIMMiMVMOEFCjJtmTJKp84xhgIsY+yJTbZUMurTgqCQ5lqCGasvTPUQf4wM/ZyCLYU6ZGnoYRqvIKX8+//imYppEJtPtwXKrMNwYDc9f3FZk08G6lj8JurDuDuDR9z5byJd53KFTI11q3Vc2Cdwg4w+NnKSHBPrYA/ar8y3sNSmJjmfCiuLqjnuxH+L3uQjv22GoGVOol4kvtNoIgMDRIDGu//EMlEMQi84e9bgyUi3jP0N8h+7nDsS9nm7bFRdXFnWui90Qfes7TT0tD7sY0XMGjQky4+dXCoxYKnuEASCUXu6KG2EtTdlg/Dq4WToh3vl/6YvHmU8fHH+w2i4AMhOp+5gnqXI0sVc5idCXuKb0wUp3dqUgLNuOG2wXrxhCkAWBc191f9uQaCJ6qWCGbMz6HoTZU28vsQPpmlEOIpaw7Drz/Ug7tW6zEjBFjz/9DTnJju62dX4CR9/WRmQAJ0f80LHSf0QTD4tYUzxUeQ2AT19CZL9jD/mzd5GcQXAoWbSUnCX3PtxbY5btqNARVbX6fB91QYwdrbLftJOKsaivFbgn28nbE7n832DNzJnSGnHP6rixv0k4zjZAMprEMJP1XJn0zDIIOhgbuv3HsRxded/SThfHIjM2W6/2DGzv8nU9deOTWI4vihx0vrWyqtZ4ikElBdndbga4kmi9BVh/BrLIOR+0+RJMhhsaMXvbS4R9NZHW7Ybo7ots7WPoCMW/Ti7lVzo3Tk0YmX2FD/OmmLUKISRRq7GlnsSC7bDv4PeOuEnzL51nnx0oWgATr7YiGFcl+70GyyFWyNG0XR7h+k7gQIzQ8CacvKZD8Bi0AYc7RI+AYmbfTSJ+nb/efb3LE911M8ahM1CkswxVtL67vS5R1XONIquYCUCvkI9N52qzUjpOl+///c7F8xWs3fDhMsz5Z+cBEOcQM01SuF83Nq2o7jacaQP4stsQqVR1V2uM+1BzLk//hqH5oG+nxEySQU+Qw+SVYUnG5Vk3EVV4b33l2ZrIR38NnXtiUOtq+rSTNjwRnp9Ow/seqTyn9ehc3zwiNHCtSD3k5DxacUC8EOPLlSG4rqBXSkzgfkUvy+rFaKxS8nI8No+z4aj1Lw8dd0040xCYXxWBOgv6RpS41IhmlMop2/NTlDkr21W8Vd4g7nEGwYBJVUDwksFFpwqGMbzJzCY0GfmV19VMTkBT4EKY7TArK8gnU5mp5BacJCsD5W0fnmVjsrklXzzqkMLYvPZyvd2HuXN9cuWdmEkg+7SfsuChBVJrkZSOEHZy54qlMMVt2Jw7AkBOeni58UogJcRxTHN7njH9aDNMAH4BI3Q//rNdoPvcPoPF74Q5Khusmrfhmrg4MEGNSkUcwshQHn2O4TSzWPFIoMSZhPCo+488gug44DVNui1NWDQNQIWL1wkpMP82bBc4WnpthQXJdwhGDu3onBYphLhAkZUCxxFMZNNud0c/af/d0cedS16o96ouQjD3pqzEeLvA+rHIg0WY4z/Rrvq2vdtLFgzp8BKNWFtcaFuvk2KDCEYMtlSe6v8oJVSzv+2VFQw+gCnKLtnq4+MnZURinfS6gDhUNRB2MX0hcm/FTn85bKh15zQjJ7DEi2LDkO13ehARgF5G3rTA49WhcfXkewDQRZYDl/RhoUiFwMgcj/bznJjODcXF7c7BI3qdhf3iNmIAruIl8qz7kbzl8VHPwmiaUkm2yDbV3tFQJ5fdH9FeuNd/IbZZSwZSxJ9AwvqPA9IZiZUYiOVANYwtYDpKjiig56wYtTcBY2uB0o3kckryGBN6WX7IplfEoSNIAlfWwF6SJjr/PEJM5/9BkPnsKt3zcdmEBuVzAhTe6yPdYrnOu5m5kcDsFKS6Ej5XxiL9V81gk7nWqqQMmkiajmFPaQNM6RszF9VzWZzBCpFTQ9lr9bPAa5os5ydgPqE8yumKu12YMm5jnDAqjnhDq6voPelUBX3SosDTFbESgwx8ZMzGgl5wqjiyKK3+90+0M5+1sjaxzTw9e7pf3OUhYGvu3DR0f1ZXMpbOqF3xcY8wuXD+J6lvio3ACpfFNQHEvxsXnURR+txaUK5Gp6rxjW8xVr1GunakZVmsWa5+LxTlTIp/qh+YfjB/7iXsBIk06V7Td5bS2hsMTLJx39OKb8rwpwtKBD3Ulm62tGkHSicIw8SGhIFiqu5dOr12XSRFpiwfhVoowSC48X2DqAGejGJ0QBgiVMqX8olkZeOfKGyKxwdDWt1olU6RN8DgLUIkpj0wicJEh0N6D6Nxuw76k+CRmdE8mbxvKVZfv6ieICZBTNyz8FRz/cQKkLOOt9gtBsZJeCEhOTVmAOwTUYH9tGQe22mcbTTe88oWA8BvycpVUqZzEEJoZ/wPCpVh3b3sBSrOk17grZ7/5hhsNsQzerV0IIjl9VeAdYg1CTwrYp3OOHgYlF4/kDCu/7lNsZA68dZMNhs40IfEwycRucVj9muMNe2Bq0zxVMz3zClJ0JFdrXISSh471Jk333eQujT+szVYauKadMjmRHC4eO6aGkdaF65LF1PP1KsjR/UaWwJ8ZItpgQJ/XczBcrj68ffK7urfa9IfMAjAVbX3o/sVrw0NjEZTo19cp0tyVOq8+Q4tV+32Hoyh7Rhx/841EErDj32qvNpAmrRGPDhxC6dEL6G8XhvctzMVYu8i/1UIAncxtUgx1RwmrxdgVPsm9dOVppLNaguTJT7mn0d5kZeLezO/WS/KmcR2PpOgWVoQLV8EY6gLTtdyMyGf0wk7pvLo7+ONZy5eotF/mmLQoZnxcV2EZuXFJHhp0TJzA4emuB9I2blQQCg3NknL+15SA+jRRFgCjCITalMjfRHNZLEfLxVBxIHiJkYDGMlGEzjdDuQ0LrD9k8OkQ7tWvM5bTQpUfYR7ujZlDy/B9FZ3ODhQgk4eRZw0FIKhLB2N9ybu8F9uoc4Nf2tJQYAgBPoe9Z1Hw2ct+/bFRB+ZDUqK8y0xaSqIg86aL9TIBZfpmBZ3XUvFxLOIK0MTbu1MQuZFlkKdZ3y0Ho6A9x4DQ2vhNxf1bqPLxAUDSlX8ryv8/0fq45ySPg2L2NUIqZd7zF7jNxw8GCLI8tjqIwb2BojN5dXEOhiyAkDrd/J1Jxl5q0YjklnRehd4PnQaJClwwBqrVwy1WwlS/5mX6lyXGi9tL6SOgNXQcAWvYLZRrXjlP4jg27irw8cfn7cUaRy3hX4k9S2lz4zFsD3xcY3etm6SLybFU3YV6pcpbWgsOMDdhkZ21o4BqKYOWJ++eI4CuAMTWQYrYY1OWt6h+PKaJrcNGxeXzUiQT5crvnm+wyEZIQBcs4x3Fjp8s2Gs6B37aWbJUJIMfFP1NVfl7xHFpo0bAK4uu0vmY/N5Wzvj6gfYUGoOGnQ7dQSVNRCRjNRd8PmNkY0rvwpa8YbOmNMSui0d3YAmuaCEXNa8HeWr5wl/wJk400Ej31PB0FlDnM/x4KBxPvBlI2D/jZoipIW3IoCFzCa7vpa1G3lRApzJNou1LYRXygxtZAA+vcowzENgmKme0vCmFpbAs0yqFj/8xGdtrPNx+E/gI7dBtR1706InFcvAFJc5MnNUzfzgMVgb/cDhOlQf1AOGoTCHuC+r+upTKg7LKdlvZYp5JfaCphNcTVC7HNxASV60TnXDOu/+D03Qh1rkPx9p5sRbpQHkmupAPjoD2jlB+NSP3va640ZuM3ER0Hmn8HBeJkVGKYfot9QXId6wTfrViazZLbt7fA+ooK5BCvxDaLkPDUhofjINAb1AwpT0fE8QvjatbXw7wWNklMTjKkqZepoIyKqZu6MYA9L5bKG9rZ7jT56zpmaZdpapezdXjeGF7RYGIFZK+PK2f+14v54GRVqZY8b05DiX8RqzytxynhswiWyBrz5XwRHpTfMFtOOI7RjwiqY5B7t+E1UBmH/uF1T3plxjJNJfy2qxMUe43C2fbtNWpwHwmP3LbJJuqRcfRmXFohjvBC/r3hKeWI54gk70TKv+2I2DcDl/RohRoCqGspSsFOAbTKhAmWe0nHngPDGbXPkIiY6faHL6MWfb/QNNcVrO8H3KCtu8Tn/4e8V/Fg4aWyH7hc2GNYgtaD03allCPdbk5PP01Mlx5JilFosHUAYEYYn3W6mF9gNQbfdQWTGbY21XEQlmu9LK6+o6gNsdlSTQdh/x4Jknc",
		}).
		Post("https://mssdk.bytedance.com/web/common")
	if err != nil {
		return "", fmt.Errorf("request msToken api error: %w", err)
	}

	for _, resp := range resp.Cookies() {
		if strings.HasPrefix(resp.Name, "msToken") {
			d.msToken.value = resp.Value
			d.msToken.lastUpdated = time.Now()
			return d.msToken.value, nil
		}
	}
	if d.msToken.value == "" {
		return "", fmt.Errorf("get msToken error: %w", err)
	}

	return d.msToken.value, nil
}

// TODO
func (d *Douyin) getCookie() (string, error) {
	return "", nil
}

type cookieParams struct {
	verifyFp string
	fp       string
	uifid    string
}

func (d *Douyin) cookieQuery() (cookieParams, error) {
	cookieToMap := map[string]string{}
	cookie, err := d.getCookie()
	if err != nil {
		return cookieParams{}, fmt.Errorf("get cookie error: %w", err)
	}
	split := strings.Split(cookie, "; ")
	for _, v := range split {
		kv := strings.Split(v, "=")
		if len(kv) == 2 {
			cookieToMap[kv[0]] = kv[1]
		}
	}
	return cookieParams{
		verifyFp: cookieToMap["s_v_web_id"],
		fp:       cookieToMap["s_v_web_id"],
		uifid:    cookieToMap["UIFID"],
	}, nil
}
func (d *Douyin) authQueryParams() (map[string]string, error) {
	cookieToMap := map[string]string{}
	cookie, err := d.getCookie()
	if err != nil {
		return nil, fmt.Errorf("get cookie error: %w", err)
	}
	split := strings.Split(cookie, "; ")
	for _, v := range split {
		kv := strings.Split(v, "=")
		if len(kv) == 2 {
			cookieToMap[kv[0]] = kv[1]
		}
	}
	return map[string]string{
		"verifyFp": cookieToMap["s_v_web_id"],
		"fp":       cookieToMap["s_v_web_id"],
		"uifid":    cookieToMap["UIFID"],
	}, nil
}
