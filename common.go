package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"gitlab.com/calyxos/device-flasher/internal/devicediscovery"
	"gitlab.com/calyxos/device-flasher/internal/factoryimage"
	"gitlab.com/calyxos/device-flasher/internal/flash"
	"gitlab.com/calyxos/device-flasher/internal/imagediscovery"
	"gitlab.com/calyxos/device-flasher/internal/platformtools"
	"gitlab.com/calyxos/device-flasher/internal/platformtools/adb"
	"gitlab.com/calyxos/device-flasher/internal/platformtools/fastboot"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"time"
)

func pathValidation() error {
	// check path is provided
	if path == "" {
		return fmt.Errorf("-image flag must be specified")
	}

	// check path exists
	pathInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("unable to find provided path %v: %v", path, err)
	}

	// non parallel only supports passing a file to be more explicit
	if !parallel && pathInfo.IsDir() {
		return fmt.Errorf("-image must be a file (not a directory)")
	}
	return nil
}

func imageDiscovery(logger *logrus.Logger) error {
	logger.Debug("running image discovery")
	var err error
	images, err = imagediscovery.Discover(path)
	if err != nil {
		return fmt.Errorf("image discovery failed for %v: %v", path, err)
	}
	return nil
}

func setupPlatformTools(logger *logrus.Logger) error {
	logger.Debug("setting up platformtools")
	toolsVersion := getToolsVersion(path)
	toolZipCacheDir, tmpToolExtractDir, err := platformToolsDirs(string(toolsVersion))
	if err != nil {
		return fmt.Errorf("failed to setup platformtools temp directories: %v", err)
	}
	platformTools, err = platformtools.New(&platformtools.Config{
		CacheDir:             toolZipCacheDir,
		HttpClient:           &http.Client{Timeout: time.Minute * 5},
		HostOS:               hostOS,
		ToolsVersion:         toolsVersion,
		DestinationDirectory: tmpToolExtractDir,
		Logger:               logger,
	})
	if err != nil {
		return fmt.Errorf("failed to setup platformtools: %v", err)
	}

	// adb setup
	logger.Debug("setting up adb")
	adbTool, err = adb.New(platformTools.Path(), hostOS)
	if err != nil {
		return fmt.Errorf("failed to setup adb: %v", err)
	}
	err = adbTool.KillServer()
	if err != nil {
		logger.Debugf("failed to kill adb server: %v", err)
	}
	err = adbTool.StartServer()
	if err != nil {
		return fmt.Errorf("failed to start adb server: %v", err)
	}

	// fastboot setup
	logger.Debug("setting up fastboot")
	fastbootTool, err = fastboot.New(platformTools.Path(), hostOS)
	if err != nil {
		return fmt.Errorf("failed to setup fastboot: %v", err)
	}
	return nil
}

func deviceDiscovery(logger *logrus.Logger) error {
	var err error
	devicesMap, err = devicediscovery.New(adbTool, fastbootTool, logger).DiscoverDevices()
	if err != nil {
		return fmt.Errorf("failed to run device discovery: %v", err)
	}
	logger.Info("Discovered the following device(s):")
	for _, device := range devicesMap {
		logger.Infof("📲 id=%v codename=%v (%v)", device.ID, device.Codename, device.DiscoveryTool)
	}
	fmt.Println()
	return nil
}

func factoryImageExtraction(logger *logrus.Logger) error {
	for _, d := range devicesMap {
		deviceLogger := logger.WithFields(logrus.Fields{"id": d.ID, "codename": d.Codename})
		if _, ok := images[string(d.Codename)]; !ok {
			deviceLogger.Warnf("no image discovered for device")
			continue
		}

		var factoryImage *factoryimage.FactoryImage
		if fi, ok := factoryImages[string(d.Codename)]; ok {
			deviceLogger.Debug("re-using existing factory image")
			factoryImage = fi
		} else {
			deviceLogger.Debug("creating temporary directory for extracting factory image for device")
			tmpFactoryDir, err := tempExtractDir("factory")
			if err != nil {
				return fmt.Errorf("failed to create temp dir for factory image: %v", err)
			}
			factoryImage = factoryimage.New(&factoryimage.Config{
				HostOS:           hostOS,
				ImagePath:        images[string(d.Codename)],
				WorkingDirectory: tmpFactoryDir,
				Logger:           logger,
			})
		}

		err := factoryImage.Extract()
		if err != nil {
			return fmt.Errorf("failed to extract factory image: %v", err)
		}

		factoryImages[string(d.Codename)] = factoryImage
		flashableDevices = append(flashableDevices, d)
	}
	if len(flashableDevices) <= 0 {
		return fmt.Errorf("there are no flashable devices")
	}
	if !parallel && len(flashableDevices) > 1 {
		return fmt.Errorf("discovered multiple devices and --parallel flag is not enabled")
	}
	return nil
}

func flashDevices(logger *logrus.Logger) error {
	g, _ := errgroup.WithContext(context.Background())
	for _, d := range flashableDevices {
		currentDevice := d
		g.Go(func() error {
			deviceLogger := logger.WithFields(logrus.Fields{
				"prefix": currentDevice.String(),
			})
			deviceLogger.Infof("starting to flash device")
			err := flash.New(&flash.Config{
				HostOS:                    hostOS,
				FactoryImage:              factoryImages[string(currentDevice.Codename)],
				PlatformTools:             platformTools,
				ADB:                       adbTool,
				Fastboot:                  fastbootTool,
				Logger:                    logger,
				LockUnlockValidationPause: flash.DefaultLockUnlockValidationPause,
				LockUnlockRetries:         flash.DefaultLockUnlockRetries,
				LockUnlockRetryInterval:   flash.DefaultLockUnlockRetryInterval,
			}).Flash(currentDevice)
			if err != nil {
				deviceLogger.Error(err)
				return err
			}
			deviceLogger.Infof("finished flashing device")
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		return fmt.Errorf("device flashing error: %v", err)
	}
	return nil
}
