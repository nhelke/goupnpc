package main

import (
	"fmt"
	"os"
	"time"

	l4g "code.google.com/p/log4go"
	"github.com/nhelke/goupnpc/goupnp"
)

func main() {
	l4g.AddFilter("stdout", l4g.WARNING, l4g.NewConsoleLogWriter())

	if len(os.Args) < 2 {
		printUsage()
	} else {
		if os.Args[1] == "s" {
			igd := <-goupnp.DiscoverIGD()
			fmt.Println("Found IGD", igd)
			status := <-igd.GetConnectionStatus()
			fmt.Println(status, goupnp.IsPrivateIPAddress(status.IP))
			for portMapping := range igd.ListRedirections() {
				fmt.Println(portMapping)
			}
			myMapping := <-igd.AddLocalPortRedirection(6881, goupnp.TCP)
			fmt.Println(myMapping)
		}
	}

	time.Sleep(1 * time.Second)
}

func printUsage() {
	fmt.Println("That is not how you use me")
}
