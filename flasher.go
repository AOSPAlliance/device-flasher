// Copyright 2020 CIS Maxwell, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var executable, _ = os.Executable()
var cwd = filepath.Dir(executable)

const OS = runtime.GOOS
const PLATFORM_TOOLS_ZIP = "platform-tools-latest-" + OS + ".zip"

var adb *exec.Cmd
var fastboot *exec.Cmd

var input string

var altosImage string
var altosPath string
var device string
var devices []string

var (
	Warn  = Yellow
	Error = Red
)

var (
	Red    = Color("\033[1;31m%s\033[0m")
	Yellow = Color("\033[1;33m%s\033[0m")
)

func Color(color string) func(...interface{}) string {
	return func(args ...interface{}) string {
		return fmt.Sprintf(color,
			fmt.Sprint(args...))
	}
}

func fatalln(err error) {
	log, _ := os.OpenFile("error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	_, _ = fmt.Fprintln(os.Stderr, Error(err.Error()))
	_, _ = fmt.Fprintln(log, err.Error())
	log.Close()
	os.Exit(1)
}

func errorln(err string) {
	log, _ := os.OpenFile("error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	defer log.Close()
	_, _ = fmt.Fprintln(log, err)
	_, _ = fmt.Fprintln(os.Stderr, Error(err))
}

func main() {
	_ = os.Remove("error.log")
	err := getPlatformTools()
	if err != nil {
		errorln("Cannot continue without Android platform tools. Exiting...")
		fatalln(err)
	}
	if OS == "linux" {
		checkUdevRules()
	}
	fmt.Println("Do the following for each device:")
	fmt.Println("Connect to a wifi network and ensure that no SIM cards are installed")
	fmt.Println("Enable Developer Options on device (Settings -> About Phone -> tap \"Build number\" 7 times)")
	fmt.Println("Enable USB debugging on device (Settings -> System -> Advanced -> Developer Options) and allow the computer to debug (hit \"OK\" on the popup when USB is connected)")
	fmt.Println("Enable OEM Unlocking (in the same Developer Options menu)")
	fmt.Print("When done, press enter to continue")
	_, _ = fmt.Scanln(&input)
	getDevices(*adb)
	if len(devices) == 0 {
		getDevices(*fastboot)
		if len(devices) == 0 {
			fatalln(errors.New("No device connected. Exiting..."))
		}
	}
	fmt.Println("Detected " + strconv.Itoa(len(devices)) + " devices: " + strings.Join(devices, ", "))
	device = getProp("ro.product.device", devices[0])
	if device == "" {
		device = getVar("product", devices[0])
		if device == "" {
			fatalln(errors.New("Cannot determine device model. Exiting..."))
		}
	}
	checkPrerequisiteFiles()
	prepareFactoryImage()
	flashDevices()
}

func checkPrerequisiteFiles() {
	files, err := ioutil.ReadDir(cwd)
	if err != nil {
		fatalln(err)
	}
	for _, file := range files {
		file := file.Name()
		if strings.Contains(file, strings.ToLower(device)) && strings.HasSuffix(file, ".zip") {
			if strings.Contains(file, "factory") {
				altosImage = file
			}
		}
	}
	if altosImage == "" {
		fatalln(errors.New("Cannot continue without altOS factory image. Exiting..."))
	}
}

func prepareFactoryImage() {
	err := errors.New("")
	if altosImage != "" {
		fmt.Println("Preparing altOS factory image for flashing")
		err = extractZip(path.Base(altosImage), cwd)
		if err != nil {
			errorln("Cannot continue without the device altOS factory image. Exiting...")
			fatalln(err)
		}
	}
	files, err := ioutil.ReadDir(cwd)
	if err != nil {
		fatalln(err)
	}
	for _, file := range files {
		if file.IsDir() && strings.HasPrefix(file.Name(), strings.ToLower(device)) {
			_, err := os.Stat(cwd + file.Name() + string(os.PathSeparator) + "flash-all.sh")
			if err != nil {
				altosPath = file.Name()
			}
		}
	}
}

func getPlatformTools() error {
	platformToolsPath := cwd + string(os.PathSeparator) + "platform-tools" + string(os.PathSeparator)
	adbPath := platformToolsPath + "adb"
	fastbootPath := platformToolsPath + "fastboot"
	if OS == "windows" {
		adbPath += ".exe"
		fastbootPath += ".exe"
		path := os.Getenv("Path")
		newpath := platformToolsPath + string(os.PathListSeparator) + path
		os.Setenv("Path", newpath)
	} else {
		path := os.Getenv("PATH")
		newpath := platformToolsPath + string(os.PathListSeparator) + path
		os.Setenv("PATH", newpath)
	}
	adb = exec.Command(adbPath)
	fastboot = exec.Command(fastbootPath)
	_, err := os.Stat(platformToolsPath)
	if err == nil {
		killAdb()
	}
	err = extractZip(PLATFORM_TOOLS_ZIP, cwd)
	if err != nil {
		fmt.Println("There are missing Android platform tools in PATH. Attempting to download https://dl.google.com/android/repository/" + PLATFORM_TOOLS_ZIP)
		err = downloadFile("https://dl.google.com/android/repository/" + PLATFORM_TOOLS_ZIP)
		if err != nil {
			return err
		}
		err = extractZip(PLATFORM_TOOLS_ZIP, cwd)
	}
	return err
}

func checkUdevRules() {
	_, err := os.Stat("/etc/udev/rules.d/")
	if os.IsNotExist(err) {
		err = exec.Command("sudo", "mkdir", "/etc/udev/rules.d/").Run()
		if err != nil {
			errorln("Cannot continue without udev rules. Exiting...")
			fatalln(err)
		}
		_, err = os.Stat("99-android.rules")
		if os.IsNotExist(err) {
			err = downloadFile("https://raw.githubusercontent.com/invisiblek/udevrules/master/99-android.rules")
			if err != nil {
				errorln("Cannot continue without udev rules. Exiting...")
				fatalln(err)
			}
		}
		err = exec.Command("sudo", "cp", "99-android.rules", "/etc/udev/rules.d/").Run()
		if err != nil {
			errorln("Cannot continue without udev rules. Exiting...")
			fatalln(err)
		}
		_ = exec.Command("sudo", "udevadm", "control", "--reload-rules").Run()
		_ = exec.Command("sudo", "udevadm", "trigger").Run()
	}
}

func killAdb() {
	platformToolCommand := *adb
	platformToolCommand.Args = append(platformToolCommand.Args, "kill-server")
	err := platformToolCommand.Run()
	if err != nil {
		errorln(err.Error())
	}
}

func getDevices(platformToolCommand exec.Cmd) {
	platformToolCommand.Args = append(adb.Args, "devices")
	output, _ := platformToolCommand.Output()
	devices = strings.Split(string(output), "\n")
	if platformToolCommand.Path == adb.Path {
		devices = devices[1 : len(devices)-2]
	} else if platformToolCommand.Path == fastboot.Path {
		devices = devices[:len(devices)-1]
	}
	for i, device := range devices {
		devices[i] = strings.Split(device, "\t")[0]
	}
}

func getProp(prop string, device string) string {
	platformToolCommand := *adb
	platformToolCommand.Args = append(adb.Args, "-s", device, "shell", "getprop", prop)
	out, err := platformToolCommand.Output()
	if err != nil {
		return ""
	}
	return strings.Trim(string(out), "[]\n\r")
}

func getVar(prop string, device string) string {
	platformToolCommand := *fastboot
	platformToolCommand.Args = append(adb.Args, "-s", device, "getvar", prop)
	out, err := platformToolCommand.CombinedOutput()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, prop) {
			return strings.Trim(strings.Split(line, " ")[1], "\r")
		}
	}
	return ""
}

