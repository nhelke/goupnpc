package goupnp

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	l4g "code.google.com/p/log4go"
)

const (
	TCP protocol = 1 << iota
	UDP
)

type PortMapping struct {
	InternalPort uint16
	ExternalPort uint16
	Protocol     protocol
	InternalHost net.IP
	Description  string
	Enabled      bool
	Lease        uint
}

func (self *PortMapping) String() string {
	return fmt.Sprint(self.InternalHost, ":", self.InternalPort, "<=",
		self.ExternalPort, self.Protocol, ` "`, self.Description, `" (`,
		self.Enabled, ", ", self.Lease, ")")
}

type IGD struct {
	controlURL *url.URL
	upnptype   string
	iface      net.IP
}

func (self *IGD) String() string {
	return self.controlURL.String()
}

// This function returns a channel which will be sent the first IGD it finds in
// traversing `net.InterfaceAddrs()` with IP addresses in the private network
// range.
//
// The channel this function returns should be listened on to avoid leaking
// goroutines. Additionally the listener must check whether the channel the
// value returned by the channel against nil, to ensure that an IGD was indeed
// found.
func DiscoverIGD() (ret chan *IGD) {
	// Create the channel we will return
	ret = make(chan *IGD)

	// Do the work asynchronously
	go func() {
		// For each and every local address in the private network range
		bindLocalAddrs := localPrivateAddrs()
		l4g.Debug("Found %d private network interfaces", len(bindLocalAddrs))
		for i := 0; i < len(bindLocalAddrs); i++ {
			// Use SSDP to search for a UPnP-enabled IGD
			descURL, ok := discoverIGDDescriptionURL(bindLocalAddrs[i])

			if ok {
				// If we found one, we go fetch its description XML
				resp, err := http.Get(descURL.String())
				if err == nil {
					// We got something back, lets not leak it
					defer resp.Body.Close()
					// We read in the whole description into memory We might
					// envisage at a later date putting an upperbound on the
					// buffer, however there is no risk of buffer overflow, so
					// it is a low priority
					body, err := ioutil.ReadAll(resp.Body)
					if err == nil {
						l4g.Debug("Description XML:\n%s", string(body))
						// Parse the XML and extract relevant information
						upnptype, controlURL, err := getConnectionControlURL(body)
						if err == nil {
							var igd IGD
							// It worked, lets now try and wrap it in an igd struct
							igd.controlURL, err = url.Parse(controlURL)
							if err != nil {
								l4g.Warn("Failed to parse URL %v", controlURL)
							} else {
								// The URL was good, lets track the type as
								// well, in order to make the correct calls down
								// the line
								igd.upnptype = upnptype
								// We now add the local binding address to
								// enable the simple AddLocalPortRedirection
								// method
								igd.iface = bindLocalAddrs[i].IP

								ret <- &igd
								return
							}
						} else {
							l4g.Warn("Bad XML: %v", err)
						}
					} else {
						l4g.Warn("Error reading response")
					}
				}
			}
		}

		// If we get here we did not find an IGD, so we close the channel This
		// will have the effect of returning nil and will indicate the closure
		// to listeners.
		close(ret)
	}()
	return
}

type ConnectionStatus struct {
	Connected bool
	IP        net.IP
}

func (self *IGD) GetConnectionStatus() (ret chan *ConnectionStatus) {
	ret = make(chan *ConnectionStatus)

	go func() {
		x, ok := self.soapRequest("GetStatusInfo", statusRequestStringReader(self.upnptype))
		if ok && strings.EqualFold(x.Body.Status.NewConnectionStatus, "Connected") {
			y, ok := self.soapRequest("GetExternalIPAddress", externalIPRequestStringReader(self.upnptype))

			if ok {
				ipString := y.Body.IP.NewExternalIPAddress
				ip := net.ParseIP(ipString)
				if ip != nil {
					ret <- &ConnectionStatus{true, ip}
					return
				} else {
					l4g.Warn("Failed to parse IP string %v", ipString)
				}
			} else {
				l4g.Warn("Failed to get IP address after estabilishing the connection was ok")
			}
		} else if ok && strings.EqualFold(x.Body.Status.NewConnectionStatus, "Disconnected") {
			ret <- &ConnectionStatus{false, nil}
			return
		}
		close(ret)
	}()

	return
}

func (self *IGD) AddLocalPortRedirection(port uint16, proto protocol) (ret chan *PortMapping) {
	ret = make(chan *PortMapping)

	go func() {
		description := fmt.Sprintf("goupnp %s %d %s", self.iface, port, proto)
		_, ok := self.soapRequest("AddPortMapping",
			createPortMappingStringReader(self.upnptype, port,
				proto, self.iface, description))
		if ok {
			portMapping := PortMapping{
				InternalPort: port,
				ExternalPort: port,
				Enabled:      true,
				Description:  description,
				InternalHost: self.iface,
				Protocol:     proto,
			}

			ret <- &portMapping
		} else {
			close(ret)
		}
	}()

	return
}

func (self *IGD) DeletePortRedirection(portMappings ...*PortMapping) {

}

func (self *IGD) ListRedirections() (ret chan *PortMapping) {
	ret = make(chan *PortMapping, 10)

	go func() {
		var (
			ok bool = true
			i  uint = 0
			x  *soapEnvelope
		)
		for ; ; i++ {
			x, ok = self.soapRequest("GetGenericPortMappingEntry",
				portMappingRequestStringReader(self.upnptype, i))
			if ok {
				portMapping := PortMapping{
					InternalPort: x.Body.PortMapping.InternalPort,
					ExternalPort: x.Body.PortMapping.ExternalPort,
					Enabled:      x.Body.PortMapping.Enabled != 0,
					Description:  x.Body.PortMapping.Description,
					InternalHost: net.ParseIP(x.Body.PortMapping.InternalClient),
				}
				switch x.Body.PortMapping.Protocol {
				case "TCP", "tcp":
					portMapping.Protocol = TCP
				case "UDP", "udp":
					portMapping.Protocol = UDP
				}
				ret <- &portMapping
			} else {
				close(ret)
				break
			}
		}
	}()

	return
}

// This function returns true if and only if the passed IP address belongs to
// one of the ranges reserved in RFC 1918 for use in private networks
//
// This function is only part of this package as it is used internally and is
// public as it is deemed useful for developers to assertain whether or not a
// given external IP address such as one returned by GetConnectionStatus is
// public or not and as creating a standalone package just for this one function
// seemed excessive.
func IsPrivateIPAddress(addr net.IP) bool {
	ip4 := addr.To4()
	if ip4 == nil || !ip4.IsGlobalUnicast() {
		return false
	}
	var (
		aAddr = net.IPv4(10, 0, 0, 0)
		aMask = net.IPv4Mask(255, 0, 0, 0)

		bAddr = net.IPv4(172, 16, 0, 0)
		bMask = net.IPv4Mask(255, 240, 0, 0)

		cAddr = net.IPv4(192, 168, 0, 0)
		cMask = net.IPv4Mask(255, 255, 0, 0)
	)

	return ip4.Mask(aMask).Equal(aAddr) ||
		ip4.Mask(bMask).Equal(bAddr) ||
		ip4.Mask(cMask).Equal(cAddr)
}
