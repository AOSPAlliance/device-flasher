//go:generate mockgen -destination=mocks/mocks.go -package=mocks . FactoryImageFlasher,PlatformToolsFlasher,ADBFlasher,FastbootFlasher
package flash

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/aospalliance/device-flasher/internal/color"
	"github.com/aospalliance/device-flasher/internal/device"
	"github.com/aospalliance/device-flasher/internal/platformtools"
	"github.com/aospalliance/device-flasher/internal/platformtools/fastboot"
	"time"
)

const (
	DefaultLockUnlockValidationPause = time.Second * 5
	DefaultLockUnlockRetries         = 2
	DefaultLockUnlockRetryInterval   = time.Second * 30
)

var (
	ErrMaxLockUnlockRetries = errors.New("max retries reached")
)

type FactoryImageFlasher interface {
	Validate(codename device.Codename) error
	FlashAll(device *device.Device, path platformtools.PlatformToolsPath) error
}

type PlatformToolsFlasher interface {
	Path() platformtools.PlatformToolsPath
}

type ADBFlasher interface {
	RebootIntoBootloader(deviceId string) error
	KillServer() error
}

type FastbootFlasher interface {
	UnlockBootloader(deviceId string) error
	LockBootloader(deviceId string) error
	GetBootloaderLockStatus(deviceId string) (fastboot.FastbootLockStatus, error)
	Reboot(deviceId string) error
}

type Config struct {
	HostOS                    string
	FactoryImage              FactoryImageFlasher
	PlatformTools             PlatformToolsFlasher
	ADB                       ADBFlasher
	Fastboot                  FastbootFlasher
	Logger                    *logrus.Logger
	LockUnlockValidationPause time.Duration
	LockUnlockRetries         int
	LockUnlockRetryInterval   time.Duration
}

type Flash struct {
	hostOS                    string
	factoryImage              FactoryImageFlasher
	platformTools             PlatformToolsFlasher
	adb                       ADBFlasher
	fastboot                  FastbootFlasher
	logger                    *logrus.Logger
	lockUnlockValidationPause time.Duration
	lockUnlockRetries         int
	lockUnlockRetryInterval   time.Duration
}

func New(config *Config) *Flash {
	return &Flash{
		hostOS:                    config.HostOS,
		factoryImage:              config.FactoryImage,
		platformTools:             config.PlatformTools,
		adb:                       config.ADB,
		fastboot:                  config.Fastboot,
		logger:                    config.Logger,
		lockUnlockValidationPause: config.LockUnlockValidationPause,
		lockUnlockRetries:         config.LockUnlockRetries,
		lockUnlockRetryInterval:   config.LockUnlockRetryInterval,
	}
}

func (f *Flash) Flash(d *device.Device) error {
	logger := f.logger.WithFields(logrus.Fields{
		"prefix": d.String(),
	})

	logger.Info("validating factory image is for device")
	err := f.factoryImage.Validate(d.Codename)
	if err != nil {
		return err
	}

	if d.DiscoveryTool == device.ADB {
		logger.Info("reboot into bootloader")
		err = f.adb.RebootIntoBootloader(d.ID)
		if err != nil {
			f.logger.Debugf("ignoring adb reboot error and will attempt fastboot access: %v", err)
		}
	}

	logger.Info("checking bootloader status")
	lockStatus, err := f.fastboot.GetBootloaderLockStatus(d.ID)
	if err != nil {
		return err
	}
	if lockStatus != fastboot.Unlocked {
		logger.Info("starting unlocking bootloader process")
		logger.Info(color.Yellow("5. Please use the volume and power keys on the device to unlock the bootloader"))
		if d.CustomHooks != nil && d.CustomHooks.FlashingPreUnlock != nil {
			err := d.CustomHooks.FlashingPreUnlock(d, logger)
			if err != nil {
				logger.Warnf("pre unlock hook failed: %v", err)
			}
		}
		err := f.retrySetBootloaderStatus(d, fastboot.Unlocked)
		if err != nil {
			return err
		}
	}
	logger.Infof("bootloader is unlocked")

	logger.Info("running flash all script")
	err = f.factoryImage.FlashAll(d, f.platformTools.Path())
	if err != nil {
		return err
	}
	logger.Info("finished running flash all script")

	logger.Info("starting re-locking bootloader process")
	logger.Info(color.Yellow("6. Please use the volume and power keys on the device to lock the bootloader"))
	if d.CustomHooks != nil && d.CustomHooks.FlashingPreLock != nil {
		err := d.CustomHooks.FlashingPreLock(d, logger)
		if err != nil {
			logger.Warnf("pre lock hook failed: %v", err)
		}
	}
	err = f.retrySetBootloaderStatus(d, fastboot.Locked)
	if err != nil {
		return err
	}

	logger.Info("rebooting device")
	err = f.fastboot.Reboot(d.ID)
	if err != nil {
		logger.Warnf("failed to reboot device: %v. may need to manually reboot", err)
	}
	logger.Info(color.Yellow("7. Disable OEM unlocking from Developer Options after setting up your device"))

	return nil
}

func (f *Flash) retrySetBootloaderStatus(d *device.Device, wantedStatus fastboot.FastbootLockStatus) error {
	logger := f.logger.WithFields(logrus.Fields{
		"prefix": d.String(),
	})
	actionInProgress := "unlocking"
	actionComplete := "unlocked"
	bootloaderAction := f.fastboot.UnlockBootloader
	if wantedStatus == fastboot.Locked {
		actionInProgress = "locking"
		actionComplete = "locked"
		bootloaderAction = f.fastboot.LockBootloader
	}

	attempts := 0
	for {
		logger.Infof("%v bootloader", actionInProgress)
		err := bootloaderAction(d.ID)
		if err != nil {
			logger.Debugf("error %v bootloader: %v", actionInProgress, err)
			return err
		}
		logger.Debugf("waiting %v before checking bootloader status", f.lockUnlockValidationPause)
		time.Sleep(f.lockUnlockValidationPause)
		logger.Info("verifying bootloader status")
		lockStatus, err := f.fastboot.GetBootloaderLockStatus(d.ID)
		if err != nil {
			logger.Debugf("error verifying bootloader status: %v", err)
			return err
		}
		if lockStatus == wantedStatus {
			logger.Debugf("bootloader is now %v", actionComplete)
			break
		}
		if attempts >= f.lockUnlockRetries {
			logger.Debugf("max %v retries hit", actionInProgress)
			return fmt.Errorf("%w: %v", ErrMaxLockUnlockRetries, f.lockUnlockRetries)
		}
		logger.Infof("bootloader status is not %v yet. waiting %v before retrying", actionComplete, f.lockUnlockRetryInterval)
		time.Sleep(f.lockUnlockRetryInterval)
		attempts++
	}
	return nil
}
