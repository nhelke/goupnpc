package goupnp

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	l4g "code.google.com/p/log4go"
)

// Returns all local interface IP addresses in the private network range
// They are traversed in the order returned by `net.InterfaceAddrs()`
func localPrivateAddrs() (ret []*net.UDPAddr) {
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for i := 0; i < len(addrs); i++ {
			var ip net.IP

			if addr, ok := addrs[i].(*net.IPNet); ok {
				ip = addr.IP
			} else if addr, ok := addrs[i].(*net.IPAddr); ok {
				ip = addr.IP
			}

			if ip != nil {
				if IsPrivateIPAddress(ip) {
					l4g.Debug("Found private addr %v", ip)
					ret = append(ret, &net.UDPAddr{
						IP:   ip,
						Port: 0,
					})
				}
			}
		}
	} else {
		l4g.Warn(err)
	}
	return
}

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

// These are the various device types we need to M-SEARCH the local subnet
// for. The last one is a fallback copied from MiniUPnPC's behavior and is
// unlikely to yield usable results
//
// This slice is sorted from most specific device type to the most general.
// Be advised that the below loop relies on this ordering.
var deviceTypes = []string{
	"urn:schemas-upnp-org:device:InternetGatewayDevice:1",
	"urn:schemas-upnp-org:service:WANIPConnection:1",
	"urn:schemas-upnp-org:service:WANPPPConnection:1",
	"upnp:rootdevice",
}

// This function implements the strict minimum of SSDP in order to discover the
// an IGD on the passed localBindAddr. The function blocks until a UPnP enabled
// IGD is found or timeout of four seconds expires. Timeouts smaller than 3
// seconds are unreasonable This function's behavior is not defined if the
// passed localBindAddr is not an IP address in the private network range. You
// may wish to use goupnp.localPrivateAddrs() to obtain a list of valid such
// addresses for the localhost.
func discoverIGDDescriptionURL(localBindAddr *net.UDPAddr) (u *url.URL, ok bool) {
	multicastAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d",
		ssdpIPv4Addr, ssdpPort))
	if err != nil {
		panic("Programming error: Our UDPAddr is incorrect")
	}

	conn, err := net.ListenUDP("udp4", localBindAddr)
	var timeout time.Duration = 4 * time.Second
	if err == nil {
		// For each device type, M-SEARCH for it, return the first one found
		// As deviceTypes is sorted from most specific to least specific type
		// returning the first should work fine.
		for i := 0; i < len(deviceTypes); i++ {
			// We write our own request *à la main* as trying to use Go's
			// standard library's HTTP package turns out to be require more
			// code than writing the request by hand, because of the non-
			// standard URL
			requestString := []byte(fmt.Sprintf(format, ssdpIPv4Addr, ssdpPort,
				deviceTypes[i], timeout/time.Second))
			// Allocate a buffer for the response
			buf := make([]byte, 1500)
			// We want to timeout and move on to the next type after a couple of
			// seconds
			conn.SetDeadline(time.Now().Add(timeout))
			// Send multicast request
			conn.WriteToUDP(requestString, multicastAddr)
			// Get a response; the above timeout is still in effect as it
			// should be
			n, addr, err := conn.ReadFromUDP(buf)
			if err == nil {
				// Ugly ugly ugly workaround for URL panic on *
				adulteredReqStr := requestString
				adulteredReqStr[9] = '/'
				// Parse and interpret the response and break if successful
				l4g.Debug("Received %d bytes from %v", n, addr)
				req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(
					adulteredReqStr)))
				if err != nil {
					// Failure to parse the request represents an assertion
					// failure as we crafted the request ourselves and have
					// ensured its validity
					panic(err)
				}
				resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(
					buf[:n])), req)
				if err == nil {
					// We got something back, lets not leak it
					defer resp.Body.Close()
					l4g.Debug("Discovered device returned:\n%v", resp.Header)
					// We extract the description URL returned in the Location
					// header. The UPnP standard ensure
					urls := resp.Header["Location"]
					// We must check that the Location header exists as required
					// by the standard to avoid panicking if we get a bad
					// response missing a Location header.
					if len(urls) > 0 {
						// We have the location, bundle it up into a url.URL
						// object and return it
						u, err = url.Parse(urls[0])
						ok = err == nil
						return
					} else {
						l4g.Warn("Response did not contain Location header:\n%v",
							resp.Header)
					}
				} else {
					l4g.Warn(err)
				}
			} else {
				l4g.Warn(err)
			}
		}
	} else {
		l4g.Warn(err)
	}
	// If we get here we could not find any UPnP devices
	return // ok is false by default, signaling this failure
}

func extractConnectionControlURL(d deviceElement) (upnptype, url string, ok bool) {
	for i := 0; i < len(d.Services); i++ {
		if serviceType := d.Services[i].ServiceType; serviceType == connectionTypeStringWANIP ||
			serviceType == connectionTypeStringWANPPP {
			return serviceType, d.Services[i].ControlURL, true
		}
	}
	for i := 0; i < len(d.Devices); i++ {
		if upnptype, url, ok = extractConnectionControlURL(d.Devices[i]); ok {
			return
		}
	}
	return
}

func getConnectionControlURL(body []byte) (upnptype, url string, err error) {
	var x deviceDescription
	err = xml.Unmarshal(body, &x)
	if err == nil {
		var ok bool
		upnptype, url, ok = extractConnectionControlURL(x.Device)
		if !ok {
			err = errors.New("Control URL not found")
		} else {
			// The URLs in the DeviceDescription elements are relative
			url = x.URLBase + url
		}
	}
	return
}
