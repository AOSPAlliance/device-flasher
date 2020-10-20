package device

import "github.com/sirupsen/logrus"

type Hook func(device *Device, logger *logrus.Logger) error

type CustomHooks struct {
	DiscoveryPost     Hook
	FlashingPreUnlock Hook
	FlashingPreLock   Hook
}
