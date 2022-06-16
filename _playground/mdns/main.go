package main

import (
	"fmt"
	"time"

	"github.com/hashicorp/mdns"
)

func main() {
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go func() {
		for entry := range entriesCh {
			fmt.Printf("Got new entry: %s : %v : %v\n",
				entry.Host,
				entry.AddrV4,
				entry.Port)
			for _, f := range entry.InfoFields {
				fmt.Println("\t", f)
			}
		}
	}()

	// Start the lookup
	// mdns.Lookup("_googlecast._tcp.local", entriesCh)

	err := mdns.Query(&mdns.QueryParam{
		Service:             "_relayctl",
		Domain:              "._tcp.local",
		Timeout:             time.Second * 3,
		Entries:             entriesCh,
		WantUnicastResponse: false, // TODO(reddaly): Change this default.
		DisableIPv4:         false,
		DisableIPv6:         false,
	})
	if err != nil {
		fmt.Println(err)
	}
	close(entriesCh)
}
