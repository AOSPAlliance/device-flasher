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
		FlashingPreUnlock: additionalUnlockStep,
		FlashingPreLock:   additionalLockStep,
	}
	walleyeHooks = &CustomHooks{
		FlashingPreUnlock: additionalUnlockStep,
		FlashingPreLock:   additionalLockStep,
	}
)

func additionalUnlockStep(device *Device, logger *logrus.Logger) error {
	return loggingSteps(5, logger)
}

func additionalLockStep(device *Device, logger *logrus.Logger) error {
	return loggingSteps(6, logger)
}

func loggingSteps(step int, logger *logrus.Logger) error {
	logger.Infof("  %va. Once device boots, disconnect its cable and power it off", step)
	logger.Infof("  %vb. Then, press volume down + power to boot it into fastboot mode, and connect the cable again.", step)
	logger.Info("The installation will resume automatically")
	return nil
}
