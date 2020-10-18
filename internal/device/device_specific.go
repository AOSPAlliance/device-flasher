package device

import (
	"github.com/sirupsen/logrus"
)

const (
	DeviceJasmine Codename = "jasmine"
	DeviceWalleye Codename = "walleye"
)

var DeviceHooks = map[Codename]*CustomHooks{
	DeviceJasmine: jasmineHooks,
	DeviceWalleye: walleyeHooks,
}

var (
	jasmineHooks = &CustomHooks{
		DiscoveryPost: func(device *Device, logger *logrus.Logger) error {
			device.Codename = "jasmine_sprout"
			logger.Debugf("updated device codename: %v", device.Codename)
			return nil
		},
		FlashingPreUnlock: additionalLockUnlockStep,
		FlashingPreLock:   additionalLockUnlockStep,
	}
	walleyeHooks = &CustomHooks{
		FlashingPreUnlock: additionalLockUnlockStep,
		FlashingPreLock:   additionalLockUnlockStep,
	}
)

func additionalLockUnlockStep(device *Device, logger *logrus.Logger) error {
	logger.Info("-> Once device boots, disconnect its cable and power it off")
	logger.Info("-> Then, press volume down + power to boot it into fastboot mode, and connect the cable again.")
	logger.Info("The installation will resume automatically")
	return nil
}
