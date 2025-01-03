package backtrace

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hchichi/backtrace/bk/ipgeo"
)

type Result struct {
	i int
	s string
}

var (
	ips = []string{
		"219.141.140.10", "202.106.195.68", "221.179.155.161",
		"202.96.209.133", "210.22.97.1", "211.136.112.200",
		"58.60.188.222", "210.21.196.6", "120.196.165.24",
		"61.139.2.69", "119.6.6.6", "211.137.96.205",
	}
	names = []string{
		"北京电信", "北京联通", "北京移动",
		"上海电信", "上海联通", "上海移动",
		"广州电信", "广州联通", "广州移动",
		"成都电信", "成都联通", "成都移动",
	}
	m = map[string]string{
		// 电信线路 (CTGNet≈CN2 GIA>CN2 GT>163直连)
		"AS23764": "电信CTGNet [极品]",  // CTGNet
		"AS4809a": "电信CN2 GIA [极品]", // CN2 GIA
		"AS4809b": "电信CN2 [普通]",     // CN2 GT
		"AS4134":  "电信163 [普通]",     // 163网

		// 联通线路 (CUG>9929>4837)
		"AS10099": "联通CUG [精品]",    // CUG
		"AS9929":  "联通CUII [优质]",   // CUII
		"AS4837a": "联通VIP [优质]",    // VIP
		"AS4837":  "联通4837 [普通]",   // 4837

		// 移动线路 (CMIN2>CMI)
		"AS58807": "移动CMIN2 [精品]", // CMIN2
		"AS9808":  "移动CMI [普通]",   // CMI
		"AS58453": "移动CMI [普通]",   // CMI
	}
	routePriority = map[string]int{
		"AS23764": 100, // CTGNet [极品]
		"AS4809a": 90,  // CN2 GIA [精品]
		"AS10099": 85,  // 联通 CUG [精品]
		"AS58807": 85,  // 移动 CMIN2 [精品]
		"AS4809b": 80,  // CN2 GT [优质]
		"AS4837a": 75,  // 联通 VIP [优质]
		"AS9929":  75,  // 联通 CUII [优质]
		"AS4134":  50,  // 电信 163 [普通]
		"AS4837":  50,  // 联通 169 [普通]
		"AS9808":  50,  // 移动 CMI [普通]
		"AS58453": 50,  // 移动 CMI [普通]
	}
	asn23764Prefixes = [][]string{
		{"102.217.164.", "102.217.166.", "121.59.100.", "121.59.106.", "121.59.109.", "121.59.120.", "121.59.122.", "121.59.124.", "121.59.140.", "121.59.144.", "121.59.153.", "147.78.133.", "147.78.135.", "154.198.1.", "154.198.2.", "183.91.51."},
		{"202.55.17.", "202.55.2.", "203.12.200.", "203.128.224.", "203.128.229.", "203.14.188.", "203.19.32.", "203.19.37.", "203.19.38.", "203.196.12.", "203.196.15.", "203.196.8.", "203.20.136.", "203.25.50.", "203.33.10.", "203.33.8."},
		{"203.34.194.", "203.34.196.", "203.34.199.", "203.80.136.", "203.81.0.", "203.81.6.", "203.82.18.", "203.86.101.", "203.86.126.", "218.185.241.", "5.154.128.", "5.154.144.", "5.154.154.", "5.154.156.", "63.140.14.", "63.140.9."},
		{"69.194.163.", "69.194.165.", "69.194.166.", "69.194.171.", "69.194.172.", "69.194.175.", "79.139.10.", "79.139.4.", "79.139.6.", "79.139.8.", "95.130.192.", "95.130.194.", "95.130.199."},
	}
	testModeEnabled = false
	testIPData = map[string]*ipgeo.IPGeoData{
		"59.43.1.1": {
			Asnumber: "4809",
			Prefix:   "59.43.0.0/16",
		},
		"202.97.1.1": {
			Asnumber: "4134",
		},
		"218.105.1.1": {
			Asnumber: "9929",
		},
		"223.120.19.1": {
			Asnumber: "58807",
		},
		"69.194.1.1": {
			Asnumber: "23764",
		},
	}
)

// EnableTestMode 启用测试模式
func EnableTestMode() {
	testModeEnabled = true
}

// DisableTestMode 禁用测试模式
func DisableTestMode() {
	testModeEnabled = false
}

// 移除重复元素的函数
func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{}   // 用于存储已经遇到的元素
	result := []string{}               // 存储去重后的结果
	for _, element := range elements { // 遍历切片中的元素
		if !encountered[element] { // 如果该元素尚未遇到
			encountered[element] = true      // 将该元素标记为已经遇到
			result = append(result, element) // 将该元素加入到结果切片中
		}
	}
	return result // 返回去重后的结果切片
}

