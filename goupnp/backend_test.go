package goupnp

import (
	"encoding/xml"
	"testing"
)

const exampleBelkinSOAP = `<?xml version="1.0"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
SOAP-ENV:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<SOAP-ENV:Body>
<m:GetStatusInfoResponse
xmlns:m="urn:schemas-upnp-org:service:WANIPConnection:1">
<NewConnectionStatus>Connected</NewConnectionStatus>
<NewLastConnectionError>ERROR_NONE</NewLastConnectionError>
<NewUptime>194979</NewUptime></m:GetStatusInfoResponse></SOAP-ENV:Body>
</SOAP-ENV:Envelope>
`

func TestSOAPParsing(t *testing.T) {
	var x soapEnvelope
	err := xml.Unmarshal([]byte(exampleBelkinSOAP), &x)
	if err != nil {
		t.Errorf("%v", err)
	} else if status := x.Body.Status.NewConnectionStatus; status != "Connected" {
		t.Errorf("Status incorrectly parsed as %v", status)
	}
}

const exampleBelkinPortMappingResponse = `<?xml version="1.0"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
SOAP-ENV:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<SOAP-ENV:Body><m:GetGenericPortMappingEntryResponse
xmlns:m="urn:schemas-upnp-org:service:WANIPConnection:1">
<NewRemoteHost></NewRemoteHost>
<NewExternalPort>5900</NewExternalPort>
<NewProtocol>TCP</NewProtocol>
<NewInternalPort>5901</NewInternalPort>
<NewInternalClient>192.168.2.5</NewInternalClient>
<NewEnabled>1</NewEnabled>
<NewPortMappingDescription>cPM.Port.Map.ee97f96de8c1647a</NewPortMappingDescription>
<NewLeaseDuration>0</NewLeaseDuration>
</m:GetGenericPortMappingEntryResponse>
</SOAP-ENV:Body>
</SOAP-ENV:Envelope>
`

func TestPortMappingResponseParsing(t *testing.T) {
	var x soapEnvelope
	err := xml.Unmarshal([]byte(exampleBelkinPortMappingResponse), &x)
	if err != nil {
		t.Errorf("%v", err)
	} else {
		referenceMapping := soapPortMapping{Protocol: "TCP", ExternalPort: 5900,
			InternalPort: 5901, InternalClient: "192.168.2.5", Enabled: 1,
			Description: "cPM.Port.Map.ee97f96de8c1647a"}
		if portMapping := x.Body.PortMapping; portMapping != referenceMapping {
			t.Errorf("Port mapping incorrectly parsed as %#v", portMapping)
		}
	}
}
