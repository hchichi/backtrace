package backtrace

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

type Result struct {
	i int
	s string
}

type ASNInfo struct {
	ASN         string `json:"asn"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Route       string `json:"route"`      // 路由信息
	BackboneIP  string `json:"backboneIP"` // 骨干网IP
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
	// 线路类型映射
	lineTypes = map[string]string{
		"CTG":    "电信CTGNet [精品线路]", // CTG是电信的国际业务品牌
		"CN2GIA": "电信CN2GIA [精品线路]",
		"CN2GT":  "电信CN2GT  [普通线路]",
		"163":    "电信163    [普通线路]",
		"9929":   "联通9929   [优质线路]",
		"4837":   "联通4837   [普通线路]",
		"CUII":   "联通CUII   [精品线路]",
		"CMIN2":  "移动CMIN2  [精品线路]",
		"CMI":    "移动CMI    [普通线路]",
	}
)

// GetTestIPs 返回测试用的IP列表
func GetTestIPs() []string {
	return ips
}

func removeDuplicates(elements []ASNInfo) []ASNInfo {
	encountered := make(map[string]bool)
	result := []ASNInfo{}

	for _, element := range elements {
		if encountered[element.ASN] {
			continue
		}
		encountered[element.ASN] = true
		result = append(result, element)
	}
	return result
}

func trace(ch chan Result, i int) {
	hops, err := Trace(net.ParseIP(ips[i]))
	if err != nil {
		s := fmt.Sprintf("%v %-15s %v", names[i], ips[i], err)
		ch <- Result{i, s}
		return
	}

	var asnInfos []ASNInfo
	lastValidHop := -1

	// 遍历所有hop
	for j, h := range hops {
		hasValidNode := false
		for _, n := range h.Nodes {
			info, err := getASNInfo(n.IP.String())
			if err != nil || info == nil {
				continue
			}

			hasValidNode = true
			asnInfos = append(asnInfos, *info)

			// 如果是主要运营商的ASN,更新lastValidHop
			if isMainCarrierASN(info.ASN) {
				lastValidHop = j
			}
		}
		if hasValidNode && lastValidHop == -1 {
			lastValidHop = j
		}
	}

	// 只处理到最后一个有效hop的路由信息
	if lastValidHop >= 0 {
		asnInfos = asnInfos[:lastValidHop+1]
	}

	if len(asnInfos) > 0 {
		var tempText string
		asnInfos = removeDuplicates(asnInfos)
		tempText += fmt.Sprintf("%v ", names[i])
		tempText += fmt.Sprintf("%-15s ", ips[i])

		for _, info := range asnInfos {
			lineType := getLineType(info)
			if lineType != "" {
				switch {
				case strings.Contains(lineType, "精品线路"):
					tempText += color.New(color.FgHiYellow, color.Bold).Sprintf("%s ", lineType)
				case strings.Contains(lineType, "优质线路"):
					tempText += color.New(color.FgHiGreen, color.Bold).Sprintf("%s ", lineType)
				default:
					tempText += color.New(color.FgWhite, color.Bold).Sprintf("%s ", lineType)
				}
			}
		}

		if tempText == (fmt.Sprintf("%v ", names[i]) + fmt.Sprintf("%-15s ", ips[i])) {
			tempText += fmt.Sprintf("%v", color.New(color.FgRed, color.Bold).Sprintf("检测不到已知线路的ASN"))
		}
		ch <- Result{i, tempText}
	} else {
		s := fmt.Sprintf("%v %-15s %v", names[i], ips[i], color.New(color.FgRed, color.Bold).Sprintf("检测不到回程路由节点的IP地址"))
		ch <- Result{i, s}
	}
}

// getASNInfo 使用nexttrace的ASN数据库获取ASN信息
func getASNInfo(ip string) (*ASNInfo, error) {
	cmd := exec.Command("nexttrace", "-q", "1", "--query", ip)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var info ASNInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// isMainCarrierASN 判断是否为主要运营商ASN
func isMainCarrierASN(asn string) bool {
	// 按运营商分组的ASN列表
	telecomASNs := []string{
		"AS23764", // 电信CTGNet
		"AS4809",  // 电信CN2
		"AS4134",  // 电信163
	}
	unicomASNs := []string{
		"AS9929",  // 联通9929
		"AS4837",  // 联通4837
		"AS10099", // 联通CUII
	}
	mobileASNs := []string{
		"AS58807", // 移动CMIN2
		"AS9808",  // 移动CMI
		"AS58453", // 移动CMI
	}

	// 检查是否属于任一运营商
	for _, asNum := range telecomASNs {
		if asn == asNum {
			return true
		}
	}
	for _, asNum := range unicomASNs {
		if asn == asNum {
			return true
		}
	}
	for _, asNum := range mobileASNs {
		if asn == asNum {
			return true
		}
	}
	return false
}

// getLineType 根据ASN信息判断线路类型
func getLineType(info ASNInfo) string {
	asnNum := strings.TrimPrefix(info.ASN, "AS")
	desc := strings.ToUpper(info.Description)

	switch asnNum {
	case "23764": // CTG是电信的国际业务品牌
		return lineTypes["CTG"]
	case "4809":
		// 通过描述判断是GIA还是GT
		if strings.Contains(desc, "GT") {
			return lineTypes["CN2GT"]
		}
		if strings.Contains(desc, "GIA") {
			return lineTypes["CN2GIA"]
		}
		// 默认为CN2GT
		return lineTypes["CN2GT"]
	case "4134":
		return lineTypes["163"]
	case "9929":
		return lineTypes["9929"]
	case "4837":
		return lineTypes["4837"]
	case "10099":
		return lineTypes["CUII"]
	case "58807":
		return lineTypes["CMIN2"]
	case "9808", "58453":
		return lineTypes["CMI"]
	default:
		return ""
	}
}
