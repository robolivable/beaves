package radar

import (
	"fmt"
	"time"

	"github.com/robolivable/beaves/config"
	"tinygo.org/x/bluetooth"
)

type ID string

type Actor struct {
	ID   ID
	Name string
}

func (a *Actor) Known() bool {
	for _, id := range config.RuntimeConfig.Actors.Known {
		if a.ID == ID(id) {
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
	connectionPoolSize         int
	serviceUUID                bluetooth.UUID
	indicateCharacteristicUUID bluetooth.UUID
	indicateCharacteristic     *bluetooth.Characteristic
}

func (bts *BTSentry) Search() (chan *Event, error) {
	if err := bts.adapter.AddService(&bluetooth.Service{
		UUID: bts.serviceUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:   bts.indicateCharacteristicUUID,
				Flags:  bluetooth.CharacteristicIndicatePermission,
				Handle: bts.indicateCharacteristic,
			},
		},
	}); err != nil {
		return nil, err
	}
	advertisement := bts.adapter.DefaultAdvertisement()
	if err := advertisement.Configure(bluetooth.AdvertisementOptions{
		LocalName:    bts.advertisementName,
		ServiceUUIDs: []bluetooth.UUID{bts.serviceUUID},
	}); err != nil {
		return nil, err
	}
	response := make(chan *Event, bts.connectionPoolSize)
	bts.adapter.SetConnectHandler(func(device bluetooth.Device, connected bool) {
		go func() {
			if len(response) == bts.connectionPoolSize {
				// NOTE: this is a DDoS guard
				device.Disconnect()
				return
			}
			actor := Actor{
				ID:   ID(device.Address.String()),
				Name: device.Address.String(),
			}
			if !actor.Known() {
				device.Disconnect()
				return
			}
			response <- &Event{
				Actor:  &actor,
				Action: GetAction(connected),
				Epoch:  time.Now(),
			}
		}()
	})
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
		connectionPoolSize:         config.ConnectionPoolSize,
		serviceUUID:                serviceUUID,
		indicateCharacteristicUUID: characteristicUUID,
		indicateCharacteristic:     &bluetooth.Characteristic{},
	}, nil
}
