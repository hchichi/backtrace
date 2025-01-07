package backtrace

import (
	"fmt"
	"time"
)

func BackTrace() {
	var (
		s [12]string // 对应 ips 目标地址数量
		c = make(chan Result)
		t = time.After(time.Second * 120) // 增加超时时间到 120 秒，因为 NextTrace 可能需要更多时间
	)

	fmt.Println("正在进行路由检测...")

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
	hasTimeout := false
	failedTraces := 0
loop:
	for completedTraces < activeTraces {
		select {
		case o := <-c:
			if o.err != nil {
				failedTraces++
				s[o.i] = fmt.Sprintf("%v %-15s %v", names[o.i], ips[o.i], Red(o.err.Error()))
			} else {
				s[o.i] = o.s
			}
			completedTraces++
			// 显示进度和状态
			fmt.Printf("\r正在检测: %d/%d (成功: %d, 失败: %d)",
				completedTraces, activeTraces,
				completedTraces-failedTraces, failedTraces)
		case <-t:
			hasTimeout = true
			fmt.Println("\n部分路由跟踪超时，显示已获取的结果")
			break loop
		}
	}

	if !hasTimeout {
		if failedTraces > 0 {
			fmt.Printf("\n路由检测完成 (成功: %d, 失败: %d)\n",
				completedTraces-failedTraces, failedTraces)
		} else {
			fmt.Println("\n路由检测完成")
		}
	}

	// 显示结果
	fmt.Println("\n路由检测结果:")
	for i, r := range s {
		if r != "" {
			fmt.Println(r)
		} else if ips[i] != "" {
			fmt.Printf("%v %-15s %v\n", names[i], ips[i], Red("路由跟踪失败"))
		}
	}
}
