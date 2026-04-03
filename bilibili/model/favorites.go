package model

type FavoritesResponse struct {
	Code    int64         `json:"code"`
	Message string        `json:"message"`
	TTL     int64         `json:"ttl"`
	Data    FavoritesData `json:"data"`
}

type FavoritesData struct {
	Count  int64          `json:"count"`
	List   []FavoriteItem `json:"list"`
	Season any            `json:"season"`
}

type FavoriteItem struct {
	ID         int64  `json:"id"`
	Fid        int64  `json:"fid"`
	Mid        int64  `json:"mid"`
	Attr       int64  `json:"attr"`
	Title      string `json:"title"`
	FavState   int64  `json:"fav_state"`
	MediaCount int64  `json:"media_count"`
}
