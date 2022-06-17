package qctl

import (
	"net"
	"strings"
	"time"

	"github.com/hashicorp/mdns"
	"github.com/varunamachi/libx/errx"
)

type controller struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	AddrIP4   net.IP `json:"addrIP4"`
	Port      int    `json:"port"`
}

func discover(service string) ([]*controller, error) {
	ctls := make([]*controller, 0, 5)

	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go func() {
		for entry := range entriesCh {
			shortName := entry.Host
			comps := strings.Split(entry.Host, ".")
			if len(comps) > 0 {
				shortName = comps[0]
			}
			ctls = append(ctls, &controller{
				Name:      entry.Host,
				ShortName: shortName,
				AddrIP4:   entry.AddrV4,
				Port:      entry.Port,
			})
		}
	}()

	// Start the lookup
	// mdns.Lookup("_googlecast._tcp.local", entriesCh)

	err := mdns.Query(&mdns.QueryParam{
		Service:             service,
		Domain:              "._tcp.local",
		Timeout:             time.Second * 3,
		Entries:             entriesCh,
		WantUnicastResponse: false, // TODO - check if 'true' works
		DisableIPv4:         false,
		DisableIPv6:         false,
	})
	if err != nil {
		return nil, errx.Errf(err, "failed to discover service nodes")
	}
	close(entriesCh)
	return ctls, nil
}
