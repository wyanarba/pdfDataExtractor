package basicLogger

import "fmt"

type BasicLogger struct {
}

func (BasicLogger) Debug(args ...interface{}) {
	fmt.Print("[DBG] ")
	fmt.Println(args...)
}

func (BasicLogger) Info(args ...interface{}) {
	fmt.Print("[INF] ")
	fmt.Println(args...)
}

func (BasicLogger) Warn(args ...interface{}) {
	fmt.Print("[WRN] ")
	fmt.Println(args...)
}

func (BasicLogger) Error(args ...interface{}) {
	fmt.Print("[ERR] ")
	fmt.Println(args...)
}

func (BasicLogger) Panic(args ...interface{}) {
	fmt.Print("[PNC] ")
	fmt.Println(args...)
}
