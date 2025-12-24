package controller

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

type Switch interface {
	On(Delay time.Duration) error
	Off(Delay time.Duration) error
	Toggle(Delay time.Duration) error
	String() string
}

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

type GPIO struct {
	pin  gpio.PinIO
	name SerialName
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
	if err := g.pin.Out(s.Level()); err != nil {
		return fmt.Errorf("failed to send '%+v' to %s: %w", s, g.name, err)
	}
	return nil
}

const (
	RelayTerminal       SerialName = "GPIO17"
	RelayBackupTerminal SerialName = "GPIO27"
)

type OptoRelay struct {
	state State
	gpio  GPIO
}

func (or *OptoRelay) String() string {
	return fmt.Sprintf("OptoRelay {state: %v, terminal: %s}", or.state, or.gpio.String())
}

func (or *OptoRelay) On(d time.Duration) error {
	if or.state == On {
		return nil
	}
	time.Sleep(d)
	if err := or.gpio.Send(On); err != nil {
		or.state = Error
		return fmt.Errorf("failed to turn on relay: %w", err)
	}
	or.state = On
	return nil
}

func (or *OptoRelay) Off(d time.Duration) error {
	if or.state == Off {
		return nil
	}
	time.Sleep(d)
	if err := or.gpio.Send(Off); err != nil {
		or.state = Error
		return fmt.Errorf("failed to turn off relay: %w", err)
	}
	or.state = Off
	return nil
}

func (or *OptoRelay) Toggle(d time.Duration) error {
	if !or.state.Valid() {
		return fmt.Errorf("unable to toggle invalid state: %+v", or.state)
	}
	toggle := On
	if or.state == toggle {
		toggle = Off
	}
	time.Sleep(d)
	if err := or.gpio.Send(toggle); err != nil {
		or.state = Error
		return fmt.Errorf("failed to toggle relay: %w", err)
	}
	or.state = toggle
	return nil
}

func NewOptoRelaySwitch() (*OptoRelay, error) {
	g := GPIO{}
	if err := g.Claim(RelayTerminal); err != nil {
		_err := fmt.Errorf("failed to initialize serial module on default terminal: %w", err)
		if bErr := g.Claim(RelayBackupTerminal); bErr != nil {
			_bErr := fmt.Errorf("failed to initialize serial module on backup terminal: %w", bErr)
			return &OptoRelay{}, fmt.Errorf("%w; %w", _err, _bErr)
		}
	}
	return &OptoRelay{state: g.Receive(), gpio: g}, nil
}
