package backtrace

import (
	"net"
	"strings"
)

// RouteAnalyzer 路由分析器
type RouteAnalyzer struct {
	hops       []net.IP
	asns       map[string]bool
	routerInfo map[string][]string // 路由表信息
}

// NewRouteAnalyzer 创建新的路由分析器
func NewRouteAnalyzer() *RouteAnalyzer {
	return &RouteAnalyzer{
		asns:       make(map[string]bool),
		routerInfo: make(map[string][]string),
	}
}

// AddHop 添加一个路由跳点
func (ra *RouteAnalyzer) AddHop(ip net.IP) {
	ra.hops = append(ra.hops, ip)
}

// AddASN 添加一个ASN
func (ra *RouteAnalyzer) AddASN(asn string) {
	ra.asns[asn] = true
}

// AddRouterInfo 添加路由表信息
func (ra *RouteAnalyzer) AddRouterInfo(asn string, routes []string) {
	ra.routerInfo[asn] = routes
}

// HasConvergenceLayer 检测是否存在汇聚层
func (ra *RouteAnalyzer) HasConvergenceLayer() bool {
	// 检查是否存在多个运营商的ASN
	carriers := make(map[string]bool)
	for asn := range ra.asns {
		switch {
		case strings.HasPrefix(asn, "4134") || strings.HasPrefix(asn, "4809") || asn == "23764":
			carriers["CT"] = true // 电信
		case strings.HasPrefix(asn, "4837") || strings.HasPrefix(asn, "9929") || asn == "10099":
			carriers["CU"] = true // 联通
		case strings.HasPrefix(asn, "9808") || strings.HasPrefix(asn, "58453") || asn == "58807":
			carriers["CM"] = true // 移动
		}
	}
	return len(carriers) > 1
}

// GetRouteQuality 获取路由质量
func (ra *RouteAnalyzer) GetRouteQuality() string {
	for asn := range ra.asns {
		switch {
		// 精品线路判断
		case asn == "23764" || // CTGNet
			asn == "58807" || // CMIN2
			(asn == "4809" && ra.hasCN2GIACharacteristics()): // CN2 GIA
			return "精品线路"

		// 优质线路判断
		case asn == "10099" || // CUG
			asn == "9929" || // CUII
			(asn == "4809" && !ra.hasCN2GIACharacteristics()) || // CN2 GT
			ra.hasQualityCharacteristics(): // 其他优质线路特征
			return "优质线路"

		// 普通线路判断
		case asn == "4134" || // 电信163
			asn == "4837" || // 联通4837
			asn == "9808" || // 移动CMI
			asn == "58453": // 移动CMI
			return "普通线路"
		}
	}
	return "未知线路"
}

// hasCN2GIACharacteristics 判断是否为CN2 GIA特征
func (ra *RouteAnalyzer) hasCN2GIACharacteristics() bool {
	// 检查IP特征
	for _, ip := range ra.hops {
		if ip != nil {
			ipStr := ip.String()
			// CN2 GIA特征IP
			if strings.HasPrefix(ipStr, "59.43.80.") {
				return true
			}
		}
	}

	// 检查路由表中是否包含AS4134
	if _, has4134 := ra.asns["4134"]; has4134 {
		return false
	}

	// 检查路由表信息
	for asn, routes := range ra.routerInfo {
		if asn == "4809" {
			for _, route := range routes {
				if strings.Contains(route, "4134") {
					return false
				}
			}
		}
	}

	return true
}

// hasQualityCharacteristics 判断是否具有优质线路特征
func (ra *RouteAnalyzer) hasQualityCharacteristics() bool {
	for _, ip := range ra.hops {
		if ip != nil {
			ipStr := ip.String()
			// 联通9929特征IP
			if strings.HasPrefix(ipStr, "218.105.") || strings.HasPrefix(ipStr, "210.51.") {
				return true
			}
			// CUG特征IP
			if strings.HasPrefix(ipStr, "219.158.113.") {
				return true
			}
			// CTG特征IP
			if strings.HasPrefix(ipStr, "69.194.") || strings.HasPrefix(ipStr, "203.22.") {
				return true
			}
		}
	}
	return false
}

// GetCarrierType 获取运营商类型
func (ra *RouteAnalyzer) GetCarrierType() string {
	for asn := range ra.asns {
		switch {
		case strings.HasPrefix(asn, "4134") || strings.HasPrefix(asn, "4809") || asn == "23764":
			return "电信"
		case strings.HasPrefix(asn, "4837") || strings.HasPrefix(asn, "9929") || asn == "10099":
			return "联通"
		case strings.HasPrefix(asn, "9808") || strings.HasPrefix(asn, "58453") || asn == "58807":
			return "移动"
		}
	}
	return "未知"
}

// GetDetailedInfo 获取详细信息
func (ra *RouteAnalyzer) GetDetailedInfo() string {
	var result string

	// 添加运营商信息
	carrier := ra.GetCarrierType()
	result += "运营商: " + carrier + "\n"

	// 添加线路质量
	quality := ra.GetRouteQuality()
	result += "线路质量: " + quality + "\n"

	// 添加汇聚层信息
	if ra.HasConvergenceLayer() {
		result += "警告: 检测到汇聚层，可能影响线路判断准确性\n"
	}

	// 添加路由信息
	if len(ra.routerInfo) > 0 {
		result += "路由信息:\n"
		for asn, routes := range ra.routerInfo {
			result += "  AS" + asn + ": " + strings.Join(routes, " -> ") + "\n"
		}
	}

	return result
}