// 检查切片中是否包含某个元素的函数
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// 获取 IP 地理信息
func getIPInfo(ip string) (*ipgeo.IPGeoData, error) {
	if testModeEnabled {
		if data, ok := testIPData[ip]; ok {
			return data, nil
		}
		return nil, nil
	}
	return ipgeo.LeoIP(ip, 5*time.Second)
}

// 判断是否为国外 IP
func isInternationalIP(country string) bool {
	return country != "" && country != "CN" && country != "China"
}

// 判断是否为 CN2 GIA
func isCN2GIA(data *ipgeo.IPGeoData) bool {
	if data.Asnumber == "4809" {
		return strings.HasPrefix(data.Prefix, "59.43.")
	}
	return false
}

// 判断是否为 CN2 GT
func isCN2GT(data *ipgeo.IPGeoData) bool {
	if data.Asnumber == "4809" {
		// CN2 GT: 59.43.*.* (国外) 或 202.97.*.* (国内)
		if strings.HasPrefix(data.Prefix, "59.43.") && isInternationalIP(data.CountryEn) {
			return true
		}
		if strings.HasPrefix(data.Prefix, "202.97.") && !isInternationalIP(data.CountryEn) {
			return true
		}
	}
	return false
}

// 判断是否为 CTGNet
func isCTGNet(data *ipgeo.IPGeoData) bool {
	if data.Asnumber == "23764" {
		// CTGNet 常见的骨干网 IP 开头为 69.194 和 203.22
		return strings.HasPrefix(data.Prefix, "69.194.") || strings.HasPrefix(data.Prefix, "203.22.")
	}
	return false
}

// 判断是否为联通 VIP
func isCUVIP(data *ipgeo.IPGeoData) bool {
	if data.Asnumber == "4837" {
		// 检查路由信息
		if routes, ok := data.Router["trace"]; ok {
			for _, route := range routes {
				// 检查是否经过美西节点
				// 已知的美西节点代码：
				// - SJC: San Jose, California
				// - LAX: Los Angeles, California
				// - SFO: San Francisco, California
				usWestCodes := []string{"SJC", "LAX", "SFO"}
				routeUpper := strings.ToUpper(route)
				for _, code := range usWestCodes {
					if strings.Contains(routeUpper, code) {
						return true
					}
				}
			}
		}
	}
	return false
}

// 获取线路类型
func getRouteType(ip string) (string, string) {
	data, err := getIPInfo(ip)
	if err != nil || data == nil {
		return "", "[未知]"
	}

	// 添加 AS 前缀
	asn := data.Asnumber
	if !strings.HasPrefix(asn, "AS") && asn != "" {
		asn = "AS" + asn
	}

	// 判断路由类型
	switch {
	case isCTGNet(data):
		return "AS23764", "电信CTGNet [极品]"
	
	case isCN2GIA(data):
		return "AS4809", "电信CN2 GIA [极品]"
	
	case isCN2GT(data):
		return "AS4809", "电信CN2 GT [普通]"
	
	case data.Asnumber == "4134":
		if strings.HasPrefix(data.Prefix, "202.97.") {
			return "AS4134", "电信163 [普通]"
		}
		return "AS4134", "电信ChinaNet [普通]"
	
	case data.Asnumber == "10099":
		return "AS10099", "联通CUG [精品]"
	
	case data.Asnumber == "9929":
		if strings.HasPrefix(data.Prefix, "218.105.") || strings.HasPrefix(data.Prefix, "210.51.") {
			return "AS9929", "联通CUII/CNCNet [优质]"
		}
		return "AS9929", "联通A网 [优质]"
	
	case isCUVIP(data):
		return "AS4837a", "联通VIP [优质]"
	
	case data.Asnumber == "4837":
		return "AS4837", "联通169 [普通]"
	
	case data.Asnumber == "58807":
		return "AS58807", "移动CMIN2 [精品]"
	
	case data.Asnumber == "9808":
		return "AS9808", "移动CMI [优质]"
	
	case data.Asnumber == "58453":
		return "AS58453", "移动CMI [普通]"
	}

	return asn, data.Owner + " [未知]"
}

// ipAsn 函数返回 IP 对应的 ASN
func ipAsn(ip string) string {
	asn, _ := getRouteType(ip)
	return asn
}

// 按优先级排序 ASN 列表
func sortASNsByPriority(asns []string) []string {
	// 创建一个新的切片来存储结果
	result := make([]string, len(asns))
	copy(result, asns)

	// 将 AS4809 转换为 AS4809a
	for i, asn := range result {
		if asn == "AS4809" {
			result[i] = "AS4809a"
		}
	}

	// 按优先级排序
	sort.SliceStable(result, func(i, j int) bool {
		pi := routePriority[result[i]]
		pj := routePriority[result[j]]
		if pi == pj {
			return result[i] < result[j] // 相同优先级按 ASN 字母顺序排序
		}
		return pi > pj
	})

	return result
}

