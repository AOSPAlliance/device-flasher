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
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

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
	version            string
	enableColorsStdout = true
)

func parseFlags() {
	flag.StringVar(&path, "image", "", "factory image zip file")
	flag.BoolVar(&debug, "debug", false, "debug logging")
	flag.BoolVar(&parallel, "parallel", false, "enables flashing of multiple devices at once")
	flag.Parse()
}

func main() {
	err := execute()
	if err != nil {
		fmt.Println(color.Red(err))
		os.Exit(1)
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
