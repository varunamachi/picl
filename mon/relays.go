package mon

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/stianeikeland/go-rpio/v4"
	"github.com/varunamachi/picl/cmn"
)

var (
	ErrRelayIndexExceeded    = errors.New("mon.relay.indexExceeded")
	ErrRelayCtlUninitialized = errors.New("mon.relay.uninitialized")
)

type RelayController struct {
	pins        []rpio.Pin
	inited      bool
	isNO        bool
	cachedState []bool
}

type RelayConfig struct {
	GpioPins       []uint8 `json:"gpioPins"`
	IsNormallyOpen bool    `json:"isNormallyOpen"`
}

// func NewDefaultRelayController() *RelayController {
// 	return NewRelayController(
// 		[]uint8{22, 23, 24, 25},
// 		true,
// 	)
// }

func NewRelayController(cfg *RelayConfig) (*RelayController, error) {
	if err := rpio.Open(); err != nil {
		return nil, err
	}

	rc := &RelayController{
		pins:        make([]rpio.Pin, len(cfg.GpioPins)),
		cachedState: make([]bool, len(cfg.GpioPins)),
		isNO:        cfg.IsNormallyOpen,
		inited:      false,
	}
	for idx, pin := range cfg.GpioPins {
		pin := rpio.Pin(pin)
		pin.Output()
		pin.Write(rc.toState(false)) //Initially relay is closed!
		rc.pins[idx] = pin
	}
	rc.inited = true
	return rc, nil
}

func (rc *RelayController) SetState(slot int, state bool) error {
	if !rc.inited {
		return cmn.Errf(ErrRelayCtlUninitialized,
			"relay controller has not been initialized")
	}

	if slot < 0 || slot >= len(rc.pins) {
		return cmn.Errf(ErrRelayIndexExceeded,
			"index is less than 0 or than number of relays (%d) ", len(rc.pins))
	}
	rc.pins[slot].Output()
	pinState := rc.toState(state)
	rc.pins[slot].Write(pinState)
	rc.cachedState[slot] = state
	return nil
}

func (rc *RelayController) RefreshStates() ([]bool, error) {
	if !rc.inited {
		return nil, cmn.Errf(ErrRelayCtlUninitialized,
			"relay controller has not been initialized")
	}

	for index, pin := range rc.pins {
		rc.cachedState[index] = rc.fromState(pin.Read())
	}
	return rc.cachedState, nil
}

func (rc *RelayController) GetState(slot int) (bool, error) {
	if !rc.inited {
		return false, cmn.Errf(ErrRelayCtlUninitialized,
			"relay controller has not been initialized")
	}

	if slot < 0 || slot >= len(rc.pins) {
		return false, cmn.Errf(ErrRelayIndexExceeded,
			"index is less than 0 or than number of relays (%d) ", len(rc.pins))
	}
	return rc.cachedState[slot], nil
}

func (rc *RelayController) GetStates() ([]bool, error) {
	if !rc.inited {
		return nil, cmn.Errf(ErrRelayCtlUninitialized,
			"relay controller has not been initialized")
	}
	return rc.cachedState, nil
}

func (rc *RelayController) Close() error {
	return rpio.Close()
}

func (rc *RelayController) toState(on bool) rpio.State {
	if rc.isNO {
		if on {
			return rpio.Low
		}
		return rpio.High
	} else {
		if on {
			return rpio.High
		}
		return rpio.Low
	}
}

func (rc *RelayController) fromState(state rpio.State) bool {
	if rc.isNO {
		return state == rpio.Low
	} else {
		return state != rpio.High
	}
}

