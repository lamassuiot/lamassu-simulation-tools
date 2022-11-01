package store

import (
	"github.com/lamassuiot/lamassu-vdevice/pkg/model"
)

type DeviceStateStore struct {
	device                   *model.DeviceState
	updateDeviceStateChannel chan model.DeviceState
}

func New() (*DeviceStateStore, chan model.DeviceState) {
	deviceStateChannel := make(chan model.DeviceState)

	return &DeviceStateStore{
		updateDeviceStateChannel: deviceStateChannel,
	}, deviceStateChannel
}

func (d *DeviceStateStore) GetDeviceState() *model.DeviceState {
	return d.device
}

func (d *DeviceStateStore) SetDeviceState(device *model.DeviceState) {
	d.device = device
	go func() { d.updateDeviceStateChannel <- *device }()
}
