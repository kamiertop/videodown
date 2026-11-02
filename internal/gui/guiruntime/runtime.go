package guiruntime

import "time"

// Runtime schedules work outside the view, and delivers completion on the UI
// thread. After's returned callback cancels a scheduled timer.
type Runtime struct {
	Run          func(work func() error, done func(error))
	Background   func(work func() error, done func(error))
	Image        func(work func() error, done func(error))
	After        func(time.Duration, func()) func()
	OpenFile     func(string) error
	OpenLocation func(string) error
}
