package backtrace

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type PrinterConfig struct {
	EnableHideDstIP bool
	DestIP          string
}

// ipAsn 获取IP地址对应的ASN信息
func ipAsn(ip string) string {
	info, err := getASNInfo(ip)
	if err != nil || info == nil {
		return ""
	}
	return info.ASN
}

func RealtimePrinter(hop *Hop, ttl int, config *PrinterConfig) {
	fmt.Printf("%s  ", color.New(color.FgHiYellow, color.Bold).Sprintf("%-2d", ttl+1))

	if len(hop.Nodes) == 0 {
		fmt.Fprintf(color.Output, "%s\n",
			color.New(color.FgWhite, color.Bold).Sprintf("*"),
		)
		return
	}

	var blockDisplay = false
	for _, node := range hop.Nodes {
		if blockDisplay {
			fmt.Printf("%4s", "")
		}

		ip := node.IP.String()
		if node.IP.To4() == nil {
			if !config.EnableHideDstIP || ip != config.DestIP {
				fmt.Fprintf(color.Output, "%s",
					color.New(color.FgWhite, color.Bold).Sprintf("%-25s", ip),
				)
			} else {
				fmt.Fprintf(color.Output, "%s",
					color.New(color.FgWhite, color.Bold).Sprintf("%-25s", hideIPPart(ip)),
				)
			}
		} else {
			if !config.EnableHideDstIP || ip != config.DestIP {
				fmt.Fprintf(color.Output, "%s",
					color.New(color.FgWhite, color.Bold).Sprintf("%-15s", ip),
				)
			} else {
				fmt.Fprintf(color.Output, "%s",
					color.New(color.FgWhite, color.Bold).Sprintf("%-15s", hideIPPart(ip)),
				)
			}
		}

		asn := ipAsn(ip)
		if asn != "" {
			info := ASNInfo{ASN: asn}
			lineType := getLineType(info)
			if strings.Contains(lineType, "精品线路") {
				fmt.Fprintf(color.Output, " %s", color.New(color.FgHiYellow, color.Bold).Sprintf("AS%-6s", asn[2:]))
			} else if strings.Contains(lineType, "优质线路") {
				fmt.Fprintf(color.Output, " %s", color.New(color.FgHiGreen, color.Bold).Sprintf("AS%-6s", asn[2:]))
			} else {
				fmt.Fprintf(color.Output, " %s", color.New(color.FgWhite, color.Bold).Sprintf("AS%-6s", asn[2:]))
			}
		} else {
			fmt.Printf(" %-8s", "*")
		}

		// Print RTT times
		fmt.Printf("   ")
		for i, rtt := range node.RTT {
			if i > 0 {
				fmt.Printf(" / ")
			}
			fmt.Fprintf(color.Output, "%s",
				color.New(color.FgHiCyan, color.Bold).Sprintf("%.2f ms", rtt.Seconds()*1000),
			)
		}
		fmt.Println()
		blockDisplay = true
	}
}

func hideIPPart(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s.%s.*.*", parts[0], parts[1])
	}
	return ip
}
