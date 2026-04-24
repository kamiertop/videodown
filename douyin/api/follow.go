package api

import (
	"errors"
	"fmt"

	"github.com/kamiertop/videodown/douyin/model"
)

// FollowList 关注列表，需要分页获取，前端默认20条。这里不再传参了，就默认20条了。
// https://www-hj.douyin.com/aweme/v1/web/im/spotlight/relation/ 也返回关注列表，但是最后一个元素是自身信息
func (d *Douyin) FollowList(offset int) (model.FollowResponse, error) {
	var resp model.FollowResponse

	queryParams, err := d.publicQueryParams()
	if err != nil {
		return resp, fmt.Errorf("获取公共查询参数失败: %w", err)
	}
	publicHeaders, err := d.publicHeaders()
	if err != nil {
		return resp, fmt.Errorf("获取公共请求头失败: %w", err)
	}

	publicHeaders["Accept"] = "*/*"
	err = d.client.
		Get("https://www-hj.douyin.com/aweme/v1/web/user/following/list/").
		SetQueryParamsAnyType(map[string]any{
			"user_id":              d.userID,
			"sec_user_id":          d.secUserID,
			"offset":               offset, // 偏移量，初始为0，后续每次请求加上返回的count
			"min_time":             0,
			"max_time":             0,
			"count":                20,
			"source_type":          4,
			"gps_access":           0,
			"is_top":               1,
			"address_book_access":  0,
			"webcast_sdk_version":  "170400",
			"webcast_version_code": "170400",
		}).
		SetQueryParamsAnyType(queryParams).
		SetHeaders(publicHeaders).
		Do().
		Into(&resp)
	if err != nil {
		return resp, fmt.Errorf("请求关注列表失败: %w", err)
	}
	if resp.StatusCode != 0 {
		d.logger.Errorf("request follow list failed, status_code=%d, offset=%d", resp.StatusCode, offset)
		return resp, errors.New("请求关注列表失败")
	}

	return resp, nil
}