func flashDevices() {
	var wg sync.WaitGroup
	for _, device := range devices {
		wg.Add(1)
		go func(device string) {
			defer wg.Done()
			err := errors.New("")
			platformToolCommand := *adb
			platformToolCommand.Args = append(platformToolCommand.Args, "-s", device, "reboot", "bootloader")
			_ = platformToolCommand.Run()
			fmt.Println("Unlocking device " + device + " bootloader...")
			fmt.Println("Please use the volume and power keys on the device to confirm.")
			platformToolCommand = *fastboot
			platformToolCommand.Args = append(platformToolCommand.Args, "-s", device, "flashing", "unlock")
			_ = platformToolCommand.Start()
			time.Sleep(5 * time.Second)
			if getVar("unlocked", device) != "yes" {
				errorln("Failed to unlock device " + device + " bootloader")
				return
			}
			os.Chdir(altosPath)
			flashCmd := exec.Command("")
			if OS == "windows" {
				flashCmd = exec.Command("./flash-all.bat")
			} else {
				flashCmd = exec.Command("./flash-all.sh")
			}
			fmt.Println("Flashing altOS on device " + device + "...")
			output, err := flashCmd.Output()
			fmt.Printf("%s\n", output)
			if err != nil {
				errorln("Failed to flash altOS for device " + device)
				return
			}
			fmt.Println("Locking device " + device + " bootloader...")
			fmt.Println("Please use the volume and power keys on the device to confirm.")
			platformToolCommand = *fastboot
			platformToolCommand.Args = append(platformToolCommand.Args, "-s", device, "flashing", "lock")
			_ = platformToolCommand.Start()
			time.Sleep(5 * time.Second)
			if getVar("unlocked", device) != "no" {
				errorln("Failed to lock device " + device + " bootloader")
				return
			}
			fmt.Println("Rebooting " + device + "...")
			platformToolCommand = *fastboot
			platformToolCommand.Args = append(platformToolCommand.Args, "-s", device, "reboot")
			_ = platformToolCommand.Start()
		}(device)
	}
	wg.Wait()
	fmt.Println("Bulk flashing complete")
}

func extractZip(src, dest string) error {
	dest = filepath.Clean(dest) + string(os.PathSeparator)

	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	extractAndWriteFile := func(f *zip.File) error {
		path := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(path, dest) {
			return fmt.Errorf("%s: illegal file path", path)
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
		} else {
			os.MkdirAll(filepath.Dir(path), 0755)
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %s downloaded", Bytes(wc.Total))
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func humanateBytes(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := math.Floor(float64(s)/math.Pow(base, e)*10+0.5) / 10
	f := "%.0f %s"
	if val < 10 {
		f = "%.1f %s"
	}

	return fmt.Sprintf(f, val, suffix)
}

func Bytes(s uint64) string {
	sizes := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}
	return humanateBytes(s, 1000, sizes)
}

func downloadFile(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(path.Base(url))
	if err != nil {
		return err
	}
	defer out.Close()

	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	fmt.Println()
	return err
}
