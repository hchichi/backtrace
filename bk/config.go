package backtrace

import (
	"fmt"
	"log"
	"os"
)

// 全局配置变量
var (
	// EnableLoger 是否启用日志
	EnableLoger = false

	// Logger 日志实例
	Logger = &SimpleLogger{log.New(os.Stderr, "", log.LstdFlags)}
)

// SimpleLogger 简单的日志实现
type SimpleLogger struct {
	*log.Logger
}

// Info 记录信息日志
func (l *SimpleLogger) Info(v ...interface{}) {
	if EnableLoger {
		l.Println(v...)
	}
}

// Sync 同步日志
func (l *SimpleLogger) Sync() error {
	return nil
}

// 颜色函数
func Red(str string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", str)
}

// InitLogger 初始化日志配置
func InitLogger() {
	EnableLoger = true
}
