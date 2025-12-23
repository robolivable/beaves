package beaves

import (
	"time"

	"github.com/robolivable/beaves/config"
	"github.com/robolivable/beaves/controller"
	"github.com/robolivable/beaves/radar"
)

type Beaves struct {
	sentry radar.Proximity
}

func (b *Beaves) Manage(s controller.Switch) error {
	events, err := b.sentry.Search()
	if err != nil {
		return err
	}
	for event := range events {
		switch event.Action {
		case radar.Entering:
			if err := s.On(1 * time.Second); err != nil {
				return err
			}
		case radar.Exiting:
			if err := s.Off(1 * time.Second); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	nbts, err := radar.NewBTSentry(config.RuntimeConfig.Bluetooth)
	if err != nil {
		panic(err)
	}
	nor, err := controller.NewOptoRelaySwitch()
	if err != nil {
		panic(err)
	}
	b := Beaves{sentry: nbts}
	if err := b.Manage(nor); err != nil {
		panic(err)
	}
}
