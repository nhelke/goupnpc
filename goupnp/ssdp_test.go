package goupnp

import (
	"testing"
)

func TestDescriptionParsing(t *testing.T) {
	const belkinDescription string = `<?xml version="1.0"?>
<root xmlns="urn:schemas-upnp-org:device-1-0">
	<specVersion>
		<major>1</major>
		<minor>0</minor>
	</specVersion>
	<URLBase>http://192.168.2.1:80</URLBase>
	<device>
		<deviceType>urn:schemas-upnp-org:device:InternetGatewayDevice:1</deviceType>
		<friendlyName>Belkin N150 Wireless Router</friendlyName>
		<manufacturer>Belkin International</manufacturer>
		<manufacturerURL>http://www.Belkin.com</manufacturerURL>
		<modelDescription>Wireless Router with Ethernet Switch</modelDescription>
		<modelName>N150 Wireless Router</modelName>
		<modelNumber>F9K1001</modelNumber>
		<modelURL>http://www.Belkin.com</modelURL>
		<serialNumber>201223GB303099</serialNumber>
		<UDN>uuid:upnp-InternetGatewayDevice-1_0-08863bf24378</UDN>
		<UPC>00000-00001</UPC>
		<serviceList>
			<service>
				<serviceType>urn:schemas-upnp-org:service:Layer3Forwarding:1</serviceType>
				<serviceId>urn:upnp-org:serviceId:L3Forwarding1</serviceId>
				<controlURL>/upnp/service/Layer3Forwarding</controlURL>
				<eventSubURL>/upnp/service/Layer3Forwarding</eventSubURL>
				<SCPDURL>/upnp/service/L3Frwd.xml</SCPDURL>
			</service>
		</serviceList>
		<deviceList>
			<device>
				<deviceType>urn:schemas-upnp-org:device:WANDevice:1</deviceType>
				<friendlyName>Belkin N150 Wireless Router</friendlyName>
				<manufacturer>Belkin International</manufacturer>
				<manufacturerURL>http://www.Belkin.com</manufacturerURL>
				<modelDescription>Wireless Router with Ethernet Switch</modelDescription>
				<modelName>N150 Wireless Router</modelName>
				<modelNumber>F9K1001</modelNumber>
				<modelURL>http://www.Belkin.com</modelURL>
				<serialNumber>201223GB303099</serialNumber>
				<UDN>uuid:upnp-WANDevice-1_0-08863bf24378</UDN>
				<UPC>00000-00001</UPC>
				<serviceList>
					<service>
						<serviceType>urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1</serviceType>
						<serviceId>urn:upnp-org:serviceId:WANCommonInterfaceConfig</serviceId>
						<controlURL>/upnp/service/WANCommonInterfaceConfig</controlURL>
						<eventSubURL>/upnp/service/WANCommonInterfaceConfig</eventSubURL>
						<SCPDURL>/upnp/service/WANCICfg.xml</SCPDURL>
					</service>
				</serviceList>
				<deviceList>
					<device>
						<deviceType>urn:schemas-upnp-org:device:WANConnectionDevice:1</deviceType>
						<friendlyName>Belkin N150 Wireless Router</friendlyName>
						<manufacturer>Belkin International</manufacturer>
						<manufacturerURL>http://www.Belkin.com</manufacturerURL>
						<modelDescription>Wireless Router with Ethernet Switch</modelDescription>
						<modelName>N150 Wireless Router</modelName>
						<modelNumber>F9K1001</modelNumber>
						<modelURL>http://www.Belkin.com</modelURL>
						<serialNumber>201223GB303099</serialNumber>
						<UDN>uuid:upnp-WANConnectionDevice-1_0-08863bf24378</UDN>
						<UPC>00000-00001</UPC>
						<serviceList>
							<service>
								<serviceType>urn:schemas-upnp-org:service:WANIPConnection:1</serviceType>
								<serviceId>urn:upnp-org:serviceId:WANIPConnection</serviceId>
								<controlURL>/upnp/service/WANIPConnection</controlURL>
								<eventSubURL>/upnp/service/WANIPConnection</eventSubURL>
								<SCPDURL>/upnp/service/WANIPCn.xml</SCPDURL>
							</service>
						</serviceList>
					</device>
				</deviceList>
			</device>
		</deviceList>
		<presentationURL>/index.html</presentationURL>
	</device>
</root>
`

	upnptype, url, err := getConnectionControlURL([]byte(belkinDescription))
	if upnptype != connectionTypeStringWANIP || url != "http://192.168.2.1:80/upnp/service/WANIPConnection" || err != nil {
		t.Errorf("Type: %v, URL: %v, Error: %v", upnptype, url, err)
	}
}
