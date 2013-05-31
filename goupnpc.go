package goupnpc

import (
	"net"
	"time"

	// l4g "code.google.com/p/log4go"
)

const (
	TCP protocol = 1 << iota
	UDP
)

type PortMapping struct {
	InternalPort uint16
	ExternalPort uint16
	Protocol     protocol
}

func GetConnectionStatus() {
	discoverIGD(5 * time.Second)
}

func AddPortRedirection(lAddr net.Addr, externalPort uint16, protocol protocol) (ret chan error) {
	go addPortRedirection(ret)
	return
}

// This function attempts to create the passed mappings through a UPnP enabled
// gateway device on the LAN.
// The ExternalPort field is ignored and if the call is successful it is populated
// This means that YOU MUST NOT under any circumstances modify or read the passed
// mappings until you receive a signal back from this function
func AddPortRedirections(portMappings ...*PortMapping) {

}

func DeletePortRedirection(portMappings ...*PortMapping) {

}

func ListRedirections() {

}

func GetPresentationURL() {

}
