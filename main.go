package main

import (
	"flag"
	"fmt"
	"github.com/aospalliance/device-flasher/internal/color"
	"github.com/aospalliance/device-flasher/internal/device"
	"github.com/aospalliance/device-flasher/internal/factoryimage"
	"github.com/aospalliance/device-flasher/internal/platformtools"
	"github.com/aospalliance/device-flasher/internal/platformtools/adb"
	"github.com/aospalliance/device-flasher/internal/platformtools/fastboot"
	"github.com/aospalliance/device-flasher/internal/udev"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

var Vendor string
var title = Vendor + " Installer"
var version string

var (
	path               string
	debug              bool
	parallel           bool
	hostOS             = runtime.GOOS
	adbTool            *adb.Tool
	fastbootTool       *fastboot.Tool
	devicesMap         map[string]*device.Device
	images             map[string]string
	flashableDevices   []*device.Device
	factoryImages      = map[string]*factoryimage.FactoryImage{}
	platformTools      *platformtools.PlatformTools
	cleanupPaths       []string
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
		logger.Fatal(color.Red(err))
	}
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
