package backtrace

import (
	"testing"
	"time"
)

func TestTrace(t *testing.T) {
	ch := make(chan Result)

	// 测试多个 IP
	for i := 0; i < len(ips); i++ {
		go trace(ch, i)
	}

	// 收集结果
	results := make([]Result, len(ips))
	for i := 0; i < len(ips); i++ {
		select {
		case result := <-ch:
			results[result.i] = result
		case <-time.After(30 * time.Second):
			t.Errorf("测试超时: index %d", i)
			return
		}
	}

	// 验证结果
	for i, result := range results {
		if result.s == "" {
			t.Errorf("IP %s (%s) 没有返回结果", ips[i], names[i])
		}
	}
}

func TestIPAsn(t *testing.T) {
	// 启用测试模式
	EnableTestMode()
	defer DisableTestMode()

	testCases := []struct {
		ip       string
		expected string
	}{
		{"59.43.1.1", "AS4809"},
		{"202.97.1.1", "AS4134"},
		{"218.105.1.1", "AS9929"},
		{"223.120.19.1", "AS58807"},
		{"69.194.1.1", "AS23764"},
		{"1.1.1.1", ""}, // 未知 IP
	}

	for _, tc := range testCases {
		got := ipAsn(tc.ip)
		if got != tc.expected {
			t.Errorf("ipAsn(%s) = %s; 期望 %s", tc.ip, got, tc.expected)
		}
	}
}

func TestSortASNsByPriority(t *testing.T) {
	testCases := []struct {
		input    []string
		expected []string
	}{
		{
			[]string{"AS4134", "AS23764", "AS4809"},
			[]string{"AS23764", "AS4809a", "AS4134"},
		},
		{
			[]string{"AS9929", "AS4837", "AS10099"},
			[]string{"AS10099", "AS9929", "AS4837"},
		},
	}

	for _, tc := range testCases {
		got := sortASNsByPriority(tc.input)
		if len(got) != len(tc.expected) {
			t.Errorf("长度不匹配: 得到 %v, 期望 %v", got, tc.expected)
			continue
		}
		for i := range got {
			if got[i] != tc.expected[i] {
				t.Errorf("排序错误: 位置 %d, 得到 %s, 期望 %s", i, got[i], tc.expected[i])
			}
		}
	}
}
