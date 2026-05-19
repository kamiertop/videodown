package logger

// Mode is set at build time via ldflags, e.g. -ldflags="-X github.com/kamiertop/videodown/logger.Mode=prod"
var Mode = "dev"

func IsProdMode() bool {
	return Mode == "prod"
}

func IsDevMode() bool {
	return !IsProdMode()
}
