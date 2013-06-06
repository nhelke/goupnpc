package goupnpc

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
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
func discoverIGD(timeout time.Duration) (u string) {
	const (
		ssdpIPv4Addr = "239.255.255.250"
		ssdpPort     = 1900
		format       = "M-SEARCH * HTTP/1.1\r\n" +
			"HOST: %s:%d\r\n" +
			"ST: %s\r\n" +
			"MAN: \"ssdp:discover\"\r\n" +
			"MX: %d\r\n" +
			"\r\n"
	)

	var deviceTypes = []string{
		"urn:schemas-upnp-org:device:InternetGatewayDevice:1",
		"urn:schemas-upnp-org:service:WANIPConnection:1",
		"urn:schemas-upnp-org:service:WANPPPConnection:1",
		"upnp:rootdevice",
	}

	allIf, _ := net.ResolveUDPAddr("udp4", ":0")
	hostname := fmt.Sprintf("%s:%d", ssdpIPv4Addr, ssdpPort)
	broadcast, _ := net.ResolveUDPAddr("udp4", hostname)
	for i := 0; i < len(deviceTypes); i++ {
		conn, err := net.ListenUDP("udp4", allIf)
		if err == nil {
			// We want to timeout and move on to the next type after a couple of
			// seconds
			conn.SetDeadline(time.Now().Add(timeout))
			requestString := []byte(fmt.Sprintf(format, ssdpIPv4Addr, ssdpPort,
				deviceTypes[i], timeout/time.Second))
			// Send multicast request
			conn.WriteToUDP(requestString, broadcast)
			// Allocate a buffer for the response
			buf := make([]byte, 1500)
			// Get a response; the above timeout is still in effect as it
			// should be
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				l4g.Info(err)
			} else {
				// Parse and interpret the response and break if successful
				l4g.Info("Received %d bytes from %v", n, addr)
				req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(requestString)))
				if err != nil {
					l4g.Critical("Shit %v", err)
				}
				resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(buf[:n])), req)
				if err == nil {
					defer resp.Body.Close()
					urls := resp.Header["Location"]
					l4g.Info("%v\n%v", resp, urls)
					u = "*"
				} else {
					l4g.Critical("Shit %v", err)
				}
			}
		} else {
			l4g.Warn(err)
		}
	}
	// If we get here we could not find any UPnP devices

	return "http://172.28.165.1:1780/InternetGatewayDevice.xml"
}
