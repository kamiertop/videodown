package model

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

func TestDynamicDataUnmarshalSampleResponse(t *testing.T) {
	raw, err := os.ReadFile("../../tests/resp.json")
	if errors.Is(err, os.ErrNotExist) {
		t.Skip("tests/resp.json not found")
	}
	if err != nil {
		t.Fatal(err)
	}

	var resp struct {
		ApiResponse
		Data DynamicData `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}

	items := resp.Data.GetArchiveItems()
	if len(items) == 0 {
		t.Fatal("expected at least one video item")
	}
	if items[0].Bvid == "" {
		t.Fatal("expected first video item to include bvid")
	}
	if items[0].Title == "" || items[0].DurationText == "" || items[0].AuthorName == "" || items[0].PubTs == 0 {
		t.Fatalf("expected display fields to be populated, got %+v", items[0])
	}
}
