//go:generate mockgen -destination=mocks/mocks.go -package=mocks . DeviceDiscoverer
package devicediscovery

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/aospalliance/device-flasher/internal/device"
)

var (
	ErrNoDevicesFound    = errors.New("no devices detected with adb or fastboot")
	ErrGetDevices        = errors.New("unable to get devices")
	ErrGetDeviceCodename = errors.New("unable to get device codename")
)

type DeviceDiscoverer interface {
	GetDeviceIds() ([]string, error)
	GetDeviceCodename(deviceId string) (string, error)
	Name() string
}

type Discovery struct {
	adbTool      DeviceDiscoverer
	fastbootTool DeviceDiscoverer
	logger       *logrus.Logger
}

func New(adbTool, fastbootTool DeviceDiscoverer, logger *logrus.Logger) *Discovery {
	return &Discovery{
		adbTool:      adbTool,
		fastbootTool: fastbootTool,
		logger:       logger,
	}
}

func (d *Discovery) DiscoverDevices() (map[string]*device.Device, error) {
	d.logger.Debugf("discovering adb devices")
	devices, err := d.getDevices(d.adbTool)
	if err != nil {
		d.logger.Warn(err)
	}

	d.logger.Debugf("discovering fastboot devices")
	fastbootDevices, err := d.getDevices(d.fastbootTool)
	if err != nil {
		d.logger.Warn(err)
	}

	for k, v := range fastbootDevices {
		devices[k] = v
	}

	if len(devices) == 0 {
		return nil, ErrNoDevicesFound
	}

	return devices, nil
}

func (d *Discovery) getDevices(tool DeviceDiscoverer) (map[string]*device.Device, error) {
	toolName := tool.Name()
	devices := map[string]*device.Device{}
	deviceIds, err := tool.GetDeviceIds()
	if err != nil {
		return nil, fmt.Errorf("%v %w: %v", toolName, ErrGetDevices, err)
	}
	for _, deviceId := range deviceIds {
		d.logger.Debugf("%v getting code name for device %v", toolName, deviceId)
		deviceCodename, err := tool.GetDeviceCodename(deviceId)
		if err != nil {
			d.logger.Warnf("%v skipping device %v as getting code name failed: %v", toolName, deviceId, err)
			continue
		}
		if _, ok := devices[deviceId]; ok {
			d.logger.Warnf("%v skipping duplicate device %v", toolName, deviceId)
			continue
		}

		devices[deviceId] = device.New(deviceId, deviceCodename, toolName, d.logger)
	}
	return devices, nil
}
