package device

import "github.com/sirupsen/logrus"

type Hook func(device *Device, logger *logrus.Entry) error

type CustomHooks struct {
	DiscoveryPost     func(device *Device, logger *logrus.Logger) error
	FlashingPreUnlock Hook
	FlashingPreLock   Hook
}
