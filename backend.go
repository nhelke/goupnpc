package goupnpc

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
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
	broadcast, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", ssdpIPv4Addr,
		ssdpPort))
	for i := 0; i < len(deviceTypes); i++ {
		conn, err := net.ListenUDP("udp4", allIf)
		if err == nil {
			// We want to timeout and move on to the next type after a couple of
			// seconds
			conn.SetDeadline(time.Now().Add(timeout))
			// Send multicast request
			conn.WriteToUDP([]byte(fmt.Sprintf(format, ssdpIPv4Addr, ssdpPort,
				deviceTypes[i], timeout/time.Second)), broadcast)
			// Allocate a buffer for the response
			buf := make([]byte, 1500)
			// Get a response; the above timeout is still in effect as it
			// should be
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				l4g.Info(err)
			} else {
				// Parse and interpret the response and break if successful
				l4g.Info("Received from %v:", addr)
				r := bytes.Split(buf[:n], []byte("\r\n"))
				if len(r) > 0 {
					headers := make(map[string]string)
					for i := 1; i < len(r); i++ {
						line := string(r[i])
						l4g.Info("%s", line)
						kv := strings.SplitN(line, ":", 2)
						if len(kv) == 2 {
							headers[strings.TrimSpace(strings.ToUpper(kv[0]))] =
								strings.TrimSpace(kv[1])
						}
					}
					if location, ok := headers["LOCATION"]; ok {
						u, err = url.Parse(location)
						if err == nil {
							return
						} else {
							l4g.Warn("Response missing LOCATION\n")
						}
					}
				} else {
					l4g.Warn("Malformed response, skipping device type")
					l4g.Debug("%s", string(buf[:n]))
				}
				resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(buf[:n])), nil)
				if err == nil {
					l4g.Info("%v", resp)
				}
			}
		} else {
			l4g.Warn(err)
		}
	}
	// If we get here we could not find any UPnP devices

	return
}
