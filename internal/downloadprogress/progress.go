package downloadprogress

import "github.com/kamiertop/videodown/internal/syncmap"

// WeightedPercent 把单个阶段的字节进度映射到整条任务进度，用于展示统一进度条
func WeightedPercent(start, weight float64, downloaded, total int64) float64 {
	if total <= 0 {
		return start
	}
	ratio := float64(downloaded) / float64(total)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	return start + ratio*weight
}

func ClampPercent(percent float64) float64 {
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

// Tracker holds the last reported percentage for each download key.
type Tracker struct {
	percent syncmap.Map[string, float64]
}

// Advance records a percentage unless it would move progress backwards.
func (t *Tracker) Advance(key string, percent float64) (float64, bool) {
	return t.percent.Update(key, func(previous float64, loaded bool) (float64, bool) {
		return percent, !loaded || !(percent < previous)
	})
}

// Store records a percentage regardless of the previous value.
func (t *Tracker) Store(key string, percent float64) {
	t.percent.Store(key, percent)
}

// Reset clears the percentage so the next download can start from zero.
func (t *Tracker) Reset(key string) {
	t.percent.Delete(key)
}