// 获取 ASN 对应的优先级
func getRoutePriority(asn string) int {
	switch asn {
	case "AS23764": // CTGNet [极品]
		return 100
	case "AS4809", "AS4809a": // CN2 GIA [精品]
		return 90
	case "AS4809b": // CN2 GT [优质]
		return 80
	case "AS10099": // CUG [精品]
		return 85
	case "AS58807": // CMIN2 [精品]
		return 85
	case "AS4837a": // 联通 VIP [优质]
		return 75
	case "AS9929": // CUII [优质]
		return 75
	case "AS4134", "AS4837", "AS9808", "AS58453": // 普通线路
		return 50
	default:
		return 0
	}
}

// 主跟踪函数
func trace(ch chan Result, i int) {
	hops, err := Trace(net.ParseIP(ips[i]))
	if err != nil {
		s := fmt.Sprintf("%v %-15s %v", names[i], ips[i], err)
		ch <- Result{i, s}
		return
	}

	var asns []string
	var geoData *ipgeo.IPGeoData

	for _, h := range hops {
		for _, n := range h.Nodes {
			// 使用 Leo API 查询 IP 信息
			geoData, err = getIPInfo(n.IP.String())
			if err == nil && geoData != nil && geoData.Asnumber != "" {
				asns = append(asns, "AS"+geoData.Asnumber)
			} else {
				// 如果 API 查询失败，使用本地判断
				asn := ipAsn(n.IP.String())
				if asn != "" {
					asns = append(asns, asn)
				}
			}
		}
	}

	if asns != nil && len(asns) > 0 {
		var tempText string
		asns = removeDuplicates(asns)
		asns = sortASNsByPriority(asns) // 按优先级排序

		tempText += fmt.Sprintf("%v ", names[i])
		tempText += fmt.Sprintf("%-15s ", ips[i])

		for _, asn := range asns {
			asnDescription, exists := m[asn]
			if !exists {
				if strings.HasPrefix(asn, "AS") {
					tempText += color.New(color.FgHiWhite, color.Bold).Sprintf("%s ", asn)
				}
				continue
			}

			// 根据线路类型选择颜色
			switch {
			// 电信线路
			case asn == "AS23764": // CTGNet [极品]
				tempText += color.New(color.FgHiYellow, color.Bold, color.BgRed).Sprint(asnDescription) + " "
			case asn == "AS4809": // CN2 GIA [极品]
				tempText += color.New(color.FgHiYellow, color.Bold, color.BgHiRed).Sprint(asnDescription) + " "
			case asn == "AS4809b": // CN2 GT [优质]
				tempText += color.New(color.FgYellow, color.Bold).Sprint(asnDescription) + " "
			case asn == "AS4134": // 163 [普通]
				tempText += color.New(color.FgHiWhite).Sprint(asnDescription) + " "

			// 联通线路
			case asn == "AS10099": // CUG [精品]
				tempText += color.New(color.FgHiBlue, color.Bold, color.BgHiMagenta).Sprint(asnDescription) + " "
			case asn == "AS9929": // CUII/CNCNet [优质]
				tempText += color.New(color.FgBlue, color.Bold).Sprint(asnDescription) + " "
			case asn == "AS4837a": // 联通 VIP [优质]
				tempText += color.New(color.FgHiBlue, color.Bold).Sprint(asnDescription) + " "
			case asn == "AS4837": // 联通 169 [普通]
				tempText += color.New(color.FgHiBlue).Sprint(asnDescription) + " "

			// 移动线路
			case asn == "AS58807": // CMIN2 [精品]
				tempText += color.New(color.FgHiGreen, color.Bold, color.BgHiCyan).Sprint(asnDescription) + " "
			case asn == "AS9808": // CMI [优质]
				tempText += color.New(color.FgGreen, color.Bold).Sprint(asnDescription) + " "
			case asn == "AS58453": // CMI [普通]
				tempText += color.New(color.FgGreen).Sprint(asnDescription) + " "

			// 未知线路
			default:
				tempText += color.New(color.FgHiWhite, color.Faint).Sprintf("%s ", asnDescription)
			}
		}

		if tempText == (fmt.Sprintf("%v ", names[i]) + fmt.Sprintf("%-15s ", ips[i])) {
			tempText += color.New(color.FgRed, color.Bold).Sprint("检测不到已知线路的ASN")
		}
		ch <- Result{i, tempText}
	} else {
		s := fmt.Sprintf("%v %-15s %v", names[i], ips[i], color.New(color.FgRed, color.Bold).Sprint("检测不到回程路由节点的IP地址"))
		ch <- Result{i, s}
	}
}

// hasAnyPrefix checks if the IP starts with any prefix in the given slice of slices
func hasAnyPrefix(ip string, prefixSlices [][]string) bool {
	for _, prefixes := range prefixSlices {
		for _, prefix := range prefixes {
			if strings.HasPrefix(ip, prefix) {
				return true
			}
		}
	}
	return false
}
