package controller

import (
	"fmt"
	"time"

	"github.com/robolivable/beaves/log"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

type State int

const (
	Unknown State = iota
	On
	Off
	Error
)

func (s State) Level() gpio.Level {
	switch s {
	case On:
		return gpio.High
	default:
		return gpio.Low
	}
}

func (s State) Valid() bool {
	switch s {
	case On, Off:
		return true
	default:
		return false
	}
}

func GetState(l gpio.Level) State {
	switch l {
	case gpio.High:
		return On
	case gpio.Low:
		return Off
	}
	return Unknown
}

type SerialName string

const (
	RelayTerminal       SerialName = "GPIO17"
	RelayBackupTerminal SerialName = "GPIO27"
)

type GPIO struct {
	pin  gpio.PinIO
	name SerialName

	debounce time.Duration
	last     time.Time
}

func (g *GPIO) String() string {
	return fmt.Sprintf("GPIO {name: %s}", g.name)
}

func (g *GPIO) Claim(sn SerialName) error {
	if _, err := host.Init(); err != nil {
		return fmt.Errorf("host failed to initialize while claiming %s: %w", sn, err)
	}
	if g.pin = gpioreg.ByName(string(sn)); g.pin == nil {
		return fmt.Errorf("failed to claim: pin %s is not present on host", sn)
	}
	g.name = sn
	return nil
}

func (g *GPIO) Receive() State {
	return GetState(g.pin.Read())
}

func (g *GPIO) Send(s State) error {
	if time.Now().Before(g.last.Add(g.debounce)) {
		log.DebugMemoize("GPIO: Send: debounced: %v", s)
		return nil
	}
	if err := g.pin.Out(s.Level()); err != nil {
		return fmt.Errorf("failed to send '%+v' to %s: %w", s, g.name, err)
	}
	g.last = time.Now()
	return nil
}
