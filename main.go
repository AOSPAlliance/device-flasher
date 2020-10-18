package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"gitlab.com/calyxos/device-flasher/internal/device"
	"gitlab.com/calyxos/device-flasher/internal/devicediscovery"
	"gitlab.com/calyxos/device-flasher/internal/factoryimage"
	"gitlab.com/calyxos/device-flasher/internal/flash"
	"gitlab.com/calyxos/device-flasher/internal/imagediscovery"
	"gitlab.com/calyxos/device-flasher/internal/platformtools"
	"gitlab.com/calyxos/device-flasher/internal/platformtools/adb"
	"gitlab.com/calyxos/device-flasher/internal/platformtools/fastboot"
	"gitlab.com/calyxos/device-flasher/internal/udev"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	path               string
	debug              bool
	hostOS             = runtime.GOOS
	adbTool            *adb.Tool
	cleanupDirectories []string
)

func parseFlags() {
	flag.StringVar(&path, "path", "", "factory image zip file or directory")
	flag.BoolVar(&debug, "debug", false, "debug logging")
	flag.Parse()
}

func main() {
	parseFlags()
	cleanupOnCtrlC()
	defer cleanup()

	logger := logrus.New()
	if debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	if path == "" {
		logger.Info("no -path was specified, using current working directory")
		dir, err := os.Getwd()
		if err != nil {
			logger.Fatalf("unable to get current working directory: %v", err)
		}
		path = dir
	}

	// check path exists
	pathInfo, err := os.Stat(path)
	if err != nil {
		logger.Fatalf("unable to find provided path %v: %v", path, err)
	}

	// image discovery
	logger.Debug("running image discovery")
	images, err := imagediscovery.Discover(path)
	if err != nil {
		logger.Fatalf("image discovery failed for path %v: %v", path, err)
	}

	// setup udev if running linux
	if hostOS == "linux" {
		err := udev.Setup(logger, udev.DefaultUDevRules)
		if err != nil {
			logger.Fatalf("failed to setup udev: %v", err)
		}
	}

	// platform tools setup
	logger.Debug("setting up platformtools")
	toolsVersion := getToolsVersion(pathInfo)
	toolZipCacheDir, tmpToolExtractDir, err := platformToolsDirs(string(toolsVersion))
	if err != nil {
		logger.Fatalf("failed to setup platformtools temp directories: %v", err)
	}
	platformTools, err := platformtools.New(&platformtools.Config{
		CacheDir:             toolZipCacheDir,
		HttpClient:           &http.Client{Timeout: time.Minute * 5},
		HostOS:               hostOS,
		ToolsVersion:         toolsVersion,
		DestinationDirectory: tmpToolExtractDir,
		Logger:               logger,
	})
	if err != nil {
		logger.Fatalf("failed to setup platformtools: %v", err)
	}

	// adb setup
	logger.Debug("setting up adb")
	adbTool, err = adb.New(platformTools.Path(), hostOS)
	if err != nil {
		logger.Fatalf("failed to setup adb: %v", err)
	}
	err = adbTool.KillServer()
	if err != nil {
		logger.Debugf("failed to kill adb server: %v", err)
	}
	err = adbTool.StartServer()
	if err != nil {
		logger.Fatalf("failed to start adb server: %v", err)
	}

	// fastboot setup
	logger.Debug("setting up fastboot")
	fastbootTool, err := fastboot.New(platformTools.Path(), hostOS)
	if err != nil {
		logger.Fatalf("failed to setup fastboot: %v", err)
	}

	// device discovery
	logger.Info("-> Connect to a wifi network and ensure that no SIM cards are installed")
	logger.Info("-> Enable Developer Options on device (Settings -> About Phone -> tap \"Build number\" 7 times)")
	logger.Info("-> Enable USB debugging on device (Settings -> System -> Advanced -> Developer Options) and allow the computer to debug (hit \"OK\" on the popup when USB is connected)")
	logger.Info("-> Enable OEM Unlocking (in the same Developer Options menu)")
	logger.Info("Press ENTER to continue")
	_, _ = fmt.Scanln()
	devicesMap, err := devicediscovery.New(adbTool, fastbootTool, logger).DiscoverDevices()
	if err != nil {
		logger.Fatalf("failed to run device discovery: %v", err)
	}
	logger.Info("Discovered the following device(s):")
	for _, device := range devicesMap {
		logger.Infof("📲 id=%v codename=%v (%v)", device.ID, device.Codename, device.DiscoveryTool)
	}
	logger.Info("")

	// factory image extraction
	flashableDevices := []*device.Device{}
	factoryImages := map[string]*factoryimage.FactoryImage{}
	for _, d := range devicesMap {
		if _, ok := images[string(d.Codename)]; !ok {
			logger.Warnf("no image discovered for device id=%v codename=%v", d.ID, d.Codename)
			continue
		}
		logger.Debug("creating temporary directory for extracting factory image")
		tmpFactoryDir, err := tempExtractDir("factory")
		if err != nil {
			logger.Fatalf("failed to create temp dir for factory image: %v", err)
		}
		factoryImages[string(d.Codename)], err = factoryimage.New(&factoryimage.Config{
			HostOS:           hostOS,
			ImagePath:        images[string(d.Codename)],
			WorkingDirectory: tmpFactoryDir,
			Logger:           logger,
		})
		if err != nil {
			logger.Fatalf("failed to extract factory image: %v", err)
		}
		flashableDevices = append(flashableDevices, d)
	}
	if len(flashableDevices) <= 0 {
		logger.Fatal("there are no flashable devices")
	}

	// flash devices
	logger.Info("")
	logger.Info("Flashing the following device(s):")
	for _, d := range flashableDevices {
		logger.Infof("📲 id=%v codename=%v image=%v", d.ID, d.Codename, factoryImages[string(d.Codename)].ImagePath)
	}
	logger.Info("Press ENTER to continue")
	_, _ = fmt.Scanln()
	g, _ := errgroup.WithContext(context.Background())
	for _, d := range flashableDevices {
		currentDevice := d
		g.Go(func() error {
			deviceLogger := logger.WithFields(logrus.Fields{
				"deviceId":       currentDevice.ID,
				"deviceCodename": currentDevice.Codename,
			})
			deviceLogger.Infof("starting to flash device")
			err := flash.New(&flash.Config{
				HostOS:        hostOS,
				FactoryImage:  factoryImages[string(currentDevice.Codename)],
				PlatformTools: platformTools,
				ADB:           adbTool,
				Fastboot:      fastbootTool,
				Logger:        logger,
			}).Flash(currentDevice)
			if err != nil {
				deviceLogger.Error(err)
				return err
			}
			deviceLogger.Infof("finished flashing device")
			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		logger.Fatal(err)
	}
}

func getToolsVersion(pathInfo os.FileInfo) platformtools.SupportedVersion {
	// TODO: jasmine specific hack
	toolsVersion := platformtools.Version_30_0_4
	if strings.Contains(pathInfo.Name(), "jasmine") {
		toolsVersion = platformtools.Version_29_0_6
	}
	return toolsVersion
}

func platformToolsDirs(toolsVersion string) (string, string, error) {
	toolZipCacheDir := filepath.Join(os.TempDir(), "platform-tools", toolsVersion)
	err := os.MkdirAll(toolZipCacheDir, os.ModePerm)
	if err != nil {
		return "", "", fmt.Errorf("failed to setup tools cache dir %v: %w", toolZipCacheDir, err)
	}
	tmpToolExtractDir, err := tempExtractDir("platformtools")
	if err != nil {
		return "", "", err
	}
	return toolZipCacheDir, tmpToolExtractDir, nil
}

func tempExtractDir(usage string) (string, error) {
	tmpToolExtractDir, err := ioutil.TempDir("", fmt.Sprintf("device-flasher-extracted-%v", usage))
	if err != nil {
		return "", err
	}
	cleanupDirectories = append(cleanupDirectories, tmpToolExtractDir)
	return tmpToolExtractDir, nil
}

func cleanup() {
	if adbTool != nil {
		err := adbTool.KillServer()
		if err != nil {
			fmt.Printf("cleanup error killing adb server: %v\n", err)
		}
	}
	for _, dir := range cleanupDirectories {
		err := os.RemoveAll(dir)
		if err != nil {
			fmt.Printf("cleanup error removing dir %v: %v\n", dir, err)
		}
	}
}

func cleanupOnCtrlC() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		cleanup()
		os.Exit(0)
	}()
}
