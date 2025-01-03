package backtrace

import (
	"testing"
	"time"

	"github.com/hchichi/backtrace/bk/ipgeo"
)

func TestRouteDetection(t *testing.T) {
	testCases := []struct {
		ip          string
		description string
	}{
		// 电信线路测试
		{"69.194.1.1", "CTGNet"},
		{"203.22.1.1", "CTGNet"},
		{"59.43.1.1", "CN2"},
		{"202.97.1.1", "163"},

		// 联通线路测试
		{"103.41.124.1", "CUG"},
		{"218.105.1.1", "CUII"},
		{"219.158.1.1", "联通"},

		// 移动线路测试
		{"223.120.19.1", "CMIN2"},
		{"223.118.1.1", "CMI"},

		// 特殊情况测试
		{"1.1.1.1", "Cloudflare"},
		{"8.8.8.8", "Google"},
	}

	t.Log("=== 开始线路检测测试 ===")
	for _, tc := range testCases {
		asn, desc := getRouteType(tc.ip)
		t.Logf("IP: %-15s ASN: %-10s 描述: %s", tc.ip, asn, desc)
		// 给 API 一些时间冷却
		time.Sleep(100 * time.Millisecond)
	}
	t.Log("=== 线路检测测试完成 ===")
}

func TestInternationalDetection(t *testing.T) {
	testCases := []struct {
		ip          string
		description string
	}{
		// 国内 IP
		{"202.97.1.1", "中国电信"},
		{"211.136.1.1", "中国移动"},

		// 国外 IP
		{"69.194.1.1", "美国"},
		{"203.22.1.1", "香港"},
		{"8.8.8.8", "美国"},
	}

	t.Log("=== 开始国内外检测测试 ===")
	for _, tc := range testCases {
		data, err := ipgeo.LeoIP(tc.ip, 5*time.Second)
		if err != nil {
			t.Logf("IP: %-15s 错误: %v", tc.ip, err)
			continue
		}
		t.Logf("IP: %-15s 国家: %-10s ASN: %-10s ISP: %s",
			tc.ip, data.CountryEn, data.Asnumber, data.Isp)
		// 给 API 一些时间冷却
		time.Sleep(100 * time.Millisecond)
	}
	t.Log("=== 国内外检测测试完成 ===")
}

func TestRouteQuality(t *testing.T) {
	testCases := []struct {
		asn      string
		priority int
	}{
		{"AS23764", 100}, // CTGNet
		{"AS4809a", 90},  // CN2 GIA
		{"AS10099", 85},  // CUG
		{"AS58807", 85},  // CMIN2
		{"AS4809b", 80},  // CN2 GT
		{"AS4837a", 75},  // 联通VIP
		{"AS9929", 75},   // CUII
		{"AS4134", 50},   // 163
		{"AS4837", 50},   // 联通
		{"AS9808", 50},   // CMI
		{"AS58453", 50},  // CMI
		{"AS1234", 0},    // 未知
	}

	t.Log("=== 开始线路优先级测试 ===")
	for _, tc := range testCases {
		priority := getRoutePriority(tc.asn)
		if priority != tc.priority {
			t.Errorf("ASN %s 优先级错误: 期望 %d, 得到 %d", tc.asn, tc.priority, priority)
		}
		t.Logf("ASN: %-10s 优先级: %d", tc.asn, priority)
	}
	t.Log("=== 线路优先级测试完成 ===")
}
