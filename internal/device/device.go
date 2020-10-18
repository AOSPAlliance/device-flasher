package device

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type ToolName string

const (
	ADB      ToolName = "adb"
	Fastboot ToolName = "fastboot"
)

type Codename string

type Device struct {
	ID            string
	Codename      Codename
	DiscoveryTool ToolName
	CustomHooks   *CustomHooks
}

func New(deviceId, codename, discoveryTool string, logger *logrus.Logger) *Device {
	d := &Device{
		ID:            deviceId,
		Codename:      Codename(codename),
		DiscoveryTool: ToolName(discoveryTool),
		CustomHooks:   DeviceHooks[Codename(codename)],
	}
	if hook, ok := DeviceHooks[d.Codename]; ok {
		if hook.DiscoveryPost != nil {
			hook.DiscoveryPost(d, logger)
		}
	}
	return d
}

func (d *Device) String() string {
	return fmt.Sprintf("id=%v codename=%v", d.ID, d.Codename)
}
