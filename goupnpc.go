package goupnpc

import (
	"encoding/xml"
	"io/ioutil"
	"net"
	"net/http"

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
}

func GetConnectionStatus() {
	url, _ := discoverIGD()
	resp, err := http.Get(url)
	if err == nil {
		l4g.Info("%v", resp.Header)
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			l4g.Info("%v", string(body))
			var x struct {
				XMLName     xml.Name `xml:"urn:schemas-upnp-org:device-1-0 root"`
				SpecVersion struct {
					Major int `xml:"major"`
				} `xml:"specVersion"`
			}
			err := xml.Unmarshal(body, &x)
			if err == nil {
				l4g.Info("%v", x)
				fxml, err := xml.MarshalIndent(x, "", "  ")
				if err == nil {
					l4g.Info("%v", string(fxml))
				}
			} else {
				l4g.Warn("Bad XML: %v", err)
			}
		} else {
			l4g.Warn("Error reading response")
		}
	}
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
