package backtrace

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/nxtrace/NTrace-core/trace"
)

// PrintResult 遍历 traceroute 返回的结果，并根据 ASN、Whois 和 IP 前缀对关键信息进行高亮显示
func PrintResult(result *trace.Result) {
	for ttl, hops := range result.Hops {
		for _, hop := range hops {
			if hop.Address == nil {
				fmt.Printf("%d：*\n", ttl+1)
				continue
			}
			isBackbone := false
			if hop.Geo != nil {
				asn := hop.Geo.Asnumber
				whois := hop.Geo.Whois
				ip := hop.Address.String()
				if asn == "58807" || asn == "10099" || asn == "4809" || asn == "9929" || asn == "23764" ||
					strings.Contains(whois, "CTG-CN") || strings.Contains(whois, "[CNC-BACKBONE]") ||
					strings.Contains(whois, "[CUG-BACKBONE]") || strings.Contains(whois, "[CMIN2-NET]") ||
					strings.HasPrefix(ip, "59.43.") {
					isBackbone = true
				}
			}
			if isBackbone {
				color.New(color.FgHiYellow, color.Bold).Printf("%d - %s (AS%s)\n", ttl+1, hop.Address.String(), hop.Geo.Asnumber)
			} else {
				fmt.Printf("%d - %s (AS%s)\n", ttl+1, hop.Address.String(), hop.Geo.Asnumber)
			}
		}
	}
}
