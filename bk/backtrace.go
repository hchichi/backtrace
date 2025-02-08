package backtrace

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/fatih/color"
)

var (
	EnableLoger      = false
	BackTraceVersion = "v1.0.0"
)

func BackTrace() {
	var wg sync.WaitGroup
	ch := make(chan Result, len(ips))

	for i := range ips {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			analyzeRoute(ch, i)
		}(i)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	results := make([]Result, len(ips))
	for r := range ch {
		results[r.i] = r
	}

	for _, r := range results {
		fmt.Println(r.s)
	}
}

func analyzeRoute(ch chan Result, i int) {
	hops, err := Trace(net.ParseIP(ips[i]))
	if err != nil {
		s := fmt.Sprintf("%-8s %-15s %v", names[i], ips[i], err)
		ch <- Result{i, s}
		return
	}

	var routes []string
	var lastASN string

	for _, h := range hops {
		for _, n := range h.Nodes {
			info, err := getASNInfo(n.IP.String())
			if err != nil || info == nil {
				continue
			}

			// 避免重复的ASN
			if info.ASN == lastASN {
				continue
			}
			lastASN = info.ASN

			lineType := getLineType(*info)
			if lineType != "" {
				routes = append(routes, lineType)
			}
		}
	}

	// 格式化输出
	s := fmt.Sprintf("%-8s %-15s", names[i], ips[i])
	if len(routes) > 0 {
		for i, route := range routes {
			if i > 0 {
				s += " "
			}
			switch {
			case strings.Contains(route, "精品线路"):
				s += color.New(color.FgHiYellow, color.Bold).Sprintf("%-20s", route)
			case strings.Contains(route, "优质线路"):
				s += color.New(color.FgHiGreen, color.Bold).Sprintf("%-20s", route)
			default:
				s += color.New(color.FgWhite, color.Bold).Sprintf("%-20s", route)
			}
		}
	} else {
		s += color.New(color.FgRed, color.Bold).Sprintf("检测不到已知线路的ASN")
	}

	ch <- Result{i, s}
}
