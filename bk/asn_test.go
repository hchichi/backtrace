package backtrace

import (
	"fmt"
	"testing"
	"github.com/hchichi/backtrace/bk/ipgeo"
)

func TestRouteDetectionNew(t *testing.T) {
	// 启用测试模式
	EnableTestMode()
	defer DisableTestMode()

	// 测试特定场景
	testCases := []struct {
		name     string
		ip       string
		expected string
	}{
		{"北京电信163", "ipv4.pek-4134.endpoint.nxtrace.org", "电信163 [普通]"},
		{"上海联通169", "ipv4.sha-4837.endpoint.nxtrace.org", "联通169 [普通]"},
		{"广州移动CMI", "ipv4.can-9808.endpoint.nxtrace.org", "移动CMI [普通]"},
		{"合肥电信163", "61.132.163.68", "电信163 [普通]"},
		{"CTG回程", "69.194.1.1", "电信CTG回程 [顶级]"},
	}

	fmt.Println("\n=== 测试特定场景 ===")
	for _, tc := range testCases {
		plain, colored := getRouteType(tc.ip)
		if plain != tc.expected {
			t.Errorf("%s: 期望 %s, 得到 %s", tc.name, tc.expected, plain)
		}
		fmt.Printf("%s: %s\n", tc.name, colored)
	}
}

func TestInvalidIPsNew(t *testing.T) {
	// 启用测试模式
	EnableTestMode()
	defer DisableTestMode()

	invalidIPs := []struct {
		ip       string
		expected string
	}{
		{"invalid.ip", "未知"},
		{"999.999.999.999", "未知"},
		{"not.exist.ip.address", "未知"},
	}

	fmt.Println("\n=== 测试无效IP ===")
	for _, ip := range invalidIPs {
		plain, colored := getRouteType(ip.ip)
		if plain != ip.expected {
			t.Errorf("无效IP %s: 期望 %s, 得到 %s", ip.ip, ip.expected, plain)
		}
		fmt.Printf("无效IP (%s): %s\n", ip.ip, colored)
	}
}

func TestCTGBackhaul(t *testing.T) {
	// 启用测试模式
	EnableTestMode()
	defer DisableTestMode()

	testCases := []struct {
		name     string
		data     *ipgeo.IPGeoData
		expected bool
	}{
		{
			name: "CTG回程",
			data: &ipgeo.IPGeoData{
				Asnumber: "23764",
			},
			expected: true,
		},
		{
			name: "非CTG回程",
			data: &ipgeo.IPGeoData{
				Asnumber: "4134",
			},
			expected: false,
		},
	}

	fmt.Println("\n=== 测试CTG回程判断 ===")
	for _, tc := range testCases {
		result := isCTGBackhaul(tc.data)
		if result != tc.expected {
			t.Errorf("%s: 期望 %v, 得到 %v", tc.name, tc.expected, result)
		}
		fmt.Printf("%s: %v\n", tc.name, result)
	}
}

func TestNextTrace(t *testing.T) {
	// 测试nexttrace调用
	fmt.Println("\n=== 测试NextTrace调用 ===")
	ip := "1.1.1.1"
	output, err := useNextTrace(ip)
	if err != nil {
		fmt.Printf("NextTrace调用失败: %v\n", err)
	} else {
		fmt.Printf("NextTrace调用成功，输出前100个字符: %s\n", output[:min(len(output), 100)])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
