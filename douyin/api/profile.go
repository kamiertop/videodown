package api

import (
	"fmt"
	"runtime"

	"github.com/kamiertop/videodown/douyin/model"
)

// Info 个人信息接口
// 没找到单独返回个人信息的接口，这个接口返回的第一个是个人信息（没想明白为啥这么设计）
func (d *Douyin) Info() (model.MyInfo, error) {
	var resp model.MyInfo
	uiFid, err := d.getUiFid()
	if err != nil {
		return resp, err
	}
	d.client.
		Get("https://www-hj.douyin.com/aweme/v1/web/im/user/info/").
		SetQueryParamsAnyType(map[string]any{
			"device_platform":     "webapp",
			"aid":                 6383,
			"channel":             "channel_pc_web",
			"pc_client_type":      1,
			"pc_libra_divert":     runtime.GOOS,
			"update_version_code": 170400,
			"support_h256":        1,
			"support_dash":        1,
			"version_code":        "170400",
			"version_name":        "17.4.0",
			"cookie_enabled":      true,
			"screen_width":        1920,
			"screen_height":       1080,
			"browser_language":    "zh-CN",
			"borwser_platform":    fmt.Sprintf("%s %s", runtime.GOOS, "X86_64"),
			"browser_name":        "Chrome",
			"browser_version":     "146.0.0.0",
			"browser_online":      true,
			"engine_name":         "Blink",
			"engine_version":      "146.0.0.0",
			"os_name":             runtime.GOOS,
			"os_version":          "x86_64",
			"cpu_core_num":        runtime.NumCPU(),
			"device_memory":       8,
			"platform":            "PC",
			"downlink":            0.4,
			"effective_type":      "4g",
			"round_trip_time":     50,
			"webid":               "7624551505828415026",
			"uifid":               uiFid,
			"msToken":             "zj8n9sK7mXoZlHqg2a3u1eQyLh5v6n9w4x7y8z0A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6a7b8c9d0e1f2g3h4i5j6k7l8m9n0o1p2q3r4s5t6u7v8w9x0y1z2A3B4C5D6E7F8G9H0I1J2K3L4M5N6O7P8Q9R0S1T2U3V4W5X6Y7",
			"a_bogus":             "",
			"verifyFp":            "",
			"fp":                  "",
		})

	return model.MyInfo{}, nil
}
