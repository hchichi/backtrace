package backtrace

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
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
		"AS23764": "电信CTGNet  [精品线路]", // CTGNet
		"AS4809a": "电信CN2GIA  [精品线路]", // CN2 GIA
		"AS4809b": "电信CN2GT   [优质线路]", // CN2 GT
		"AS4134":  "电信163     [普通线路]", // 电信163
		"AS10099": "联通CUG     [优质线路]", // 联通CUG
		"AS9929":  "联通9929    [优质线路]", // 联通9929
		"AS4837":  "联通4837    [普通线路]", // 联通4837
		"AS58807": "移动CMIN2   [精品线路]", // 移动CMIN2
		"AS9808":  "移动CMI     [优质线路]", // 移动CMI
		"AS58453": "移动CMI     [普通线路]", // 移动CMI
	}
)

func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{} // 用于存储已经遇到的元素
	result := []string{}             // 存储去重后的结果
	for v := range elements {        // 遍历切片中的元素
		if encountered[elements[v]] == true { // 如果该元素已经遇到过
			// 存在过就不加入了
		} else {
			encountered[elements[v]] = true      // 将该元素标记为已经遇到
			result = append(result, elements[v]) // 将该元素加入到结果切片中
		}
	}
	return result // 返回去重后的结果切片
}

func Gold(str string) string {
	return fmt.Sprintf("\033[38;5;220m%s\033[0m", str)
}

func LightGold(str string) string {
	return fmt.Sprintf("\033[38;5;178m%s\033[0m", str)
}

func Rose(str string) string {
	return fmt.Sprintf("\033[38;5;161m%s\033[0m", str)
}

func LightRose(str string) string {
	return fmt.Sprintf("\033[38;5;168m%s\033[0m", str)
}

func RoyalBlue(str string) string {
	return fmt.Sprintf("\033[38;5;69m%s\033[0m", str)
}

func LightBlue(str string) string {
	return fmt.Sprintf("\033[38;5;75m%s\033[0m", str)
}

func Gray(str string) string {
	return fmt.Sprintf("\033[38;5;245m%s\033[0m", str)
}

type ASNInfo struct {
	Number      string              // ASN号
	Name        string              // 运营商名称
	Route       string              // IP段
	Country     string              // 国家
	Description string              // 描述
	Whois       string              // Whois信息
	Router      map[string][]string // 路由表信息
}

type LineQuality int

const (
	Premium LineQuality = iota // 精品线路
	Quality                    // 优质线路
	Regular                    // 普通线路
	Unknown                    // 未知线路
)

func getLineQuality(asn string, whois string, ip string) LineQuality {
	// 通过多个维度判断线路质量
	switch {
	// 精品线路判断
	case asn == "23764" || // CTGNet
		asn == "58807" || // CMIN2
		(asn == "4809" && strings.HasPrefix(ip, "59.43.80.")) || // CN2 GIA
		strings.HasPrefix(ip, "69.194.") || // CTGNet IP特征
		strings.HasPrefix(ip, "203.22.") || // CTGNet IP特征
		strings.Contains(whois, "CTG-CN") || // CTG 特征
		strings.Contains(whois, "CTGNET"): // CTG 特征
		return Premium

	// 优质线路判断
	case asn == "10099" || // CUG
		asn == "9929" || // CUII
		(asn == "4809" && strings.HasPrefix(ip, "59.43.")) || // CN2 GT
		strings.HasPrefix(ip, "218.105.") || // 联通9929特征
		strings.HasPrefix(ip, "210.51.") || // 联通9929特征
		strings.HasPrefix(ip, "219.158.113.") || // CUG特征IP
		strings.Contains(whois, "CNC-BACKBONE") ||
		strings.Contains(whois, "CUG-BACKBONE"):
		return Quality

	// 普通线路判断
	case asn == "4134" || // 电信163
		asn == "4837" || // 联通4837
		asn == "9808" || // 移动CMI
		asn == "58453" || // 移动CMI
		strings.HasPrefix(ip, "202.97.") || // 163特征IP
		strings.HasPrefix(ip, "219.158.") || // 联通4837特征IP
		strings.HasPrefix(ip, "211.136.") || // 移动普通网络特征IP
		strings.HasPrefix(ip, "221.179."): // 移动普通网络特征IP
		return Regular

	default:
		return Unknown
	}
}

