package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/fatih/color"
	backtrace "github.com/oneclickvirt/backtrace/bk"
)

type IpInfo struct {
	Ip      string `json:"ip"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
	Org     string `json:"org"`
}

func checkNexttraceInstalled() bool {
	_, err := exec.LookPath("nexttrace")
	return err == nil
}

func installNexttrace() error {
	fmt.Println(color.YellowString("正在安装 nexttrace..."))
	cmd := exec.Command("bash", "-c", "curl -L https://github.com/nxtrace/Ntrace-V1/raw/main/nt_install.sh | bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func updateNexttrace() error {
	fmt.Println(color.YellowString("正在更新 nexttrace..."))
	cmd := exec.Command("nexttrace", "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	var showVersion, showIpInfo, help bool
	backtraceFlag := flag.NewFlagSet("backtrace", flag.ContinueOnError)
	backtraceFlag.BoolVar(&help, "h", false, "显示帮助信息")
	backtraceFlag.BoolVar(&showVersion, "v", false, "显示版本信息")
	backtraceFlag.BoolVar(&showIpInfo, "s", true, "禁用显示IP信息")
	backtraceFlag.BoolVar(&backtrace.EnableLoger, "e", false, "启用日志")
	backtraceFlag.Parse(os.Args[1:])

	if help {
		fmt.Printf("使用方法: %s [选项]\n", os.Args[0])
		backtraceFlag.PrintDefaults()
		return
	}

	if showVersion {
		fmt.Println(backtrace.BackTraceVersion)
		return
	}

	fmt.Println("项目地址: https://github.com/ilychi/backtrace")

	if showIpInfo {
		rsp, err := http.Get("http://ipinfo.io")
		if err != nil {
			fmt.Printf(color.RedString("获取IP信息失败: %v\n"), err)
		} else {
			info := IpInfo{}
			err = json.NewDecoder(rsp.Body).Decode(&info)
			if err != nil {
				fmt.Printf(color.RedString("解析IP信息失败: %v\n"), err)
			} else {
				fmt.Printf("%s%s %s%s %s%s\n",
					color.GreenString("国家: "), color.WhiteString(info.Country),
					color.GreenString("城市: "), color.WhiteString(info.City),
					color.GreenString("服务商: "), color.BlueString(info.Org))
			}
		}
	}

	// 检查并安装/更新nexttrace
	if !checkNexttraceInstalled() {
		if err := installNexttrace(); err != nil {
			fmt.Println(color.RedString("nexttrace 安装失败，请手动安装"))
			return
		}
	} else {
		if err := updateNexttrace(); err != nil {
			fmt.Println(color.YellowString("nexttrace 更新失败，将使用现有版本"))
		}
	}

	// 执行路由检测
	backtrace.BackTrace()

	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		fmt.Println("按回车键退出...")
		fmt.Scanln()
	}
}
