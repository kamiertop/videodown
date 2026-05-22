package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	stdRuntime "runtime"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/kamiertop/videodown/logger"
)

const (
	// themeKey 主题设置，默认为 "light"
	themeKey = "theme"
	// storageKey 存储目录，默认为可执行文件所在目录的 download 子目录
	storageKey = "storage"
	// sleepTimeKey 下载完一个视频之后的随机休眠时间，单位为秒，默认为0-60 秒
	sleepTimeKey = "sleepTime"
	// allowGroupOnSaveKey 保存时是否自动分组，默认为 true
	allowGroupOnSaveKey = "allowGroupOnSave"
	// concurrencyNumKey 保存同时下载的视频数量，默认为 1
	concurrencyNumKey = "concurrencyNum"
)

type Settings struct {
	*badger.DB
	logger *logger.Logger
	ctx    context.Context
}

// Startup is called at application Startup
func (s *Settings) Startup(ctx context.Context) {
	s.ctx = ctx
}

// DomReady is called after front-end resources have been loaded
func (s *Settings) DomReady(_ context.Context) {
	// Add your action here
}

func (s *Settings) Context() context.Context {
	return s.ctx
}

// BeforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (s *Settings) BeforeClose(_ context.Context) (prevent bool) {
	return false
}

func (s *Settings) init() error {
	executable, err := os.Executable()
	if err != nil {
		s.logger.Errorf("Get VideoDown Executable Path Error: %v", err)
		return err
	}

	return s.DB.Update(func(txn *badger.Txn) error {
		defaultValue := map[string]string{
			themeKey:             "light",
			storageKey:           filepath.Join(filepath.Dir(executable), "download"),
			allowGroupOnSaveKey:  "true",
			sleepTimeKey:         "60",
			concurrencyNumKey:    "1",
			autoUpdateKey:        "false",
			parsePlayURLNumKey:   "3",
			parsePlayURLSleepKey: "5",
			// 其他设置项的默认值
		}
		s.logger.Infof("set default storage path: %s", defaultValue[storageKey])
		var errList error
		for key, value := range defaultValue {
			if _, err := txn.Get([]byte(key)); errors.Is(err, badger.ErrKeyNotFound) {
				s.logger.Infof("No %s found, setting to default: %s", key, value)
				// 只有在 key 不存在时才设置默认值，避免覆盖用户已修改的设置。
				if err := txn.Set([]byte(key), []byte(value)); err != nil {
					errList = errors.Join(errList, fmt.Errorf("failed to set key: [%s], value: [%s], err: %w", key, value, err))
				}
			}
		}

		return errList
	})
}

func NewSettingsWithMemory(logger *logger.Logger) *Settings {
	db, err := badger.Open(badger.
		DefaultOptions("").
		WithInMemory(true).
		WithLoggingLevel(badger.ERROR))
	if err != nil {
		panic(err)
	}
	s := &Settings{DB: db, logger: logger.WithName("Settings")}

	if err := s.init(); err != nil {
		panic(err)
	}

	return s
}

func NewSettings(logger *logger.Logger) (*Settings, error) {
	dbPath := "videodown.db"
	if execPath, err := os.Executable(); err == nil {
		dbPath = filepath.Join(filepath.Dir(execPath), dbPath)
	}
	db, err := badger.Open(badger.DefaultOptions(dbPath).WithLogger(logger).WithLoggingLevel(badger.ERROR))
	if err != nil {
		panic(err)
	}

	s := &Settings{DB: db, logger: logger.WithName("Settings")}

	return s, s.init()
}

func NewSettingsForDebug(logger *logger.Logger, path string) (*Settings, error) {
	db, err := badger.Open(badger.
		DefaultOptions(path).
		WithLogger(logger).
		WithLoggingLevel(badger.ERROR),
	)

	if err != nil {
		panic(err)
	}

	s := &Settings{DB: db, logger: logger.WithName("Settings")}

	return s, s.init()
}

// GetTheme 获取主题设置
func (s *Settings) GetTheme() (string, error) {
	theme, err := s.GetKey(themeKey)
	if err != nil {
		s.logger.Errorf("failed to get theme: %v", err)
		return "", errors.New("获取主题设置失败")
	}

	return theme, nil
}

// SetTheme 设置主题，前端调用时会传入 "light"、"dark"等，落库保存供下次启动时加载使用
func (s *Settings) SetTheme(theme string) error {
	if err := s.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(themeKey), []byte(theme))
	}); err != nil {
		s.logger.Errorf("Failed to set new theme [%s], err: %v", theme, err)
		return errors.New("设置主题失败")
	}
	s.logger.Infof("Theme set to: %s", theme)

	return nil
}

// GetStorage 获取存储目录设置
func (s *Settings) GetStorage() (string, error) {
	path, err := s.GetKey(storageKey)
	if err != nil {
		s.logger.Errorf("failed to get storage path: %v", err)
		return "", errors.New("获取存储目录失败")
	}

	return path, nil
}

