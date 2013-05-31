package goupnp

import (
	"net"
	"testing"
)

func TestIsPrivateIPAddress(t *testing.T) {
	privateAddrs := []net.IP{
		net.IPv4(192, 168, 2, 32),
		net.IPv4(10, 230, 46, 52),
		net.IPv4(172, 22, 8, 61),
	}

	for i := 0; i < len(privateAddrs); i++ {
		if !IsPrivateIPAddress(privateAddrs[i]) {
			t.Errorf("Incorrectly did not identify %v as a private IPv4 Address", privateAddrs[i])
		}
	}

	publicAddrs := []net.IP{
		net.IPv4(184, 85, 61, 15),
		net.IPv4(137, 164, 29, 67),
		net.IPv4(8, 8, 8, 8),
	}

	for i := 0; i < len(publicAddrs); i++ {
		if IsPrivateIPAddress(publicAddrs[i]) {
			t.Errorf("Incorrectly identified %v as a private IPv4 Address", publicAddrs[i])
		}
	}
}
