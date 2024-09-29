package backtrace

import (
	"fmt"
	"net"
	"strings"

	. "github.com/oneclickvirt/defaultset"
)

type Result struct {
	i int
	s string
}

var (
	ips = []string{
		// "219.141.136.12", "202.106.50.1",
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
		// [] 前的字符串个数，中文占2个字符串
		"AS4809a": "电信CN2GIA [精品线路]",
		"AS4809b": "电信CN2GT  [优质线路]",
		"AS4134":  "电信163    [普通线路]",
		"AS9929":  "联通9929   [优质线路]",
		"AS4837":  "联通4837   [普通线路]",
		"AS58807": "移动CMIN2  [精品线路]",
		"AS9808":  "移动CMI    [普通线路]",
		"AS58453": "移动CMI    [普通线路]",
		"AS23764": "电信CTG    [精品线路]",
	}
	asn23764Prefixes = [][]string{
		{"102.217.164.", "102.217.166.", "121.59.100.", "121.59.106.", "121.59.109.", "121.59.120.", "121.59.122.", "121.59.124.", "121.59.140.", "121.59.144.", "121.59.153.", "147.78.133.", "147.78.135.", "154.198.1.", "154.198.2.", "183.91.51."},
		{"202.55.17.", "202.55.2.", "203.12.200.", "203.128.224.", "203.128.229.", "203.14.188.", "203.19.32.", "203.19.37.", "203.19.38.", "203.196.12.", "203.196.15.", "203.196.8.", "203.20.136.", "203.25.50.", "203.33.10.", "203.33.8."},
		{"203.34.194.", "203.34.196.", "203.34.199.", "203.80.136.", "203.81.0.", "203.81.6.", "203.82.18.", "203.86.101.", "203.86.126.", "218.185.241.", "5.154.128.", "5.154.144.", "5.154.154.", "5.154.156.", "63.140.14.", "63.140.9."},
		{"69.194.163.", "69.194.165.", "69.194.166.", "69.194.171.", "69.194.172.", "69.194.175.", "79.139.10.", "79.139.4.", "79.139.6.", "79.139.8.", "95.130.192.", "95.130.194.", "95.130.199."},
	}
)

// 移除重复元素的函数
func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{} // 用于存储已经遇到的元素
	result := []string{}             // 存储去重后的结果
	for _, element := range elements { // 遍历切片中的元素
		if !encountered[element] {    // 如果该元素尚未遇到
			encountered[element] = true  // 将该元素标记为已经遇到
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

// 主跟踪函数
func trace(ch chan Result, i int) {
	hops, err := Trace(net.ParseIP(ips[i]))
	if err != nil {
		s := fmt.Sprintf("%v %-15s %v", names[i], ips[i], err)
		ch <- Result{i, s}
		return
	}
	var asns []string
	for _, h := range hops {
		for _, n := range h.Nodes {
			asn := ipAsn(n.IP.String())
			if asn != "" {
				asns = append(asns, asn)
			}
		}
	}

	if asns != nil && len(asns) > 0 {
		var tempText string
		asns = removeDuplicates(asns)
		tempText += fmt.Sprintf("%v ", names[i])
		hasAS4134 := contains(asns, "AS4134")
		hasAS4809 := contains(asns, "AS4809")
		hasAS23764 := contains(asns, "AS23764")

		// 优先检查 CTG
		if hasAS23764 {
			asns = append([]string{"AS23764"}, asns...)
		} else if hasAS4134 && hasAS4809 {
			asns = append([]string{"AS4809b"}, asns...)
		} else if hasAS4809 {
			asns = append([]string{"AS4809a"}, asns...)
		}

		tempText += fmt.Sprintf("%-15s ", ips[i])
		for _, asn := range asns {
			asnDescription, exists := m[asn]
			if !exists {
				continue
			}
			switch asn {
			case "", "AS4809": // 被 AS4809a 和 AS4809b 替代了
				continue
			case "AS9929":
				if !strings.Contains(tempText, asnDescription) {
					tempText += HiIndianRed(asnDescription) + " "
				}
			case "AS58807":
				if !strings.Contains(tempText, asnDescription) {
					tempText += HiLightBlue(asnDescription) + " "
				}
			case "AS4809a", "AS23764":
				if !strings.Contains(tempText, asnDescription) {
					tempText += HiGoldenrod(asnDescription) + " "
				}
			case "AS4809b":
				if !strings.Contains(tempText, asnDescription) {
					tempText += HiDarkOrange(asnDescription) + " "
				}
			default:
				if !strings.Contains(tempText, asnDescription) {
					tempText += HiSlateGray(asnDescription) + " "
				}
			}
		}
		if tempText == (fmt.Sprintf("%v ", names[i]) + fmt.Sprintf("%-15s ", ips[i])) {
			tempText += fmt.Sprintf("%v", Red("检测不到已知线路的ASN"))
		}
		ch <- Result{i, tempText}
	} else {
		s := fmt.Sprintf("%v %-15s %v", names[i], ips[i], Red("检测不到回程路由节点的IP地址"))
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

// ipAsn function
func ipAsn(ip string) string {
	switch {
	case strings.HasPrefix(ip, "59.43"):
		return "AS4809"
	case strings.HasPrefix(ip, "202.97"):
		return "AS4134"
	case strings.HasPrefix(ip, "218.105"), strings.HasPrefix(ip, "210.51"):
		return "AS9929"
	case strings.HasPrefix(ip, "219.158"):
		return "AS4837"
	case strings.HasPrefix(ip, "223.120.19"), strings.HasPrefix(ip, "223.120.17"), strings.HasPrefix(ip, "223.120.16"),
		strings.HasPrefix(ip, "223.120.140"), strings.HasPrefix(ip, "223.120.130"), strings.HasPrefix(ip, "223.120.131"),
		strings.HasPrefix(ip, "223.120.141"):
		return "AS58807"
	case strings.HasPrefix(ip, "223.118"), strings.HasPrefix(ip, "223.119"), strings.HasPrefix(ip, "223.120"), strings.HasPrefix(ip, "223.121"):
		return "AS58453"
	case hasAnyPrefix(ip, asn23764Prefixes):
		return "AS23764"
	default:
		return ""
	}
}