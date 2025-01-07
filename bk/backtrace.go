package backtrace

import (
	"fmt"
	"time"

	. "github.com/oneclickvirt/defaultset"
)

func BackTrace() {
	var (
		s [12]string // 对应 ips 目标地址数量
		c = make(chan Result)
		t = time.After(time.Second * 30) // 增加超时时间到 30 秒
	)

	// 启动所有 trace 协程
	activeTraces := 0
	for i := range ips {
		if ips[i] != "" {
			go trace(c, i)
			activeTraces++
		}
	}

	// 等待结果或超时
	completedTraces := 0
loop:
	for completedTraces < activeTraces {
		select {
		case o := <-c:
			s[o.i] = o.s
			completedTraces++
		case <-t:
			fmt.Println(Red("部分路由跟踪超时，显示已获取的结果"))
			break loop
		}
	}

	// 显示结果
	for i, r := range s {
		if r != "" {
			fmt.Println(r)
		} else if ips[i] != "" {
			fmt.Printf("%v %-15s %v\n", names[i], ips[i], Red("路由跟踪失败"))
		}
	}
}
