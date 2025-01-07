package main

import (
	"flag"
	"fmt"

	backtrace "github.com/ilychi/backtrace/bk"
)

var (
	version   = backtrace.BackTraceVersion
	buildTime = "unknown"
)

func main() {
	var (
		showVersion bool
		showHelp    bool
		enableLog   bool
		showIP      bool
	)

	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&showHelp, "h", false, "Show help information")
	flag.BoolVar(&enableLog, "e", false, "Enable logging")
	flag.BoolVar(&showIP, "s", true, "Show IP information (default: true)")
	flag.Parse()

	if showHelp {
		fmt.Printf("Usage: backtrace [options]\n")
		flag.PrintDefaults()
		return
	}

	if showVersion {
		fmt.Printf("Version: %s\nBuild Time: %s\n", version, buildTime)
		return
	}

	if enableLog {
		backtrace.InitLogger()
	}

	backtrace.BackTrace()
}
