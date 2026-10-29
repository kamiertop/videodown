package navigation

import "testing"

func TestBilibiliAuthorRoutes(t *testing.T) {
	for _, id := range []uint64{1, 42, 18446744073709551615} {
		path := BilibiliAuthorPath(id)
		mid, ok := BilibiliAuthorID(path)
		if !ok || mid != id {
			t.Fatalf("author route: %s", path)
		}
		page, ok := ByPath(path)
		if !ok || page.ID != "bilibili" || page.CurrentSection(path) != "up" {
			t.Fatal("author route lost its platform/section")
		}
	}
	for _, path := range []string{"/bilibili/up/0", "/bilibili/up/-1", "/bilibili/up/abc", "/bilibili/up/1/extra", "/bilibili/up/18446744073709551616"} {
		if _, ok := ByPath(path); ok {
			t.Fatalf("unexpected author route: %s", path)
		}
	}
}

func TestBilibiliCollectionRoute(t *testing.T) {
	page, ok := ByPath("/bilibili/collection")
	if !ok || page.ID != "bilibili" || page.CurrentSection("/bilibili/collection") != "collection" {
		t.Fatal("collection route did not resolve as a Bilibili section")
	}
}

func TestDouyinNestedRoutes(t *testing.T) {
	for _, path := range []string{"/douyin/user/MS4wLjAB_example", "/douyin/favorite/folder", "/douyin/favorite/folder/42", "/douyin/favorite/collection/42"} {
		page, ok := ByPath(path)
		if !ok || page.ID != "douyin" {
			t.Fatalf("missing Douyin route %s", path)
		}
		if section := page.CurrentSection(path); section != "user" && section != "favorite" {
			t.Fatalf("nested route lost section: %s", section)
		}
	}
	for _, path := range []string{"/douyin/user/a/extra", "/douyin/favorite/unknown", "/douyin/dynamic/123"} {
		if _, ok := ByPath(path); ok {
			t.Fatalf("accepted undeclared route %s", path)
		}
	}
}
