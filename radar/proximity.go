package radar

import (
	"fmt"
	"strings"
	"time"

	"github.com/robolivable/beaves/config"
	"github.com/robolivable/beaves/log"
	"tinygo.org/x/bluetooth"
)

type ID string

type Actor struct {
	ID   ID
	Name string
}

func (a *Actor) Known() bool {
	for _, id := range config.RuntimeConfig.Actors.Known {
		if strings.EqualFold(string(a.ID), id) {
			return true
		}
	}
	return false
}

type Action int

const (
	Entering Action = iota
	Exiting
)

func (a Action) String() string {
	if a == Entering {
		return "Entering"
	}
	return "Exiting"
}

func GetAction(connected bool) Action {
	if connected {
		return Entering
	}
	return Exiting
}

type Event struct {
	Actor *Actor

	Action Action

	Epoch time.Time
}

func (e *Event) String() string {
	return fmt.Sprintf("Event {actor: %+v, action: %+v, epoch: %+v}", e.Actor, e.Action.String(), e.Epoch)
}

type Payload struct {
	Recipient *Actor

	Header  string
	Message string
}

type Proximity interface {
	Search() (chan *Event, error)
	Message(Payload *Payload) error
}

type BTSentry struct {
	adapter                    *bluetooth.Adapter
	advertisementName          string
	advertisementDelayMs       int
	connectionPoolSize         int
	serviceUUID                bluetooth.UUID
	indicateCharacteristicUUID bluetooth.UUID
	indicateCharacteristic     *bluetooth.Characteristic

	disconnectionLimitDelayMs int
}

func (bts *BTSentry) Search() (chan *Event, error) {
	response := make(chan *Event, bts.connectionPoolSize)
	bts.adapter.SetConnectHandler(func(device bluetooth.Device, connected bool) {
		log.InfoMemoize("new connection {device: %+v, connected: %t}", device, connected)
		if len(response) == bts.connectionPoolSize {
			// NOTE: this is a DDoS guard
			time.Sleep(time.Duration(100) * time.Millisecond)
			device.Disconnect()
			return
		}
		actor := Actor{
			ID:   ID(device.Address.String()),
			Name: device.Address.String(),
		}
		if !actor.Known() {
			log.InfoMemoize("unknown actor: %v", actor)
			go func() {
				time.Sleep(time.Duration(bts.disconnectionLimitDelayMs) * time.Millisecond)
				device.Disconnect()
			}()
			return
		}
		go func() {
			response <- &Event{
				Actor:  &actor,
				Action: GetAction(connected),
				Epoch:  time.Now(),
			}
		}()
	})
	advertisement := bts.adapter.DefaultAdvertisement()
	go func() {
		defer func() {
			log.Info("closing resposne channel")
			close(response)
		}()
		for {
			if err := advertisement.Configure(bluetooth.AdvertisementOptions{
				LocalName:         bts.advertisementName,
				AdvertisementType: bluetooth.AdvertisingTypeInd,
			}); err != nil {
				log.Error(err.Error())
				return
			}
			log.Info("configured %s", bts.advertisementName)
			if err := advertisement.Start(); err != nil {
				log.Error(err.Error())
				return
			}
			log.Info("advertising %s", bts.advertisementName)
			time.Sleep(time.Duration(bts.advertisementDelayMs) * time.Millisecond)
			if err := advertisement.Stop(); err != nil {
				log.Error(err.Error())
				return
			}
			log.Info("stopped advertising %s", bts.advertisementName)
		}
	}()
	return response, nil
}

func (bts *BTSentry) Message(payload *Payload) error {
	m := []byte(fmt.Sprintf("%s %s", payload.Header, payload.Message))
	if _, err := bts.indicateCharacteristic.Write(m); err != nil {
		return err
	}
	return nil
}

func NewBTSentry(config config.Bluetooth) (*BTSentry, error) {
	serviceUUID, _ := bluetooth.ParseUUID(config.ServiceID)
	characteristicUUID, _ := bluetooth.ParseUUID(config.IndicateCharacteristicID)
	adapter := bluetooth.DefaultAdapter
	if err := adapter.Enable(); err != nil {
		return nil, err
	}
	return &BTSentry{
		adapter:                    adapter,
		advertisementName:          config.AdvertisementName,
		advertisementDelayMs:       config.AdvertisementDelayMs,
		connectionPoolSize:         config.ConnectionPoolSize,
		serviceUUID:                serviceUUID,
		indicateCharacteristicUUID: characteristicUUID,
		indicateCharacteristic:     &bluetooth.Characteristic{},
		disconnectionLimitDelayMs:  config.DisconnectionDelayMs,
	}, nil
}
