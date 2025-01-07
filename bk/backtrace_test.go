package backtrace

import (
	"os"
	"testing"
)

func TestBackTrace(t *testing.T) {
	// 如果在 CI 环境中运行，跳过实际的网络测试
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}
	BackTrace()
}

// 添加 mock 测试，用于 CI 环境
func TestBackTraceMock(t *testing.T) {
	// 保存原始的 ips 和 names
	originalIps := ips
	originalNames := names

	// 使用较少的测试数据
	ips = []string{
		"127.0.0.1",
		"8.8.8.8",
	}
	names = []string{
		"本地测试",
		"公网测试",
	}

	// 测试完成后恢复原始数据
	defer func() {
		ips = originalIps
		names = originalNames
	}()

	// 执行测试
	BackTrace()
}