func getLineColor(asn string, quality LineQuality) func(string) string {
	switch {
	case strings.HasPrefix(asn, "4809"), // 电信
		strings.HasPrefix(asn, "4134"),
		strings.HasPrefix(asn, "23764"):
		switch quality {
		case Premium:
			return Gold
		case Quality:
			return LightGold
		default:
			return Gray
		}

	case strings.HasPrefix(asn, "9929"), // 联通
		strings.HasPrefix(asn, "4837"),
		strings.HasPrefix(asn, "10099"):
		switch quality {
		case Premium, Quality:
			return Rose
		default:
			return Gray
		}

	case strings.HasPrefix(asn, "58807"), // 移动
		strings.HasPrefix(asn, "9808"),
		strings.HasPrefix(asn, "58453"):
		switch quality {
		case Premium:
			return RoyalBlue
		case Quality:
			return LightBlue
		default:
			return Gray
		}

	default:
		return Gray
	}
}

type ASNCache struct {
	cache    map[string]ASNInfo
	cacheMux sync.RWMutex
	ttl      time.Duration
}

var (
	asnCache = &ASNCache{
		cache: make(map[string]ASNInfo),
		ttl:   24 * time.Hour, // 缓存24小时
	}
)

func isPrivateIP(ip string) bool {
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false
	}

	// 检查是否是私有 IP 地址
	privateIPBlocks := []string{
		"127.0.0.0/8",    // localhost
		"10.0.0.0/8",     // 私有网络
		"172.16.0.0/12",  // 私有网络
		"192.168.0.0/16", // 私有网络
		"169.254.0.0/16", // 链路本地
		"198.18.0.0/15",  // 网络基准测试
	}

	for _, block := range privateIPBlocks {
		_, ipnet, err := net.ParseCIDR(block)
		if err != nil {
			continue
		}
		if ipnet.Contains(ipAddr) {
			return true
		}
	}
	return false
}

