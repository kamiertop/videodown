package download

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kamiertop/videodown/utils"
)

func TestCleanDouyinTitle(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{
			name: "尾部话题标签被剥掉",
			in:   "猫咪总把东西推下桌子的原因找到了 #猫咪 #养猫日常 #萌宠",
			want: "猫咪总把东西推下桌子的原因找到了",
		},
		{
			name: "无标签标题原样保留",
			in:   "普通标题",
			want: "普通标题",
		},
		{
			name: "纯标签标题保留原文",
			in:   "#挑战 #日常",
			want: "#挑战 #日常",
		},
		{
			name: "正文中间的井号不受影响",
			in:   "含#号的正文结尾",
			want: "含#号的正文结尾",
		},
		{
			// 无空白分隔的标签串与正文无法可靠区分，保守起见整体保留；
			// 残留的标签只影响观感，超长安全由 safeFileName 兜底。
			name: "无空格分隔的标签串保守保留",
			in:   "正文 #tag1#tag2",
			want: "正文 #tag1#tag2",
		},
		{
			name: "全角空格分隔的标签",
			in:   "正文　#猫咪　#萌宠　",
			want: "正文",
		},
	}
	for _, c := range cases {
		if got := cleanDouyinTitle(c.in); got != c.want {
			t.Errorf("%s: cleanDouyinTitle(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}

func TestSafeFileName(t *testing.T) {
	// 超过安全上限的名字退回视频 ID，短名字原样保留。
	long := strings.Repeat("抖", utils.MaxFileNameBytes/3+1)
	if got := safeFileName(long, "7123456789012345678"); got != "7123456789012345678" {
		t.Errorf("safeFileName(over-long) = %q, want fallback id", got)
	}
	short := strings.Repeat("抖", 10)
	if got := safeFileName(short, "7123456789012345678"); got != short {
		t.Errorf("safeFileName(short) = %q, want %q", got, short)
	}
}

// 全链路回归：超长简介 → 剥尾部标签 → 仍超限 → 退回视频 ID → 成功落盘。
// 修复前超长名字会让 UniqueFilePath 无限循环，下载 worker 和前端一起卡死。
func TestOverLongTitleFallsBackToIDAndCreatesFile(t *testing.T) {
	const awemeID = "7123456789012345678"
	desc := strings.Repeat("看完这条视频你就懂了为什么猫咪总爱把东西推下桌子", 20) +
		" #猫咪 #养猫日常 #萌宠 #宠物博主"

	// downloadTask 的清洗与命名路径
	title := cleanDouyinTitle(desc)
	if strings.Contains(title, "#") {
		t.Fatalf("trailing hashtags not stripped: %q", title)
	}
	values := map[string]string{"title": title, "id": awemeID}
	fileName := utils.ApplyFilenameTemplate("{title}", values)
	if fileName == "" {
		fileName = "douyin"
	}
	fileName = safeFileName(fileName, awemeID)
	if fileName != awemeID {
		t.Fatalf("over-long title should fall back to aweme id, got %q (%d bytes)", fileName, len(fileName))
	}

	dir := t.TempDir()
	outPath := utils.UniqueFilePath(filepath.Join(dir, fileName+".mp4"))
	if err := os.WriteFile(outPath, []byte("x"), 0o644); err != nil {
		t.Fatalf("create file: %v", err)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("stat created file: %v", err)
	}
}