/*
curl 'https://www-hj.douyin.com/aweme/v1/web/user/following/list/?device_platform=webapp&aid=6383&channel=channel_pc_web&user_id=369908084126571&sec_user_id=MS4wLjABAAAAQT1uJyGAwUGj6-KEv7Qd-BdrvBaf_lfo261XvWeWqvY&offset=40&min_time=0&max_time=0&count=20&source_type=4&gps_access=0&address_book_access=0&is_top=1&pc_client_type=1&pc_libra_divert=Linux&support_h265=1&support_dash=1&webcast_sdk_version=170400&webcast_version_code=170400&version_code=170400&version_name=17.4.0&cookie_enabled=true&screen_width=1920&screen_height=1080&browser_language=zh-CN&browser_platform=Linux+x86_64&browser_name=Chrome&browser_version=147.0.0.0&browser_online=true&engine_name=Blink&engine_version=147.0.0.0&os_name=Linux&os_version=x86_64&cpu_core_num=24&device_memory=32&platform=PC&downlink=1.5&effective_type=4g&round_trip_time=100&webid=7624551505828415026&verifyFp=verify_mnj1nqe4_KfnaXUjl_6eBF_4lhZ_AICz_3R3ZbTjX9Dqp&fp=verify_mnj1nqe4_KfnaXUjl_6eBF_4lhZ_AICz_3R3ZbTjX9Dqp&msToken=X_q14x7d2lEWiik1ox_CNamZqE_RNP2p3WChZbWxRZyTP5ELTM09utoLBRg2S7oIwx3dO6kwriI5iXlXPznuweMCFIMqHMPAAEi0sSjr_dSD3SQZ926TX1FTsKTyHfIXmsw1VihlKAoBxJthTAhON9tMlbBCqAaEekdFDBkqcRrJs6R3xcaZyMM%3D' \
-H 'accept-language: zh-CN,zh;q=0.9' \
-H 'cache-control: no-cache' \
-b 'enter_pc_once=1; UIFID_TEMP=0ee2bb4d1dbbbad62c3519ca61edd2ce41ae42602b6f1337b85aa1470e4e362e323cc956e9e307925ebbcb182b5ceac89e7cca4dac61d6a1ff54f7ef764c1b1efd3f68a4808bb41f687ce54e231691ee; hevc_supported=true; passport_csrf_token=dba0529e1c451eaa19076ddf8d481a26; passport_csrf_token_default=dba0529e1c451eaa19076ddf8d481a26; is_dash_user=1; bd_ticket_guard_client_web_domain=2; n_mh=mnC8uhbiFoI4iqreqfLzb8_HCoavtMGNna34PGdpejY; is_staff_user=false; has_biz_token=false; __security_mc_1_s_sdk_cert_key=608c6a3f-4692-a7ed; __security_mc_1_s_sdk_crypt_sdk=4130f761-450d-b0be; __security_server_data_status=1; UIFID=0ee2bb4d1dbbbad62c3519ca61edd2ce41ae42602b6f1337b85aa1470e4e362e323cc956e9e307925ebbcb182b5ceac84d7bd07aab3ec884e865098b7074c2800498fd94351c2ea519ac10fad4d3fc474953cd12c728a5617f9515e06e517032822e83dcb3cd8e8933ead95043419981296f13fb8b0c1d17d6727ff99fd6fe82a8a8e80c56c9e34b537087bce8101acc18859d0ab1e697e235487ea7d5a3f2cf; my_rd=2; live_use_vvc=%22false%22; __druidClientInfo=JTdCJTIyY2xpZW50V2lkdGglMjIlM0E2MTAlMkMlMjJjbGllbnRIZWlnaHQlMjIlM0E3MjElMkMlMjJ3aWR0aCUyMiUzQTYxMCUyQyUyMmhlaWdodCUyMiUzQTcyMSUyQyUyMmRldmljZVBpeGVsUmF0aW8lMjIlM0ExLjIwMzEyNSUyQyUyMnVzZXJBZ2VudCUyMiUzQSUyMk1vemlsbGElMkY1LjAlMjAoWDExJTNCJTIwTGludXglMjB4ODZfNjQpJTIwQXBwbGVXZWJLaXQlMkY1MzcuMzYlMjAoS0hUTUwlMkMlMjBsaWtlJTIwR2Vja28pJTIwQ2hyb21lJTJGMTQ2LjAuMC4wJTIwU2FmYXJpJTJGNTM3LjM2JTIyJTdE; __live_version__=%221.1.5.1039%22; volume_info=%7B%22isMute%22%3Afalse%2C%22volume%22%3A0.5%2C%22isUserMute%22%3Afalse%7D; PhoneResumeUidCacheV1=%7B%22369908084126571%22%3A%7B%22time%22%3A1775823688709%2C%22noClick%22%3A0%7D%7D; live_debug_info=%7B%22roomId%22%3A%227627101021457730347%22%2C%22resolution%22%3A%7B%22width%22%3A480%2C%22height%22%3A852%7D%2C%22fps%22%3A164%2C%22audioDataRate%22%3A44100%2C%22speed%22%3A%7B%22type%22%3A%22flv_prop_speed%22%2C%22demuxFrames%22%3A5733%2C%22decodeFrames%22%3A5708%2C%22decodedFrames%22%3A5706%2C%22renderFrames%22%3A5697%2C%22totalByteSize%22%3A36451129%2C%22currentSpeed%22%3A126%2C%22avgSpeed%22%3A1008000%2C%22recentSpeed%22%3A1096000%7D%2C%22droppedFrames%22%3A995%2C%22totalFrames%22%3A5784%2C%22videoBuffer%22%3A%5B%5D%2C%22src%22%3A%22https%3A%2F%2Fpull-q5.douyincdn.com%2Fstage%2Fstream-407403829890646844_Stage0T000ld5.flv%3Farch_hrchy%3Dh1%26exp_hrchy%3Dh1%26expire%3D1776428498%26major_anchor_level%3Dcommon%26sign%3Dbcdaa5f3a61916fde58220c531bb7b2d%26t_id%3D037-20260410202137BD84C521B31B711EC11B-60XQU1%26unique_id%3Dstream-407403829890646844_828_flv_Stage0T000ld5%26volcSecret%3Dbcdaa5f3a61916fde58220c531bb7b2d%26volcTime%3D1776428498%26_session_id%3D037-20260410202137BD84C521B31B711EC11B-60XQU1.1775823698697.88238%26rsi%3D1%26abr_pts%3D-800%22%2C%22linkmicInfo%22%3A%7B%22uiLayout%22%3A0%2C%22playModes%22%3A%5B%5D%2C%22allDevices%22%3A%22%E8%BF%9E%E7%BA%BF%E8%AE%BE%E5%A4%87%EF%BC%9A%E7%94%B3%E8%AF%B7%E8%BF%9E%E7%BA%BF%E5%90%8E%E6%89%8D%E8%8E%B7%E5%8F%96%22%2C%22audioInputs%22%3A%5B%5D%2C%22videoInputs%22%3A%5B%5D%7D%2C%22href%22%3A%22https%3A%2F%2Flive.douyin.com%2F948919062446%3Ffrom_app_resume%3Dlive%22%7D; publish_badge_show_info=%220%2C0%2C0%2C1776949059318%22; FOLLOW_LIVE_POINT_INFO=%22MS4wLjABAAAAQT1uJyGAwUGj6-KEv7Qd-BdrvBaf_lfo261XvWeWqvY%2F1776960000000%2F0%2F1776954761540%2F0%22; strategyABtestKey=%221776993351.603%22; ttwid=1%7CleBzuSU5UDD0m1-LzPQ4p48WK65Q6EWs6raufeEqk3E%7C1776993663%7Ced629c47627473a281e839b672d36a0e6ade779fc1aa29188e40916cb6064c56; passport_assist_user=CkDOg9tgKy7yu7Uxempuq9KQY7o0e36Utsv2whc4WhgCYttUYCnzR-L5SBxVWvCg6onmj9alMrz7d0--BuEbn9JZGkoKPAAAAAAAAAAAAABQV20LC2qt0YOGH2YaSbdrxmMmac86YhuGq7HToxNHaiu8w1iLfy7nKlKNj3IEMJiXtBCq048OGImv1lQgASIBAzub5gU%3D; sid_guard=12da904d6afe4ca3fc096ce4ddeae6fd%7C1776993668%7C5184000%7CTue%2C+23-Jun-2026+01%3A21%3A08+GMT; uid_tt=4b2df44a5c1d8a589a8f41f2b5cdbb74; uid_tt_ss=4b2df44a5c1d8a589a8f41f2b5cdbb74; sid_tt=12da904d6afe4ca3fc096ce4ddeae6fd; sessionid=12da904d6afe4ca3fc096ce4ddeae6fd; sessionid_ss=12da904d6afe4ca3fc096ce4ddeae6fd; session_tlb_tag=sttt%7C5%7CEtqQTWr-TKP8CWzk3erm_f_________dRTjM_avfn4JPBTw62CaWBzdUOV90i0FzU3loGrigBpk%3D; sid_ucp_v1=1.0.0-KGI3NGVhNDllYmVjYjA0ODRlMDkwMjc3YTE3MGZhYjEyMGMzNGM3ZDEKIAjr5tD_3o1UEISLq88GGO8xIAww68z8hgY4BUD7B0gEGgJsZiIgMTJkYTkwNGQ2YWZlNGNhM2ZjMDk2Y2U0ZGRlYWU2ZmQ; ssid_ucp_v1=1.0.0-KGI3NGVhNDllYmVjYjA0ODRlMDkwMjc3YTE3MGZhYjEyMGMzNGM3ZDEKIAjr5tD_3o1UEISLq88GGO8xIAww68z8hgY4BUD7B0gEGgJsZiIgMTJkYTkwNGQ2YWZlNGNhM2ZjMDk2Y2U0ZGRlYWU2ZmQ; __security_mc_1_s_sdk_sign_data_key_web_protect=d4a467a5-4375-9245; login_time=1776993668792; _bd_ticket_crypt_cookie=d192e7d9d1f95c443b063b788e9eb846; DiscoverFeedExposedAd=%7B%7D; FOLLOW_NUMBER_YELLOW_POINT_INFO=%22MS4wLjABAAAAQT1uJyGAwUGj6-KEv7Qd-BdrvBaf_lfo261XvWeWqvY%2F1777046400000%2F0%2F0%2F1776995703737%22; FOLLOW_RED_POINT_INFO=%221%22; is_support_rtm_web_ts=1; stream_recommend_feed_params=%22%7B%5C%22cookie_enabled%5C%22%3Atrue%2C%5C%22screen_width%5C%22%3A1920%2C%5C%22screen_height%5C%22%3A1080%2C%5C%22browser_online%5C%22%3Atrue%2C%5C%22cpu_core_num%5C%22%3A24%2C%5C%22device_memory%5C%22%3A32%2C%5C%22downlink%5C%22%3A1.5%2C%5C%22effective_type%5C%22%3A%5C%224g%5C%22%2C%5C%22round_trip_time%5C%22%3A100%7D%22; bd_ticket_guard_client_data=eyJiZC10aWNrZXQtZ3VhcmQtdmVyc2lvbiI6MiwiYmQtdGlja2V0LWd1YXJkLWl0ZXJhdGlvbi12ZXJzaW9uIjoxLCJiZC10aWNrZXQtZ3VhcmQtcmVlLXB1YmxpYy1rZXkiOiJCQnN3Z2xMNzZvY3c3cGIzZ1hQL3pvUjIrT0oxSTlCRGRjMjFHNUQvL1RxVUdJZkcyS0hJbjJpZkxRTWFhdmV3cVFYMWtjd3dERjI0MWVMdzJ4L2cwRXc9IiwiYmQtdGlja2V0LWd1YXJkLXdlYi12ZXJzaW9uIjoyfQ%3D%3D; home_can_add_dy_2_desktop=%221%22; biz_trace_id=f90efde1; odin_tt=576a384c1b5144a99a6a5b32885649a59a5d95ad25882727557696729c012c5bcf5c68a1b94b34413a5a005ce8c5da40bac9f935606f381ac9b84168e913ce06; IsDouyinActive=true; sdk_source_info=7e276470716a68645a606960273f276364697660272927676c715a6d6069756077273f276364697660272927666d776a68605a607d71606b766c6a6b5a7666776c7571273f275e58272927666a6b766a69605a696c6061273f27636469766027292762696a6764695a7364776c6467696076273f275e582729277672715a646971273f2763646976602729277f6b5a666475273f2763646976602729276d6a6e5a6b6a716c273f2763646976602729276c6b6f5a7f6367273f27636469766027292771273f2731323c32323533343532323234272927676c715a75776a716a666a69273f2763646976602778; bit_env=6L2aI30Zur6I13VAVKwG8fWS1wpIgKTL8z_S5ZnxH0Sm8S7F3iyVBCEeAycLs2wkrbfy6keePR7eSpsGK7tPsNv7BqUgWJ6Irv_aS9RDimi1lpGLmpjhG6nLAuL-MjK_4ROPTr81vA5HDXDh5_APnl1gVhjLbWDVn6rZKd4FeAleguX7vtwUqtGn6IZYXplmjKWUBUhjeI57AZ7-9WvHd--ni5UFtKK8Rdj3AhlA2g2E_2q-FQ9lgW-0R0tCkAEWz_itu5W-3cD1YK_5Avk5cNnJMTq3IUNAtuuLy30jUFtJ5vX0wULCzHD_Ed0tDi37kS912RiNf4unlsk2eFRCFG4MhjIS-FLLX6_GqSw4vJY19Tak_Qlt1lDmw0SAmwDRNr7HkjlPpnEnxnG2yGxOdJQoz03qQZllN6Y1q3eSxQaxnCkNQ34IOg_GZShyX__1FYFFGqnWq1X8IKFA8HK78SBaUTP6eQxE5l6hIWawoKjI8uJ5fAN80GzoqQCqNEE8; gulu_source_res=eyJwX2luIjoiYTEwZTY0NjM0MzAxZTNkYWZkNzc4MGFmYmNlMTA2OWFmMTdkNzEwNjUzZWM4ZjZhMTg0NTQ2YjNlMmUzMWRlNCJ9; bd_ticket_guard_client_data_v2=eyJyZWVfcHVibGljX2tleSI6IkJCc3dnbEw3Nm9jdzdwYjNnWFAvem9SMitPSjFJOUJEZGMyMUc1RC8vVHFVR0lmRzJLSEluMmlmTFFNYWF2ZXdxUVgxa2N3d0RGMjQxZUx3MngvZzBFdz0iLCJ0c19zaWduIjoidHMuMi4yNWUwY2NjMTBjMWUyOTA3OGE1ZDM5NDBjNzAyNWMzNjVhMDViOGM2OWY3ZTY1MGIxZDcyNTg5ZDkxMjgwZWM2YzRmYmU4N2QyMzE5Y2YwNTMxODYyNGNlZGExNDkxMWNhNDA2ZGVkYmViZWRkYjJlMzBmY2U4ZDRmYTAyNTc1ZCIsInJlcV9jb250ZW50Ijoic2VjX3RzIiwicmVxX3NpZ24iOiJzalc3TFRPYmhRczM2UDR6bjQ5ZTk2UmUzbk1icUFram9RU1BxV3pDaEt3PSIsInNlY190cyI6IiNwQ0hEWGJnR2M5ZGM4N1VLL0tXVG1HT0lxaTcrWVNwZWpQbjNnR2RsMVhjUHFGZFZpNTdQb3piRzl1TDQifQ%3D%3D; SelfTabRedDotControl=%5B%7B%22id%22%3A%227615141649374513187%22%2C%22u%22%3A85%2C%22c%22%3A85%7D%2C%7B%22id%22%3A%227492338664471529510%22%2C%22u%22%3A223%2C%22c%22%3A224%7D%2C%7B%22id%22%3A%227542008678845515827%22%2C%22u%22%3A27%2C%22c%22%3A27%7D%2C%7B%22id%22%3A%227499517529748932617%22%2C%22u%22%3A15%2C%22c%22%3A18%7D%5D' \
-H 'origin: https://www.douyin.com' \
-H 'pragma: no-cache' \
-H 'priority: u=1, i' \
-H 'referer: https://www.douyin.com/' \
-H 'sec-ch-ua: "Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"' \
-H 'sec-ch-ua-mobile: ?0' \
-H 'sec-ch-ua-platform: "Linux"' \
-H 'sec-fetch-dest: empty' \
-H 'sec-fetch-mode: cors' \
-H 'sec-fetch-site: same-site' \
-H 'user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36'
*/
