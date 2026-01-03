package main

import (
	"time"

	"github.com/robolivable/beaves/config"
	"github.com/robolivable/beaves/controller"
	"github.com/robolivable/beaves/log"
	"github.com/robolivable/beaves/radar"
)

type Beaves struct {
	Proximity radar.Proximity // proximity driver

	Delay time.Duration // minimum time to wait between operations
	last  time.Time
}

func (b *Beaves) Operate(s controller.Switch) error {
	if time.Now().Before(b.last.Add(b.Delay)) {
		return nil
	}
	log.Debug("pressing button")
	if err := s.On(time.Duration(1) * time.Second); err != nil {
		return err
	}
	if err := s.Off(time.Duration(1) * time.Second); err != nil {
		return err
	}
	b.last = time.Now()
	return nil
}

func (b *Beaves) Manage(s controller.Switch) error {
	log.Debug("managing switch on %s", s.String())
	events, err := b.Proximity.Search()
	if err != nil {
		return err
	}

eventloop:
	for {
		time.Sleep(time.Duration(config.RuntimeConfig.EventLoopDelayMs) * time.Millisecond)
		proc := []*radar.Event{}

	loaderloop:
		for {
			select {
			default:
				break loaderloop
			case event, ok := <-events:
				if !ok {
					break eventloop
				}
				proc = append(proc, event)
			}
		}

		if len(proc) == 0 {
			continue
		}

		event := proc[len(proc)-1]
		log.Debug("%s", event.String())

		switch event.Action {
		case radar.Entering, radar.Exiting:
			if err := b.Operate(s); err != nil {
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
	b := Beaves{
		Proximity: nbts,
		Delay:     time.Duration(config.RuntimeConfig.OperationDelayMs) * time.Millisecond,
	}
	if err := b.Manage(nor); err != nil {
		panic(err)
	}
}
