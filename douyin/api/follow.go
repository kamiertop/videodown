package api

import "github.com/kamiertop/videodown/douyin/model"

// FollowList 关注列表，需要分页获取，前端默认20条。这里不再传参了，就默认20条了。
// https://www-hj.douyin.com/aweme/v1/web/im/spotlight/relation/ 也返回关注列表，但是最后一个元素是自身信息
func (d *Douyin) FollowList() (model.Follow, error) {
	d.client.
		Get("https://www-hj.douyin.com/aweme/v1/web/user/following/list/").
		SetQueryParamsAnyType(map[string]any{
			"count": 20, // 默认20条
		})
	return model.Follow{}, nil
}
