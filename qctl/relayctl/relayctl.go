package relayctl

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/mdns"
	"github.com/varunamachi/libx/errx"
	"github.com/varunamachi/libx/httpx"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

type controller struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	Host      string `json:"host"`
	AddrIP4   net.IP `json:"addrIP4"`
	Port      int    `json:"port"`
}

func discover(service string) (<-chan *controller, error) {
	ctls := make(chan *controller, 4)
	entriesCh := make(chan *mdns.ServiceEntry)
	go func() {
		defer close(ctls)
		for entry := range entriesCh {
			if strings.Contains(entry.Name, service) {
				shortName := entry.Host
				comps := strings.Split(entry.Host, ".")
				if len(comps) > 0 {
					shortName = comps[0]
				}

				ctls <- &controller{
					Name:      entry.Name,
					ShortName: shortName,
					Host:      entry.Host,
					AddrIP4:   entry.AddrV4,
					Port:      entry.Port,
				}

			}
			// iox.PrintJSON(entry)
		}
	}()

	err := mdns.Query(&mdns.QueryParam{
		Service:             service,
		Domain:              "._tcp.local",
		Timeout:             time.Second * 3,
		Entries:             entriesCh,
		WantUnicastResponse: true,
		DisableIPv4:         false,
		DisableIPv6:         false,
	})
	if err != nil {
		return nil, errx.Errf(err, "failed to discover service nodes")
	}
	defer close(entriesCh)
	return ctls, nil
}

type ctlClientWrapper struct {
	client *httpx.Client
	ctlr   *controller
	err    error
}

func getClientTo(service, inShortName string) (<-chan ctlClientWrapper, error) {
	outChan := make(chan ctlClientWrapper)
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go func() {
		defer close(outChan)
		for entry := range entriesCh {
			if strings.Contains(entry.Name, service) {
				shortName := entry.Host
				comps := strings.Split(entry.Host, ".")
				if len(comps) > 0 {
					shortName = comps[0]
				}
				if inShortName == shortName {
					ctl := &controller{
						Name:      entry.Name,
						ShortName: shortName,
						Host:      entry.Host,
						AddrIP4:   entry.AddrV4,
						Port:      entry.Port,
					}

					hostAddr := fmt.Sprintf(
						"http://%v:%d", entry.AddrV4, entry.Port)
					client := httpx.New(hostAddr, "")
					outChan <- ctlClientWrapper{client, ctl, nil}
					return
				}
			}
		}
		outChan <- ctlClientWrapper{
			nil, nil, errors.New("controller not found"),
		}
	}()

	err := mdns.Query(&mdns.QueryParam{
		Service:             service,
		Domain:              "._tcp.local",
		Timeout:             time.Second * 3,
		Entries:             entriesCh,
		WantUnicastResponse: true,
		DisableIPv4:         false,
		DisableIPv6:         false,
	})
	if err != nil {
		return nil, errx.Errf(err, "failed to discover service nodes")
	}
	defer close(entriesCh)

	return outChan, nil
}
