package main

import (
	"time"

	"github.com/robolivable/beaves/config"
	"github.com/robolivable/beaves/controller"
	"github.com/robolivable/beaves/log"
	"github.com/robolivable/beaves/radar"
)

type Beaves struct {
	sentry radar.Proximity
}

func (b *Beaves) Manage(s controller.Switch) error {
	log.Debug("managing switch on %s", s.String())
	events, err := b.sentry.Search()
	if err != nil {
		return err
	}
	for event := range events {
		log.Info("%s", event.String())
		switch event.Action {
		case radar.Entering:
			log.Debug("openning relay")
			if err := s.On(time.Duration(1) * time.Second); err != nil {
				log.Error(err.Error())
				continue
			}
		case radar.Exiting:
			log.Debug("closing relay")
			if err := s.Off(time.Duration(1) * time.Second); err != nil {
				log.Error(err.Error())
				continue
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