func getASNInfo(ip string) (*ASNInfo, error) {
	// 检查是否是私有 IP
	if isPrivateIP(ip) {
		return nil, fmt.Errorf("private IP address")
	}

	// 先检查缓存
	asnCache.cacheMux.RLock()
	if info, ok := asnCache.cache[ip]; ok {
		asnCache.cacheMux.RUnlock()
		return &info, nil
	}
	asnCache.cacheMux.RUnlock()

	// 尝试使用 IP 特征判断
	if asn := ipAsn(ip); asn != "" {
		info := classifyASN(asn)
		// 更新缓存
		asnCache.cacheMux.Lock()
		asnCache.cache[ip] = *info
		asnCache.cacheMux.Unlock()
		return info, nil
	}

	// 尝试使用 RIPE API
	resp, err := http.Get(fmt.Sprintf("https://stat.ripe.net/data/network-info/data.json?resource=%s", ip))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			ASNs []struct {
				ASN interface{} `json:"asn"`
			} `json:"asns"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Data.ASNs) == 0 {
		return nil, fmt.Errorf("no ASN found for IP %s", ip)
	}

	// 处理 ASN 值，可能是字符串或数字
	var asnStr string
	switch v := result.Data.ASNs[0].ASN.(type) {
	case string:
		asnStr = v
	case float64:
		asnStr = fmt.Sprintf("%.0f", v)
	default:
		return nil, fmt.Errorf("unexpected ASN type for IP %s", ip)
	}

	// 移除可能的 "AS" 前缀
	asnStr = strings.TrimPrefix(asnStr, "AS")

	info := classifyASN(asnStr)

	// 更新缓存
	asnCache.cacheMux.Lock()
	asnCache.cache[ip] = *info
	asnCache.cacheMux.Unlock()

	return info, nil
}

func getQualityDesc(quality LineQuality) string {
	switch quality {
	case Premium:
		return "精品线路"
	case Quality:
		return "优质线路"
	case Regular:
		return "普通线路"
	default:
		return "未知线路"
	}
}

func trace(ch chan Result, i int) {
	// 首先尝试使用 NextTrace
	hops, err := useNextTrace(ips[i])
	if err != nil {
		// 如果 NextTrace 失败，使用本地实现
		hops, err = TraceRoute(net.ParseIP(ips[i]))
		if err != nil {
			s := fmt.Sprintf("%v %-15s %v", names[i], ips[i], err)
			ch <- Result{i, s}
			return
		}
	}

	var tempText string
	tempText += fmt.Sprintf("%v ", names[i])
	tempText += fmt.Sprintf("%-15s ", ips[i])

	seenASNs := make(map[string]bool)
	hasValidHop := false

	// 检查汇聚层
	hasConvergence, convergenceIndex := isConvergenceLayer(hops)
	if hasConvergence {
		hops = hops[:convergenceIndex+1] // 只保留汇聚层之前的跳点
	}

	for _, hopIP := range hops {
		if hopIP == nil {
			continue
		}

		hasValidHop = true
		asnInfo, err := getASNInfo(hopIP.String())
		if err != nil {
			fmt.Printf("Debug: Error getting ASN info for IP %s: %v\n", hopIP.String(), err)
			continue
		}

		if !seenASNs[asnInfo.Number] {
			seenASNs[asnInfo.Number] = true

			quality := getLineQuality(asnInfo.Number, asnInfo.Whois, hopIP.String())
			colorFunc := getLineColor(asnInfo.Number, quality)

			// 显示ASN信息
			tempText += colorFunc(fmt.Sprintf("%s ", asnInfo.Name))

			// 显示描述信息
			tempText += colorFunc(asnInfo.Description) + " "
		}
	}

	if !hasValidHop {
		tempText += Red("无法获取路由跳点")
	} else if len(seenASNs) == 0 {
		tempText += Red("检测不到已知线路的ASN")
	}

	ch <- Result{i, tempText}
}

func TraceRoute(ip net.IP) ([]net.IP, error) {
	const maxTTL = 30
	var hops []net.IP

	// 创建 ICMP 连接
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	for ttl := 1; ttl <= maxTTL; ttl++ {
		// 设置 TTL
		if err := c.IPv4PacketConn().SetTTL(ttl); err != nil {
			return nil, err
		}

		// 创建 ICMP 消息
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID:   os.Getpid() & 0xffff,
				Seq:  ttl,
				Data: []byte("HELLO-R-U-THERE"),
			},
		}

		wb, err := wm.Marshal(nil)
		if err != nil {
			return nil, err
		}

		// 发送 ICMP 消息
		if _, err := c.WriteTo(wb, &net.IPAddr{IP: ip}); err != nil {
			return nil, err
		}

		// 设置读取超时
		if err := c.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
			return nil, err
		}

		// 读取响应
		rb := make([]byte, 1500)
		n, peer, err := c.ReadFrom(rb)
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				continue
			}
			return nil, err
		}

		rm, err := icmp.ParseMessage(1, rb[:n])
		if err != nil {
			return nil, err
		}

		switch rm.Type {
		case ipv4.ICMPTypeTimeExceeded:
			// 添加中间路由器的 IP
			if peerIP, ok := peer.(*net.IPAddr); ok {
				hops = append(hops, peerIP.IP)
			}
		case ipv4.ICMPTypeEchoReply:
			// 到达目标 IP
			if peerIP, ok := peer.(*net.IPAddr); ok {
				hops = append(hops, peerIP.IP)
			}
			return hops, nil
		}
	}

	return hops, nil
}

func classifyASN(asn string) *ASNInfo {
	info := &ASNInfo{Number: asn}

	// 移除可能的 "AS" 前缀
	asn = strings.TrimPrefix(asn, "AS")

	switch asn {
	case "23764":
		info.Name = "AS23764"
		info.Description = "电信CTGNet [精品线路]"
	case "4809":
		if isGIA(asn) {
			info.Name = "AS4809"
			info.Description = "电信CN2GIA [精品线路]"
		} else {
			info.Name = "AS4809"
			info.Description = "电信CN2GT [优质线路]"
		}
	case "4134":
		info.Name = "AS4134"
		info.Description = "电信163 [普通线路]"
	case "10099":
		info.Name = "AS10099"
		info.Description = "联通CUG [优质线路]"
	case "9929":
		info.Name = "AS9929"
		info.Description = "联通9929 [优质线路]"
	case "4837":
		info.Name = "AS4837"
		info.Description = "联通4837 [普通线路]"
	case "58807":
		info.Name = "AS58807"
		info.Description = "移动CMIN2 [精品线路]"
	case "9808":
		info.Name = "AS9808"
		info.Description = "移动CMI [优质线路]"
	case "58453":
		info.Name = "AS58453"
		info.Description = "移动CMI [普通线路]"
	default:
		info.Name = fmt.Sprintf("AS%s", asn)
		info.Description = "未知线路"
	}

	return info
}

func isGIA(asn string) bool {
	// 实现具体判断逻辑
	// 1. 检查IP特征(59.43.80.*)
	// 2. 检查路由表中是否包含AS4134
	// 3. 检查是否有GIA特征的Whois信息
	// 4. 检查是否有独立的C-I段
	return false // 默认返回false，需要具体IP特征判断
}

func ipAsn(ip string) string {
	// 先检查是否是特殊线路
	switch {
	// CTGNet特征IP
	case strings.HasPrefix(ip, "69.194.") ||
		strings.HasPrefix(ip, "203.22.") ||
		strings.HasPrefix(ip, "103.135.224.") ||
		strings.HasPrefix(ip, "103.135.225.") ||
		strings.HasPrefix(ip, "103.135.226.") ||
		strings.HasPrefix(ip, "103.135.227."):
		return "23764" // CTGNet

	// CN2特征IP
	case strings.HasPrefix(ip, "59.43.80."):
		return "4809" // CN2 GIA
	case strings.HasPrefix(ip, "59.43.244.") ||
		strings.HasPrefix(ip, "59.43.245.") ||
		strings.HasPrefix(ip, "59.43.246.") ||
		strings.HasPrefix(ip, "59.43.247.") ||
		strings.HasPrefix(ip, "59.43.248.") ||
		strings.HasPrefix(ip, "59.43.249."):
		return "4809" // CN2 GT

	// 联通精品网特征IP
	case strings.HasPrefix(ip, "218.105.") ||
		strings.HasPrefix(ip, "210.51.") ||
		strings.HasPrefix(ip, "113.200.") ||
		strings.HasPrefix(ip, "112.64."):
		return "9929" // 联通9929
	case strings.HasPrefix(ip, "219.158.113.") ||
		strings.HasPrefix(ip, "219.158.114.") ||
		strings.HasPrefix(ip, "219.158.115."):
		return "10099" // 联通CUG

	// 移动精品网特征IP
	case strings.HasPrefix(ip, "223.120.19.") ||
		strings.HasPrefix(ip, "223.120.17.") ||
		strings.HasPrefix(ip, "223.120.16.") ||
		strings.HasPrefix(ip, "223.120.140.") ||
		strings.HasPrefix(ip, "223.120.130.") ||
		strings.HasPrefix(ip, "223.120.131.") ||
		strings.HasPrefix(ip, "223.120.141.") ||
		strings.HasPrefix(ip, "223.120.6.") ||
		strings.HasPrefix(ip, "223.120.7."):
		return "58807" // CMIN2
	}

	// 再检查普通线路
	ipParts := strings.Split(ip, ".")
	if len(ipParts) != 4 {
		return ""
	}

	// 提取前两段
	prefix := ipParts[0] + "." + ipParts[1]

	switch {
	// 电信163
	case strings.HasPrefix(prefix, "202.97") ||
		strings.HasPrefix(prefix, "202.96") ||
		strings.HasPrefix(prefix, "61.139") ||
		strings.HasPrefix(prefix, "58.60") ||
		strings.HasPrefix(prefix, "219.141") ||
		strings.HasPrefix(prefix, "219.140") ||
		strings.HasPrefix(prefix, "180.149") ||
		strings.HasPrefix(prefix, "180.153") ||
		strings.HasPrefix(prefix, "61.182") ||
		strings.HasPrefix(prefix, "61.183"):
		return "4134"

	// 联通4837
	case strings.HasPrefix(prefix, "219.158") ||
		strings.HasPrefix(prefix, "210.22") ||
		strings.HasPrefix(prefix, "210.21") ||
		strings.HasPrefix(prefix, "202.106") ||
		strings.HasPrefix(prefix, "119.6") ||
		strings.HasPrefix(prefix, "119.7") ||
		strings.HasPrefix(prefix, "125.33") ||
		strings.HasPrefix(prefix, "125.34"):
		return "4837"

	// 移动CMI
	case strings.HasPrefix(prefix, "223.118") ||
		strings.HasPrefix(prefix, "223.119") ||
		strings.HasPrefix(prefix, "223.120") ||
		strings.HasPrefix(prefix, "223.121") ||
		strings.HasPrefix(prefix, "223.122") ||
		strings.HasPrefix(prefix, "223.123"):
		return "58453"

	// 移动普通网络
	case strings.HasPrefix(prefix, "211.136") ||
		strings.HasPrefix(prefix, "221.179") ||
		strings.HasPrefix(prefix, "120.196") ||
		strings.HasPrefix(prefix, "211.137") ||
		strings.HasPrefix(prefix, "111.24") ||
		strings.HasPrefix(prefix, "111.25"):
		return "9808"
	}

	return ""
}

func isConvergenceLayer(hops []net.IP) (bool, int) {
	if len(hops) < 2 {
		return false, -1
	}

	// 检查连续的跳点是否属于同一运营商
	prevASN := ""
	convergenceCount := 0
	convergenceStart := -1

	for i, hop := range hops {
		if hop == nil {
			continue
		}

		asnInfo, err := getASNInfo(hop.String())
		if err != nil {
			continue
		}

		if prevASN == "" {
			prevASN = asnInfo.Number
			continue
		}

		// 如果当前 ASN 与前一个相同，可能是汇聚层
		if asnInfo.Number == prevASN {
			convergenceCount++
			if convergenceCount >= 2 && convergenceStart == -1 {
				convergenceStart = i - 2
			}
		} else {
			convergenceCount = 0
			prevASN = asnInfo.Number
		}
	}

	return convergenceStart != -1, convergenceStart
}

func useNextTrace(ip string) ([]net.IP, error) {
	// 检查 NextTrace 是否安装
	whichCmd := exec.Command("which", "nexttrace")
	if err := whichCmd.Run(); err != nil {
		// NextTrace 未安装，尝试使用官方安装脚本
		fmt.Println("NextTrace 未安装，正在尝试自动安装...")

		// 检查是否有 root/sudo 权限
		checkRoot := exec.Command("id", "-u")
		output, err := checkRoot.Output()
		if err != nil {
			return nil, fmt.Errorf("检查权限失败: %v", err)
		}

		isRoot := strings.TrimSpace(string(output)) == "0"
		var installCmd *exec.Cmd
		if isRoot {
			installCmd = exec.Command("bash", "-c", "curl -Ls https://raw.githubusercontent.com/sjlleo/nexttrace/main/nt_install.sh | bash")
		} else {
			installCmd = exec.Command("sudo", "bash", "-c", "curl -Ls https://raw.githubusercontent.com/sjlleo/nexttrace/main/nt_install.sh | bash")
		}

		if output, err := installCmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("安装 NextTrace 失败: %v\n%s", err, output)
		}

		// 验证安装是否成功
		if err := exec.Command("which", "nexttrace").Run(); err != nil {
			return nil, fmt.Errorf("NextTrace 安装失败，未找到可执行文件")
		}
		fmt.Println("NextTrace 安装成功！")
	}

	// 执行 NextTrace，使用更多参数以获得更好的结果
	var traceCmd *exec.Cmd
	checkRoot := exec.Command("id", "-u")
	output, err := checkRoot.Output()
	if err != nil {
		return nil, fmt.Errorf("检查权限失败: %v", err)
	}

	isRoot := strings.TrimSpace(string(output)) == "0"
	if isRoot {
		traceCmd = exec.Command("nexttrace", "-q", "1", "-n", "1", "-M", "icmp", "-T", "2", "-r", ip)
	} else {
		traceCmd = exec.Command("sudo", "nexttrace", "-q", "1", "-n", "1", "-M", "icmp", "-T", "2", "-r", ip)
	}

	output, err = traceCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("执行 NextTrace 失败: %v\n%s", err, output)
	}

	// 解析输出
	var hops []net.IP
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// 解析每一行，提取 IP 地址
		if ip := extractIP(line); ip != nil {
			hops = append(hops, ip)
		}
	}

	if len(hops) == 0 {
		return nil, fmt.Errorf("NextTrace 未返回任何有效的跳点信息")
	}

	return hops, nil
}

func extractIP(line string) net.IP {
	// 使用正则表达式匹配 IP 地址
	re := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	matches := re.FindAllString(line, -1)
	if len(matches) > 0 {
		// 返回最后一个匹配的 IP 地址，因为通常是目标 IP
		return net.ParseIP(matches[len(matches)-1])
	}
	return nil
}
