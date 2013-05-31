package goupnpc

import (
	"fmt"
	"net"
	"net/url"
	"time"

	l4g "code.google.com/p/log4go"
)

type protocol int

func addPortRedirection(done chan error) {

}

// This function implements the strict minimum of SSDP in order to discover the
// IGDs on the various net.Interfaces() with IP addresses in the private network
// range (indicating the probable existence of a NAT IGD)
// The function blocks until a UPnP enabled IGD is found or timeout
// rounded down to the nearest second expires. Timeouts smaller than 3 seconds
// are unreasonable
func discoverIGD(timeout time.Duration) (u *url.URL) {
	const (
		ssdpIPv4Addr        = "239.255.255.250"
		ssdpPort            = 1900
		format       string = "M-SEARCH * HTTP/1.1\r\n" +
			"HOST: %s:%d\r\n" +
			"ST: %s\r\n" +
			"MAN: \"ssdp:discover\"\r\n" +
			"MX: %d\r\n" +
			"\r\n"
	)

	var deviceList = []string{
		"urn:schemas-upnp-org:device:InternetGatewayDevice:1",
		"urn:schemas-upnp-org:service:WANIPConnection:1",
		"urn:schemas-upnp-org:service:WANPPPConnection:1",
		"upnp:rootdevice",
	}

	allIf, _ := net.ResolveUDPAddr("udp4", ":0")
	broadcast, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", ssdpIPv4Addr,
		ssdpPort))
	conn, err := net.ListenUDP("udp4", allIf)
	if err == nil {
		conn.SetDeadline(time.Now().Add(10 * time.Second))
		for i := 0; i < len(deviceList); i++ {
			conn.WriteToUDP([]byte(fmt.Sprintf(format, ssdpIPv4Addr, ssdpPort,
				deviceList[i], timeout/time.Second)), broadcast)
		}
		buf := make([]byte, 1500)
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				l4g.Info(err)
				break
			}
			l4g.Info("Received from %v:\n%s", addr, string(buf[:n]))
		}
	} else {
		l4g.Warn(err)
	}

	return
}
