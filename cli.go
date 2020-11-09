// +build !GUI

package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gitlab.com/calyxos/device-flasher/internal/color"
	"gitlab.com/calyxos/device-flasher/internal/udev"
)

func setupUdev(logger *logrus.Logger) error {
	if hostOS == "linux" {
		err := udev.Setup(logger, "sudo", udev.DefaultUDevRules)
		if err != nil {
			return fmt.Errorf("failed to setup udev: %v", err)
		}
		cleanupPaths = append(cleanupPaths, udev.TempRulesFile)
	}
	return nil
}

func preparation(logger *logrus.Logger) {
	logger.Info(color.Yellow("1. Connect to a wifi network and ensure that no SIM cards are installed"))
	logger.Info(color.Yellow("2. Enable Developer Options on device (Settings -> About Phone -> tap \"Build number\" 7 times)"))
	logger.Info(color.Yellow("3. Enable USB debugging on device (Settings -> System -> Advanced -> Developer Options) and allow the computer to debug (hit \"OK\" on the popup when USB is connected)"))
	logger.Info(color.Yellow("4. Enable OEM Unlocking (in the same Developer Options menu)"))
	logger.Info(color.Yellow("Press ENTER to continue"))
	_, _ = fmt.Scanln()
}

func confirmFlash(logger *logrus.Logger) {
	logger.Info(color.Yellow("Flashing the following device(s):"))
	for _, d := range flashableDevices {
		logger.Infof(color.Yellow("📲 id=%v codename=%v image=%v"), d.ID, d.Codename, factoryImages[string(d.Codename)].ImagePath)
	}
	logger.Info(color.Yellow("Press ENTER to continue"))
	_, _ = fmt.Scanln()
}

func execute(logger *logrus.Logger) error {
	cleanupOnCtrlC()
	defer cleanup()

	err := pathValidation()
	if err != nil {
		return err
	}
	err = imageDiscovery(logger)
	if err != nil {
		return err
	}
	err = setupUdev(logger)
	if err != nil {
		return err
	}
	err = setupPlatformTools(logger)
	if err != nil {
		return err
	}
	preparation(logger)
	err = deviceDiscovery(logger)
	if err != nil {
		return err
	}
	err = factoryImageExtraction(logger)
	if err != nil {
		return err
	}
	fmt.Println()
	confirmFlash(logger)
	err = flashDevices(logger)
	if err != nil {
		return err
	}
	return nil
}
