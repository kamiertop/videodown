package storage

import (
	"encoding/json/v2"
	"strings"
)

type historyMeta struct {
	Bvid       string `json:"bvid"`
	Cid        int64  `json:"cid"`
	Title      string `json:"title"`
	UpperName  string `json:"upperName"`
	AuthorName string `json:"authorName"`
	Downloaded string `json:"downloaded"`
}

func (s *Store) setHistory(key, platform, payload string) error {
	var meta historyMeta
	if err := json.Unmarshal([]byte(payload), &meta); err != nil {
		return err
	}
	author := meta.UpperName
	if platform == "douyin" || platform == "youtube" {
		author = meta.AuthorName
	}
	_, err := s.db.NamedExec(`INSERT INTO download_history(key,platform,bvid,cid,title,author,downloaded,payload)
 VALUES(:key,:platform,:bvid,:cid,:title,:author,:downloaded,:payload)
 ON CONFLICT(key) DO UPDATE SET bvid=excluded.bvid,cid=excluded.cid,title=excluded.title,author=excluded.author,downloaded=excluded.downloaded,payload=excluded.payload`, map[string]any{
		"key": key, "platform": platform, "bvid": strings.ToUpper(strings.TrimSpace(meta.Bvid)), "cid": meta.Cid,
		"title": meta.Title, "author": author, "downloaded": meta.Downloaded, "payload": payload,
	})
	return err
}

// QueryHistory pages and filters records in SQLite rather than loading the entire history.
func (s *Store) QueryHistory(platform string, offset, limit int, keyword string) ([]string, int, error) {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()
	const filter = ` FROM download_history WHERE platform=? AND (?='' OR instr(lower(title),?)>0 OR instr(lower(author),?)>0 OR instr(lower(bvid),?)>0)`
	args := []any{platform, keyword, keyword, keyword, keyword}
	var total int
	if err = tx.Get(&total, "SELECT count(*)"+filter, args...); err != nil {
		return nil, 0, err
	}
	rows := []string{}
	args = append(args, limit, offset)
	if err = tx.Select(&rows, "SELECT payload"+filter+" ORDER BY downloaded DESC,key DESC LIMIT ? OFFSET ?", args...); err != nil {
		return nil, 0, err
	}
	if err = tx.Commit(); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
