package api

import (
	"encoding/json/v2"
	"os"
	"path/filepath"
	"testing"

	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/bilibili/util"
)

func TestDownloadHistorySQLitePagingSearchAndDeletion(t *testing.T) {
	b, store := testAPI(t)
	file := filepath.Join(t.TempDir(), "saved.mp4")
	if err := os.WriteFile(file, []byte("saved file"), 0600); err != nil {
		t.Fatal(err)
	}
	save := func(item model.DownloadHistoryItem) {
		t.Helper()
		raw, err := json.Marshal(item)
		if err != nil {
			t.Fatal(err)
		}
		key := util.DownloadCacheKey(item.Cid)
		if item.DownloadKind == "audio" {
			key = util.AudioDownloadCacheKey(item.Cid)
		}
		if err := store.SaveHistory(key, string(raw)); err != nil {
			t.Fatal(err)
		}
	}
	save(model.DownloadHistoryItem{Cid: 1, Title: "Rust 视频", UpperName: "Author", Bvid: "BVfirst", Downloaded: "2026-10-28T01:00:00Z", Path: file})
	save(model.DownloadHistoryItem{Cid: 1, Title: "Rust 音频", DownloadKind: "audio", Downloaded: "2026-10-28T02:00:00Z", Path: file})
	save(model.DownloadHistoryItem{Cid: -1, Title: "Rust 封面", DownloadKind: "cover", Downloaded: "2026-10-28T03:00:00Z", Path: file})
	page, err := b.DownloadHistoryPage(0, 2, "")
	if err != nil || page.Total != 3 || !page.HasMore || len(page.Items) != 2 || page.Items[0].DownloadKind != "cover" {
		t.Fatalf("first page: %+v %v", page, err)
	}
	page, err = b.DownloadHistoryPage(2, 2, "")
	if err != nil || page.HasMore || len(page.Items) != 1 || page.Items[0].Cid != 1 {
		t.Fatalf("second page: %+v %v", page, err)
	}
	for _, keyword := range []string{"author", "BVFIRST", "视频"} {
		page, err = b.DownloadHistoryPage(0, 50, keyword)
		if err != nil || page.Total != 1 || page.Items[0].Title != "Rust 视频" {
			t.Fatalf("filter %s: %+v %v", keyword, page, err)
		}
	}
	// New records must appear immediately, even after an earlier query.
	save(model.DownloadHistoryItem{Cid: 2, Title: "新增视频", Downloaded: "2026-10-28T04:00:00Z"})
	page, err = b.DownloadHistoryPage(0, 50, "")
	if err != nil || page.Total != 4 || page.Items[0].Cid != 2 {
		t.Fatalf("fresh read: %+v %v", page, err)
	}
	if err = b.DeleteDownloadHistory(1, "audio"); err != nil {
		t.Fatal(err)
	}
	if _, err = store.History(util.DownloadCacheKey(1)); err != nil {
		t.Fatal("audio deletion removed video", err)
	}
	if _, err = store.History(util.AudioDownloadCacheKey(1)); err == nil {
		t.Fatal("audio was not deleted")
	}
	if err = b.DeleteDownloadHistory(-1, "cover"); err != nil {
		t.Fatal(err)
	}
	if err = store.SaveHistory("douyin:downloaded:99", `{"title":"other platform"}`); err != nil {
		t.Fatal(err)
	}
	if err = b.ClearDownloadHistory(); err != nil {
		t.Fatal(err)
	}
	page, err = b.DownloadHistoryPage(0, 50, "")
	if err != nil || page.Total != 0 {
		t.Fatalf("clear: %+v %v", page, err)
	}
	if _, err = store.History("douyin:downloaded:99"); err != nil {
		t.Fatal("cleared another platform", err)
	}
	if _, err = os.Stat(file); err != nil {
		t.Fatal("history deletion removed local file", err)
	}
}
