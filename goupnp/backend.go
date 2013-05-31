package goupnp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"

	l4g "code.google.com/p/log4go"
)

type protocol int

func (self protocol) String() string {
	switch self {
	case TCP:
		return "TCP"
	case UDP:
		return "UDP"
	default:
		return "#(Bad Protocol Value)"
	}
}

type deviceElement struct {
	FriendlyName string `xml:"friendlyName"`
	Manufacturer string `xml:"manufacturer"`

	Services []struct {
		ServiceType string `xml:"serviceType"`
		ControlURL  string `xml:"controlURL"`
	} `xml:"serviceList>service"`

	Devices []deviceElement `xml:"deviceList>device",omitempty`
}

type deviceDescription struct {
	XMLName xml.Name `xml:"urn:schemas-upnp-org:device-1-0 root"`

	SpecVersion struct {
		Major int `xml:"major"`
		Minor int `xml:"minor"`
	} `xml:"specVersion"`

	URLBase string

	Device deviceElement `xml:"device"`
}

const (
	connectionTypeStringWANIP  = "urn:schemas-upnp-org:service:WANIPConnection:1"
	connectionTypeStringWANPPP = "urn:schemas-upnp-org:service:WANPPPConnection:1"
)

func statusRequestStringReader(upnptype string) io.Reader {
	return bytes.NewReader([]byte(fmt.Sprintf(statusRequestString, upnptype)))
}

const statusRequestString = `<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"
s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<s:Body>
<u:GetStatusInfo xmlns:u="%s">
</u:GetStatusInfo>
</s:Body>
</s:Envelope>
`

const externalIPRequestString = `<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"
s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<s:Body>
<u:GetExternalIPAddress xmlns:u="%s"></u:GetExternalIPAddress>
</s:Body>
</s:Envelope>
`

func externalIPRequestStringReader(upnptype string) io.Reader {
	return bytes.NewReader([]byte(fmt.Sprintf(externalIPRequestString, upnptype)))
}

const portMappingRequestString = `<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" ` +
	`s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"><s:Body>` +
	`<u:GetGenericPortMappingEntry xmlns:u="%s"><NewPortMappingIndex>%d` +
	`</NewPortMappingIndex></u:GetGenericPortMappingEntry></s:Body></s:Envelope>
`

func portMappingRequestStringReader(upnptype string, index uint) io.Reader {
	str := fmt.Sprintf(portMappingRequestString, upnptype, index)
	return bytes.NewReader([]byte(str))
}

const createPortMappingString = `<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" ` +
	`s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"><s:Body>` +
	`<u:AddPortMapping xmlns:u="%s"><NewRemoteHost></NewRemoteHost>` +
	`<NewExternalPort>%d</NewExternalPort><NewProtocol>%s</NewProtocol>` +
	`<NewInternalPort>%d</NewInternalPort>` +
	`<NewInternalClient>%s</NewInternalClient><NewEnabled>1</NewEnabled>` +
	`<NewPortMappingDescription>%s</NewPortMappingDescription>` +
	`<NewLeaseDuration>0</NewLeaseDuration>` +
	`</u:AddPortMapping></s:Body></s:Envelope>
`

func createPortMappingStringReader(upnptype string, port uint16,
	proto protocol, localAddr net.IP, description string) io.Reader {
	str := fmt.Sprintf(createPortMappingString, upnptype, port, proto, port,
		localAddr, description)
	return bytes.NewReader([]byte(str))
}

type soapEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`

	Body struct {
		IP struct {
			XMLName xml.Name `xml:"GetExternalIPAddressResponse"`

			NewExternalIPAddress string
		}
		Status struct {
			XMLName xml.Name `xml:"GetStatusInfoResponse"`

			NewConnectionStatus string
		}
		PortMapping soapPortMapping `xml:"GetGenericPortMappingEntryResponse"`
	}
}

type soapPortMapping struct {
	Protocol       string `xml:"NewProtocol"`
	ExternalPort   uint16 `xml:"NewExternalPort"`
	InternalPort   uint16 `xml:"NewInternalPort"`
	InternalClient string `xml:"NewInternalClient"`
	Enabled        int    `xml:"NewEnabled"`
	Description    string `xml:"NewPortMappingDescription"`
	Lease          uint   `xml:"NewLeaseDuration"`
}

func (self *IGD) soapRequest(requestType string,
	requestXML io.Reader) (x *soapEnvelope, ok bool) {
	req, err := http.NewRequest("POST", self.controlURL.String(), requestXML)
	if err != nil {
		panic("Programming Error: This hand crafted http.Request should not be bad")
	}
	req.Header.Add("Content-Type", "text/xml")
	req.Header.Add("SOAPAction",
		`"`+self.upnptype+"#"+requestType+`"`)
	req.Header.Add("Connection", "Close")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Pragma", "no-cache")

	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		// We got something back, lets not leak it
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			l4g.Debug("SOAP Response:\n%s", string(body))
			err := xml.Unmarshal(body, &x)
			if err == nil {
				ok = true
			} else {
				l4g.Warn(err)
			}
		} else {
			l4g.Warn(err)
		}
	} else {
		l4g.Warn(err)
	}

	return
}
