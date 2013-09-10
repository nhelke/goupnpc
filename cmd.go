// This command is useful both to test the associated goupnp library and
// its source serves as an example of how to use said library.
//
// Usage instructions can be obtained by running it without any arguments.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	l4g "code.google.com/p/log4go"
	"github.com/nhelke/goupnpc/goupnp"
)

func main() {
	l4g.AddFilter("stdout", l4g.WARNING, l4g.NewConsoleLogWriter())

	if len(os.Args) < 2 {
		printUsage()
	} else {
		discover := goupnp.DiscoverIGD()
		if os.Args[1] == "s" {
			igd := <-discover
			status := <-igd.GetConnectionStatus()
			b, err := json.MarshalIndent(status, "", "  ")
			if err != nil {
				fmt.Println("error:", err)
			}
			os.Stdout.Write(b)
			fmt.Println()
		} else if os.Args[1] == "l" {
			igd := <-discover
			var portmappings []*goupnp.PortMapping
			for portMapping := range igd.ListRedirections() {
				portmappings = append(portmappings, portMapping)
			}
			b, err := json.MarshalIndent(portmappings, "", "  ")
			if err != nil {
				fmt.Println("error:", err)
			}
			os.Stdout.Write(b)
			fmt.Println()
		} else if os.Args[1] == "a" {
			igd := <-discover
			port, _ := strconv.Atoi(os.Args[2])
			proto := goupnp.ParseProtocol(os.Args[3])
			myMapping := <-igd.AddLocalPortRedirection(uint16(port), proto)
			fmt.Printf("%+v\n", myMapping)
		} else {
			printUsage()
		}
	}

	time.Sleep(1 * time.Second)
}

func printUsage() {
	fmt.Println(
		`Usage: goupnpc s
           Print IGD Status
       goupnpc a port protocol
           Add local port mapping with internal and external ports equal to
           port and protocol equal to, well I will let you guess
       goupnpc l
           Lists all port mappings on the IGD
NOTA BENE No error checking is performed, if anything goes wrong, it will
probably panic on nil or something`)
}
