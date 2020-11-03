package main

import (
	"context"
	"flag"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/storage"
	"fyne.io/fyne/widget"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"gitlab.com/calyxos/device-flasher/internal/color"
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
	"image"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	path         string
	debug        bool
	parallel     bool
	hostOS       = runtime.GOOS
	adbTool      *adb.Tool
	cleanupPaths []string
	version      string
	enableColorsStdout = true
)

func parseFlags() {
	flag.StringVar(&path, "image", "", "factory image zip file")
	flag.BoolVar(&debug, "debug", false, "debug logging")
	flag.BoolVar(&parallel, "parallel", false, "enables flashing of multiple devices at once")
	flag.Parse()
}

func main() {
	colorable.EnableColorsStdout(&enableColorsStdout)
	fmt.Println(color.Blue("Android Factory Image Flasher v" + version))
	parseFlags()

	if path == "" && parallel == false {
		a := app.New()
		w := a.NewWindow("Android Factory Image Flasher")
		selectedFile := canvas.NewText(path, image.White)
		flashButton := widget.NewButton("Flash", func() {
			w.Close()
		})
		flashButton.Disable()
		w.SetContent(
			container.NewVBox(
				container.NewHBox(
					widget.NewButton("Select", func() {
						w.Resize(fyne.Size{
							Width:  400,
							Height: 400,
						})
						d := dialog.NewFileOpen(
							func(file fyne.URIReadCloser, err error) {
								if file != nil {
									path = strings.ReplaceAll(file.URI().String(), "file://", "")
									selectedFile.Text = path
									flashButton.Enable()
									w.Resize(fyne.Size{
										Width:  400,
										Height: 100,
									})
								}
							}, w)
						wd, _ := os.Getwd()
						lister, _ := storage.ListerForURI(storage.NewFileURI(wd))
						d.SetLocation(lister)
						d.SetFilter(storage.NewExtensionFileFilter([]string{".zip", ".tar.xz", ".tgz"}))
						d.Show()
					}),
					selectedFile,
				),
				layout.NewSpacer(),
				flashButton,
			),
		)
		w.Resize(fyne.Size{
			Width:  400,
			Height: 100,
		})
		w.SetCloseIntercept(func() {
			os.Exit(0)
		})
		w.ShowAndRun()
	}

	logger := logrus.New()
	formatter := &prefixed.TextFormatter{ForceColors: true, ForceFormatting: true}
	formatter.SetColorScheme(&prefixed.ColorScheme{
		PrefixStyle: "white",
	})
	logger.SetFormatter(formatter)
	logger.SetOutput(colorable.NewColorableStdout())
	if debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	err := execute(logger)
	if err != nil {
		logger.Error(color.Red(err))
	}
	fmt.Println()
	fmt.Print(color.Red("Press ENTER to exit"))
	_, _ = fmt.Scanln()
	os.Exit(1)
}

func execute(logger *logrus.Logger) error {
	cleanupOnCtrlC()
	defer cleanup()

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

	// image discovery
	logger.Debug("running image discovery")
	images, err := imagediscovery.Discover(path)
	if err != nil {
		return fmt.Errorf("image discovery failed for %v: %v", path, err)
	}

	// setup udev if running linux
	if hostOS == "linux" {
		err := udev.Setup(logger, udev.DefaultUDevRules)
		if err != nil {
			return fmt.Errorf("failed to setup udev: %v", err)
		}
		cleanupPaths = append(cleanupPaths, udev.TempRulesFile)
	}

	// platform tools setup
	logger.Debug("setting up platformtools")
	toolsVersion := getToolsVersion(pathInfo)
	toolZipCacheDir, tmpToolExtractDir, err := platformToolsDirs(string(toolsVersion))
	if err != nil {
		return fmt.Errorf("failed to setup platformtools temp directories: %v", err)
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
	fastbootTool, err := fastboot.New(platformTools.Path(), hostOS)
	if err != nil {
		return fmt.Errorf("failed to setup fastboot: %v", err)
	}

	// device discovery
	logger.Info(color.Yellow("1. Connect to a wifi network and ensure that no SIM cards are installed"))
	logger.Info(color.Yellow("2. Enable Developer Options on device (Settings -> About Phone -> tap \"Build number\" 7 times)"))
	logger.Info(color.Yellow("3. Enable USB debugging on device (Settings -> System -> Advanced -> Developer Options) and allow the computer to debug (hit \"OK\" on the popup when USB is connected)"))
	logger.Info(color.Yellow("4. Enable OEM Unlocking (in the same Developer Options menu)"))
	logger.Info(color.Yellow("Press ENTER to continue"))
	_, _ = fmt.Scanln()
	devicesMap, err := devicediscovery.New(adbTool, fastbootTool, logger).DiscoverDevices()
	if err != nil {
		return fmt.Errorf("failed to run device discovery: %v", err)
	}
	logger.Info("Discovered the following device(s):")
	for _, device := range devicesMap {
		logger.Infof("📲 id=%v codename=%v (%v)", device.ID, device.Codename, device.DiscoveryTool)
	}
	fmt.Println()

	// factory image extraction
	flashableDevices := []*device.Device{}
	factoryImages := map[string]*factoryimage.FactoryImage{}
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

		err = factoryImage.Extract()
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

	// flash devices
	fmt.Println()
	logger.Info(color.Yellow("Flashing the following device(s):"))
	for _, d := range flashableDevices {
		logger.Infof(color.Yellow("📲 id=%v codename=%v image=%v"), d.ID, d.Codename, factoryImages[string(d.Codename)].ImagePath)
	}
	logger.Info(color.Yellow("Press ENTER to continue"))
	_, _ = fmt.Scanln()
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
	err = g.Wait()
	if err != nil {
		return fmt.Errorf("device flashing error: %v", err)
	}
	return nil
}

func getToolsVersion(pathInfo os.FileInfo) platformtools.SupportedVersion {
	// TODO: jasmine specific hack
	toolsVersion := platformtools.Version_30_0_4
	if !pathInfo.IsDir() && strings.Contains(pathInfo.Name(), "jasmine") {
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
	cleanupPaths = append(cleanupPaths, tmpToolExtractDir)
	return tmpToolExtractDir, nil
}

func cleanup() {
	if adbTool != nil {
		err := adbTool.KillServer()
		if err != nil {
			fmt.Printf("cleanup error killing adb server: %v\n", err)
		}
	}
	for _, path := range cleanupPaths {
		err := os.RemoveAll(path)
		if err != nil {
			fmt.Printf("cleanup error removing path %v: %v\n", path, err)
		}
	}
	if hostOS == "linux" {
		_, err := os.Stat(udev.RulesPath + udev.RulesFile)
		if !os.IsNotExist(err) {
			err = exec.Command("sudo", "rm", udev.RulesPath+udev.RulesFile).Run()
			if err != nil {
				fmt.Printf("cleanup error removing udev rules file: %v\n", err)
			}
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
