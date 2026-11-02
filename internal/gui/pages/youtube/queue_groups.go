package youtube

import (
	"fmt"
	"slices"
)

type queueGroup struct {
	Key, Title string
	Expanded   bool
}

func (s *State) groupItems(group *queueGroup) []*queueItem {
	var items []*queueItem
	for _, item := range s.queue {
		if item.Group == group {
			items = append(items, item)
		}
	}
	return items
}

func (s *State) pruneQueueGroups() {
	s.queueGroups = slices.DeleteFunc(s.queueGroups, func(group *queueGroup) bool { return len(s.groupItems(group)) == 0 })
}

func (s *State) removeGroup(group *queueGroup) {
	for _, item := range s.groupItems(group) {
		s.cancelItem(item)
		s.cancelDetails(item)
		if item.Status == "active" {
			item.Status = "canceling"
		}
	}
	s.queue = slices.DeleteFunc(s.queue, func(item *queueItem) bool { return item.Group == group })
	s.pruneQueueGroups()
}

func groupSummary(items []*queueItem) string {
	done, waiting, failed, canceled, active, ready := 0, 0, 0, 0, 0, 0
	for _, item := range items {
		switch item.Status {
		case "done":
			done++
		case "ready":
			ready++
		case "queued":
			waiting++
		case "error":
			failed++
		case "canceled", "canceling":
			canceled++
		case "active":
			active++
		}
	}
	text := fmt.Sprintf("共 %d 项 · 下载中 %d · 等待 %d · 已完成 %d", len(items), active, waiting, done)
	if ready > 0 {
		text += fmt.Sprintf(" · 待下载 %d", ready)
	}
	if failed > 0 {
		text += fmt.Sprintf(" · 失败 %d", failed)
	}
	if canceled > 0 {
		text += fmt.Sprintf(" · 已取消 %d", canceled)
	}
	return text
}

func (s *State) playlistTask(row *playlistRow) *queueItem {
	return s.playlistTasks[row.Video.ID]
}

func (s *State) playlistAvailable(row *playlistRow) bool {
	item := s.playlistTask(row)
	return item == nil || item.Status == "error" || item.Status == "canceled" || item.Status == "canceling"
}

func (s *State) playlistTaskLabel(row *playlistRow) string {
	item := s.playlistTask(row)
	if item == nil {
		return ""
	}
	switch item.Status {
	case "ready", "queued":
		return "已加入队列"
	case "active":
		return "下载中"
	case "done":
		return "已完成"
	case "error":
		return "下载失败，可重新加入"
	case "canceled", "canceling":
		return "已取消，可重新加入"
	}
	return ""
}
