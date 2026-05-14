package logger

import "os"

const (
	prodEnvKey   = "mode"
	prodEnvValue = "prod"
)

// mode=prod 时写文件并收敛日志级别；其它情况按开发模式输出到 stdout。
func IsProdMode() bool {
	return os.Getenv(prodEnvKey) == prodEnvValue
}

func IsDevMode() bool {
	return !IsProdMode()
}
