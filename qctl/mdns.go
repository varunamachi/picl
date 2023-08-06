package srv

import (
	"github.com/hashicorp/mdns"
	"github.com/varunamachi/libx/errx"
)

type MdnsConfig struct {
	Instance    string `json:"instance"`
	Port        int    `json:"port"`
	ServiceName string `json:"serviceName"`
	ServiceDesc string `json:"serviceDesc"`
}

func StartMdnsService(config *MdnsConfig) (*mdns.Server, error) {

	info := []string{"Qctl service"}
	if config.ServiceName == "" {
		config.ServiceName = "_qctl._tcp.local"
	}
	service, err := mdns.NewMDNSService(
		config.Instance, config.ServiceName, "", "", config.Port, nil, info)
	if err != nil {
		errx.Errf(err, "failed to create mdns service")
	}

	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {

	}

	return server, nil
}
