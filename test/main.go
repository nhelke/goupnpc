package main

import (
	l4g "code.google.com/p/log4go"
	upnpc "github.com/nhelke/goupnpc"
	"time"
)

func main() {
	l4g.AddFilter("stdout", l4g.DEBUG, l4g.NewConsoleLogWriter())

	upnpc.GetConnectionStatus()

	l4g.Info("Closing")
	l4g.Close()

	time.Sleep(1 * time.Second)

}