func getRelayEndpoints(rc *RelayController) []*cmn.Endpoint {
	return []*cmn.Endpoint{
		{
			Method:    echo.POST,
			Path:      "/switch/:slot/:state",
			Category:  "switch",
			Desc:      "Turn the switch/relay on/off",
			Version:   "v1",
			NeedsAuth: false,
			Handler: func(etx echo.Context) error {
				if rc == nil {
					return &echo.HTTPError{
						Message: "gpio features not enabled",
						Code:    http.StatusInternalServerError,
					}
				}

				slotStr := etx.Param("slot")
				stateStr := etx.Param("state")

				slot, err := strconv.Atoi(slotStr)
				if err != nil {
					return &echo.HTTPError{
						Message:  "invalid slot number",
						Code:     http.StatusBadRequest,
						Internal: err,
					}
				}
				state := strings.EqualFold(stateStr, "true")

				if err = rc.SetState(slot, state); err != nil {
					return &echo.HTTPError{
						Message:  "failed to set state",
						Code:     http.StatusInternalServerError,
						Internal: err,
					}
				}

				vals, err := rc.GetStates()
				if err != nil {
					return &echo.HTTPError{
						Message:  "failed to get pin states",
						Code:     http.StatusInternalServerError,
						Internal: err,
					}
				}
				return etx.JSON(http.StatusOK, vals)
			},
		},
		{
			Method:    echo.POST,
			Path:      "/switch/all/:state",
			Category:  "switch",
			Desc:      "Turn all relays on or off",
			Version:   "v1",
			NeedsAuth: false,
			Handler: func(etx echo.Context) error {
				if rc == nil {
					return &echo.HTTPError{
						Message: "gpio features not enabled",
						Code:    http.StatusInternalServerError,
					}
				}

				stateStr := etx.Param("state")
				state := strings.EqualFold(stateStr, "true")

				for slot := range rc.pins {
					if err := rc.SetState(slot, state); err != nil {
						return &echo.HTTPError{
							Message:  "failed to set state",
							Code:     http.StatusInternalServerError,
							Internal: err,
						}
					}
				}

				vals, err := rc.GetStates()
				if err != nil {
					return &echo.HTTPError{
						Message:  "failed to get pin states",
						Code:     http.StatusInternalServerError,
						Internal: err,
					}
				}
				return etx.JSON(http.StatusOK, vals)
			},
		},
		{
			Method:    echo.GET,
			Path:      "/switch/:slot",
			Category:  "switch",
			Desc:      "Get stored state of the switch at a slot ",
			Version:   "v1",
			NeedsAuth: false,
			Handler: func(etx echo.Context) error {
				if rc == nil {
					return &echo.HTTPError{
						Message: "gpio features not enabled",
						Code:    http.StatusInternalServerError,
					}
				}

				slotStr := etx.Param("slot")
				slot, err := strconv.Atoi(slotStr)
				if err != nil {
					return &echo.HTTPError{
						Message:  "invalid slot number",
						Code:     http.StatusBadRequest,
						Internal: err,
					}
				}

				state, err := rc.GetState(slot)
				if err != nil {
					return &echo.HTTPError{
						Message:  "failed to get state",
						Code:     http.StatusInternalServerError,
						Internal: err,
					}
				}

				return etx.String(http.StatusOK, strconv.FormatBool(state))
			},
		},
		{
			Method:    echo.GET,
			Path:      "/switch",
			Category:  "switch",
			Desc:      "Get stored state of all switches",
			Version:   "v1",
			NeedsAuth: false,
			Handler: func(etx echo.Context) error {
				if rc == nil {
					return &echo.HTTPError{
						Message: "gpio features not enabled",
						Code:    http.StatusInternalServerError,
					}
				}

				vals, err := rc.GetStates()
				if err != nil {
					return &echo.HTTPError{
						Message:  "failed to get pin states",
						Code:     http.StatusInternalServerError,
						Internal: err,
					}
				}
				return etx.JSON(http.StatusOK, vals)
			},
		},
		{
			Method:    echo.GET,
			Path:      "/switch/count",
			Category:  "switch",
			Desc:      "Get number of switches",
			Version:   "v1",
			NeedsAuth: false,
			Handler: func(etx echo.Context) error {
				if rc == nil {
					return &echo.HTTPError{
						Message: "gpio features not enabled",
						Code:    http.StatusInternalServerError,
					}
				}

				return etx.String(http.StatusOK, strconv.Itoa(len(rc.pins)))
			},
		},
	}
}
