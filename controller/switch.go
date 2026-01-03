package controller

import (
	"fmt"
	"time"

	"github.com/robolivable/beaves/config"
	"github.com/robolivable/beaves/log"
)

type Switch interface {
	On(Delay time.Duration) error
	Off(Delay time.Duration) error
	Toggle(Delay time.Duration) error
	String() string
}

type OptoRelay struct {
	state State
	gpio  GPIO
}

func (or *OptoRelay) String() string {
	return fmt.Sprintf("OptoRelay {state: %v, terminal: %s}", or.state, or.gpio.String())
}

func (or *OptoRelay) On(d time.Duration) error {
	log.Debug("OptoRelay.On: %s", or.String())
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
	log.Debug("OptoRelay.Off: %s", or.String())
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
	log.Debug("OptoRelay.Toggle: %s", or.String())
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
	g := GPIO{debounce: time.Duration(config.RuntimeConfig.RelayDebounceMs) * time.Millisecond}
	if err := g.Claim(RelayTerminal); err != nil {
		_err := fmt.Errorf("failed to initialize serial module on default terminal: %w", err)
		if bErr := g.Claim(RelayBackupTerminal); bErr != nil {
			_bErr := fmt.Errorf("failed to initialize serial module on backup terminal: %w", bErr)
			return &OptoRelay{}, fmt.Errorf("%w; %w", _err, _bErr)
		}
	}
	return &OptoRelay{state: g.Receive(), gpio: g}, nil
}