// SetStorage 设置存储目录，前端调用时会弹出目录选择对话框，用户选择后落库保存
func (s *Settings) SetStorage() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(s.ctx, runtime.OpenDialogOptions{
		Title: "选择下载目录",
	})

	if err = s.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(storageKey), []byte(dir))
	}); err != nil {
		s.logger.Errorf("Failed to set new storage path [%s], err: %v", dir, err)
		return "", errors.New("设置存储目录失败")
	}
	s.logger.Infof("Storage path set to: %s", dir)

	return dir, nil
}

func (s *Settings) SetKey(key, value string) error {
	return s.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(value))
	})
}

func (s *Settings) GetKey(key string) (string, error) {
	var result string
	err := s.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			result = string(val)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})

	return result, err
}

func (s *Settings) DeleteKey(key string) error {
	return s.DB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// GetSleepTime 下载完一个视频之后的休眠时间；配置值按“秒”保存，避免把默认值 60 误解释成 60 纳秒。
func (s *Settings) GetSleepTime() (int64, error) {
	value, err := s.GetKey(sleepTimeKey)
	if err != nil {
		s.logger.Errorf("failed to get sleep time: %v", err)
		return 0, errors.New("获取休眠时间失败")
	}
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		s.logger.Errorf("failed to parse sleep time: %v", err)
		return 60, nil
	}

	return val, nil
}

// SetSleepTime 保存休眠秒数；前端传入 time.Duration 时统一落库为秒，便于用户理解和配置。
func (s *Settings) SetSleepTime(d int64) error {
	s.logger.Infof("setting sleep time: %d", d)

	return s.SetKey(sleepTimeKey, strconv.FormatInt(d, 10))
}

func (s *Settings) GetSavePreference() (bool, error) {
	key, err := s.GetKey(allowGroupOnSaveKey)
	if err != nil {
		return true, err
	}
	if key == "true" {
		return true, nil
	}

	return false, nil
}

// SetSavePreference 保存时是否自动分组
func (s *Settings) SetSavePreference(allowGroup bool) error {
	var b string
	if allowGroup {
		b = "true"
	} else {
		b = "false"
	}
	s.logger.Infof("Setting save preference to: %s", b)

	return s.SetKey(allowGroupOnSaveKey, b)
}

// GetConcurrencyNum 获取同时下载的视频数量
func (s *Settings) GetConcurrencyNum() (int, error) {
	val, err := s.GetKey(concurrencyNumKey)
	if err != nil {
		return 1, err
	}
	num, err := strconv.Atoi(val)
	if err != nil {
		return 1, err
	}

	return num, nil
}

// SetConcurrencyNum 保存同时下载的视频数量
func (s *Settings) SetConcurrencyNum(num int) error {
	s.logger.Infof("Setting concurrency num to %d", num)
	return s.SetKey(concurrencyNumKey, strconv.Itoa(num))
}

// OpenDownloadLocation 打开下载历史中的文件位置.
func (s *Settings) OpenDownloadLocation(path string) error {
	target := strings.TrimSpace(path)
	if target == "" {
		return errors.New("文件路径为空")
	}

	info, err := os.Stat(target)
	if err == nil {
		if !info.IsDir() {
			target = filepath.Dir(target)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		target = filepath.Dir(target)
		if _, statErr := os.Stat(target); statErr != nil {
			return errors.New("文件所在目录不存在")
		}
	} else {
		return err
	}

	var cmd *exec.Cmd
	switch stdRuntime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", target)
	case "darwin":
		cmd = exec.Command("open", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}

	return cmd.Start()
}

// OpenLocalFile 使用系统默认程序打开下载历史中的本地文件或目录.
func (s *Settings) OpenLocalFile(path string) error {
	target := strings.TrimSpace(path)
	if target == "" {
		return errors.New("文件路径为空")
	}
	if _, err := os.Stat(target); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("文件不存在")
		}
		return err
	}

	var cmd *exec.Cmd
	switch stdRuntime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", target)
	case "darwin":
		cmd = exec.Command("open", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}

	return cmd.Start()
}

// Version is set by build flags, e.g. go build -ldflags="-X utils.Version=1.0.0"
var Version string

// GetVersion returns the current version of the application.
func (s *Settings) GetVersion() string {
	return Version
}

// OnShutdown is called when the application is shutting down
func (s *Settings) OnShutdown(_ context.Context) {
	if err := s.DB.Close(); err != nil {
		s.logger.Errorf("Failed to close settings DB: %v", err)
		return
	}
	s.logger.Info("Close BadgerDB and shutdown application")
}

func (s *Settings) ClearDownloadHistory(prefixKey string) error {
	prefix := []byte(prefixKey)
	return s.Update(func(txn *badger.Txn) error {
		keys := make([][]byte, 0)
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			keys = append(keys, it.Item().KeyCopy(nil))
		}
		for _, key := range keys {
			if err := txn.Delete(key); err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
				return err
			}
		}
		return nil
	})
}
