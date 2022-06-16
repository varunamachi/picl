package qctl

import "net"

type controller struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	AddrIP4   net.IP `json:"addrIP4"`
	Port      int    `json:"port"`
}

func discover() ([]*controller, error) {
	return nil, nil
}
