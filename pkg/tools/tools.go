package tools

import (
	"bytes"
	"net"
	"strings"
)

//GetNetworkInfo return current ip and mac address
func GetNetworkInfo() (addr, ip string) {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			if iface.Flags&net.FlagUp != 0 && bytes.Compare(iface.HardwareAddr, nil) != 0 {
				addr = iface.HardwareAddr.String()
				ips, err := iface.Addrs()
				if err != nil {
					return
				}
				if len(ips) > 0 {
					ipValue := ips[0].String()
					ip = strings.Split(ipValue, "/")[0]
				}
				break
			}
		}
	}
	return
}
