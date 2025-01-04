package backtrace

import (
	"fmt"
	"os/exec"
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

// ISP类型定义
type ISPType string

const (
	CT163  ISPType = "电信163"
	CTCN2  ISPType = "电信CN2"
	CU169  ISPType = "联通169"
	CU9929 ISPType = "联通9929"
	CM     ISPType = "移动CMI"
	CMIN2  ISPType = "移动CMIN2"
	EDU    ISPType = "教育网"
	CST    ISPType = "科技网"
)

// ISP信息
type ISPInfo struct {
	ISPName ISPType
	IP      string
	IPv6    string // 当前未使用，可以考虑移除
}

// 骨干网点信息
type BackBoneInfo struct {
	Location string
	ISPs     map[ISPType]ISPInfo
}

// 骨干网点信息集合
var backBoneCollection = []BackBoneInfo{
	{
		Location: "北京",
		ISPs: map[ISPType]ISPInfo{
			CT163: {ISPName: CT163, IP: "ipv4.pek-4134.endpoint.nxtrace.org"},
			CU169: {ISPName: CU169, IP: "ipv4.pek-4837.endpoint.nxtrace.org"},
			CM:    {ISPName: CM, IP: "ipv4.pek-9808.endpoint.nxtrace.org"},
		},
	},
	{
		Location: "上海",
		ISPs: map[ISPType]ISPInfo{
			CT163: {ISPName: CT163, IP: "ipv4.sha-4134.endpoint.nxtrace.org"},
			CU169: {ISPName: CU169, IP: "ipv4.sha-4837.endpoint.nxtrace.org"},
			CM:    {ISPName: CM, IP: "ipv4.sha-9808.endpoint.nxtrace.org"},
		},
	},
	{
		Location: "广州",
		ISPs: map[ISPType]ISPInfo{
			CT163: {ISPName: CT163, IP: "ipv4.can-4134.endpoint.nxtrace.org"},
			CU169: {ISPName: CU169, IP: "ipv4.can-4837.endpoint.nxtrace.org"},
			CM:    {ISPName: CM, IP: "ipv4.can-9808.endpoint.nxtrace.org"},
		},
	},
	{
		Location: "合肥",
		ISPs: map[ISPType]ISPInfo{
			CT163: {ISPName: CT163, IP: "61.132.163.68"},
			EDU:   {ISPName: EDU, IP: "ipv4.hfe-4538.endpoint.nxtrace.org."},
			CST:   {ISPName: CST, IP: "ipv4.hfe-7497.endpoint.nxtrace.org."},
		},
	},
}

var (
	ips = []string{
		// 北京三网
		"ipv4.pek-4134.endpoint.nxtrace.org", // 北京电信
		"ipv4.pek-4837.endpoint.nxtrace.org", // 北京联通
		"ipv4.pek-9808.endpoint.nxtrace.org", // 北京移动

		// 上海三网
		"ipv4.sha-4134.endpoint.nxtrace.org", // 上海电信
		"ipv4.sha-4837.endpoint.nxtrace.org", // 上海联通
		"ipv4.sha-9808.endpoint.nxtrace.org", // 上海移动

		// 广州三网
		"ipv4.can-4134.endpoint.nxtrace.org", // 广州电信
		"ipv4.can-4837.endpoint.nxtrace.org", // 广州联通
		"ipv4.can-9808.endpoint.nxtrace.org", // 广州移动

		// 合肥电信
		"61.132.163.68",  // 合肥电信主要
		"202.102.213.68", // 合肥电信备用
	}

	names = []string{
		"北京电信", "北京联通", "北京移动",
		"上海电信", "上海联通", "上海移动",
		"广州电信", "广州联通", "广州移动",
		"合肥电信", "合肥电信(备用)",
	}
	testModeEnabled = false
	testIPData      = map[string]*ipgeo.IPGeoData{
		// CTG测试数据
		"69.194.1.1": {
			Asnumber: "23764",
			Prefix:   "69.194.0.0/16",
		},

		// 电信测试数据
		"ipv4.pek-4134.endpoint.nxtrace.org": {
			Asnumber: "4134",
			Prefix:   "202.97.0.0/16",
		},
		"ipv4.sha-4134.endpoint.nxtrace.org": {
			Asnumber: "4134",
			Prefix:   "202.97.0.0/16",
		},
		"ipv4.can-4134.endpoint.nxtrace.org": {
			Asnumber: "4134",
			Prefix:   "202.97.0.0/16",
		},
		"61.132.163.68": {
			Asnumber: "4134",
			Prefix:   "61.132.0.0/16",
		},
		"202.102.213.68": {
			Asnumber: "4134",
			Prefix:   "202.102.0.0/16",
		},

		// 联通测试数据
		"ipv4.pek-4837.endpoint.nxtrace.org": {
			Asnumber: "4837",
			Prefix:   "219.158.0.0/16",
		},
		"ipv4.sha-4837.endpoint.nxtrace.org": {
			Asnumber: "4837",
			Prefix:   "219.158.0.0/16",
		},
		"ipv4.can-4837.endpoint.nxtrace.org": {
			Asnumber: "4837",
			Prefix:   "219.158.0.0/16",
		},

		// 移动测试数据
		"ipv4.pek-9808.endpoint.nxtrace.org": {
			Asnumber: "9808",
			Prefix:   "211.136.0.0/16",
		},
		"ipv4.sha-9808.endpoint.nxtrace.org": {
			Asnumber: "9808",
			Prefix:   "211.136.0.0/16",
		},
		"ipv4.can-9808.endpoint.nxtrace.org": {
			Asnumber: "9808",
			Prefix:   "211.136.0.0/16",
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

// 线路质量定义
type RouteQuality struct {
	ASN         string
	Name        string
	Description string
	Color       color.Attribute
}

// 线路质量映射
var routeQualities = []RouteQuality{
	// 电信线路
	{"23764", "AS23764", "电信CTGNet [顶级]", color.FgHiYellow},
	{"4809", "AS4809a", "电信CN2 GIA [顶级]", color.FgHiYellow},
	{"4809", "AS4809b", "电信CN2 GT [普通]", color.FgYellow},
	{"4134", "AS4134", "电信163 [普通]", color.FgHiBlack},

	// 联通线路
	{"9929", "AS9929", "联通CUII [优质]", color.FgHiRed},
	{"10099", "AS10099", "联通CUG [优质]", color.FgHiRed},
	{"4837", "AS4837", "联通169 [普通]", color.FgHiBlack},

	// 移动线路
	{"58807", "AS58807", "移动CMIN2 [优质]", color.FgHiBlue},
	{"9808", "AS9808", "移动CMI [普通]", color.FgHiBlack},
	{"58453", "AS58453", "移动CMI [普通]", color.FgHiBlack},
}

// 判断是否为国外 IP
func isInternationalIP(data *ipgeo.IPGeoData) bool {
	return data.CountryEn != "" && data.CountryEn != "CN" && data.CountryEn != "China"
}

// 判断是否为 CN2 GIA
func isCN2GIA(data *ipgeo.IPGeoData) bool {
	return data.Asnumber == "4809" && strings.HasPrefix(data.Prefix, "59.43.")
}

// 判断是否为 CN2 GT
func isCN2GT(data *ipgeo.IPGeoData) bool {
	if data.Asnumber == "4809" {
		// CN2 GT: 59.43.*.* (国外) 或 202.97.*.* (国内)
		if strings.HasPrefix(data.Prefix, "59.43.") && isInternationalIP(data) {
			return true
		}
		if strings.HasPrefix(data.Prefix, "202.97.") && !isInternationalIP(data) {
			return true
		}
	}
	return false
}

// 判断是否为 CTGNet
func isCTGNet(data *ipgeo.IPGeoData) bool {
	return data.Asnumber == "23764"
}

// 判断是否为 CTG 回程
func isCTGBackhaul(data *ipgeo.IPGeoData) bool {
	// CTG回程特征：AS23764 + 特定网段
	if data.Asnumber == "23764" {
		// 这里可以添加更多CTG回程的特征判断
		// 例如特定的IP段、地理位置等
		return true
	}
	return false
}

// 检查nexttrace是否可用
func isNextTraceAvailable() bool {
	_, err := exec.LookPath("nexttrace")
	return err == nil
}

// 使用nexttrace进行路由检测
func useNextTrace(ip string) (string, error) {
	if !isNextTraceAvailable() {
		return "", fmt.Errorf("nexttrace not installed")
	}
	cmd := exec.Command("nexttrace", "--tcp", "--no-rdns", ip)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// 获取路由类型
func getRouteType(ip string) (string, string) {
	// 首先尝试使用nexttrace
	if isNextTraceAvailable() {
		if output, err := useNextTrace(ip); err == nil {
			// 解析nexttrace输出
			routeInfo, err := parseNextTraceOutput(output)
			if err == nil {
				return routeInfo.Description, routeInfo.ColoredDescription
			}
		}
	}

	// 降级使用本地路由检测
	data, err := getIPInfo(ip)
	if err != nil || data == nil {
		return "未知", color.New(color.FgRed).Sprint("未知")
	}

	// 获取ASN信息
	asn := data.Asnumber
	if asn == "" {
		return "未知", color.New(color.FgRed).Sprint("未知")
	}

	// 特殊线路判断
	if isCTGBackhaul(data) {
		return "电信CTG回程 [顶级]", color.New(color.FgHiYellow).Sprint("电信CTG回程 [顶级]")
	}

	if isCTGNet(data) {
		for _, q := range routeQualities {
			if q.ASN == "23764" {
				return q.Description, color.New(q.Color).Sprint(q.Description)
			}
		}
	}

	if isCN2GIA(data) {
		for _, q := range routeQualities {
			if q.ASN == "4809" && q.Name == "AS4809a" {
				return q.Description, color.New(q.Color).Sprint(q.Description)
			}
		}
	}

	if isCN2GT(data) {
		for _, q := range routeQualities {
			if q.ASN == "4809" && q.Name == "AS4809b" {
				return q.Description, color.New(q.Color).Sprint(q.Description)
			}
		}
	}

	// 普通线路判断
	for _, q := range routeQualities {
		if q.ASN == asn {
			return q.Description, color.New(q.Color).Sprint(q.Description)
		}
	}

	return "未知", color.New(color.FgRed).Sprint("未知")
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

// 获取路由优先级
func getRoutePriority(asn string) int {
	// 移除AS前缀
	if len(asn) > 2 && asn[:2] == "AS" {
		asn = asn[2:]
	}

	// 特殊线路优先级
	switch asn {
	case "23764": // CTG
		return 100
	case "4809a": // CN2 GIA
		return 90
	case "10099", "58807": // 电信CN2
		return 85
	case "4809b": // CN2 GT
		return 80
	case "4837a", "9929": // 联通精品
		return 75
	case "4134", "4837", "9808", "58453": // 普通线路
		return 50
	default:
		return 0
	}
}

// 按优先级排序 ASN 列表
func sortASNsByPriority(asns []string) []string {
	sort.Slice(asns, func(i, j int) bool {
		return getRoutePriority(asns[i]) > getRoutePriority(asns[j])
	})
	return asns
}

// 主跟踪函数
func trace(ch chan Result, i int) {
	defer func() {
		if err := recover(); err != nil {
			ch <- Result{i: i, s: fmt.Sprintf("[错误：%v]", err)}
		}
	}()

	if i >= len(ips) {
		ch <- Result{i: i, s: "[超出范围]"}
		return
	}

	ip := ips[i]
	data, err := getIPInfo(ip)
	if err != nil {
		ch <- Result{i: i, s: fmt.Sprintf("[错误：%v]", err)}
		return
	}

	if data == nil {
		ch <- Result{i: i, s: "[未知]"}
		return
	}

	asn, desc := getRouteType(ip)
	ch <- Result{i: i, s: fmt.Sprintf("%s %s", asn, desc)}
}

// 添加nexttrace输出解析函数
func parseNextTraceOutput(output string) (*RouteInfo, error) {
	// 实现nexttrace输出解析逻辑
	// 返回解析后的路由信息
	return nil, fmt.Errorf("not implemented")
}

type RouteInfo struct {
	Description        string
	ColoredDescription string
	ASN                string
	ISP                string
	// 其他需要的字段
}
