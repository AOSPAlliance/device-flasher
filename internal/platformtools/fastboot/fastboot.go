package fastboot

import (
	"errors"
	"fmt"
	"gitlab.com/calyxos/device-flasher/internal/platformtools"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	fastbootExecutable = platformtools.Fastboot
)

var (
	ErrorCommandFailure    = errors.New("failed running command")
	ErrorUnlockBootloader  = errors.New("failed to unlock bootloader")
	ErrorLockBootloader    = errors.New("failed to lock bootloader")
	ErrorRebootFailure     = errors.New("failed to reboot")
	ErrorUnknownLockStatus = errors.New("unknown unlocked value returned")
)

type FastbootLockStatus int

const (
	Unknown FastbootLockStatus = iota
	Unlocked
	Locked
)

type Tool struct {
	executable string
	hostOS     string
}

func New(path platformtools.PlatformToolsPath, hostOS string) (*Tool, error) {
	executable := filepath.Join(string(path), string(fastbootExecutable))
	if hostOS == "windows" {
		executable = executable + ".exe"
	}
	if _, err := os.Stat(executable); os.IsNotExist(err) {
		return nil, err
	}
	return &Tool{
		executable: fmt.Sprintf("%v/%v", path, fastbootExecutable),
	}, nil
}

func (t *Tool) GetDeviceIds() ([]string, error) {
	resp, err := t.command([]string{"devices"})
	if err != nil {
		return nil, err
	}
	devices := strings.Split(string(resp), "\n")
	devices = devices[:len(devices)-1]
	for i, device := range devices {
		devices[i] = strings.Split(device, "\t")[0]
	}
	return devices, nil
}

func (t *Tool) GetDeviceCodename(deviceId string) (string, error) {
	return t.getVar("product", deviceId)
}

func (t *Tool) SetBootloaderLockStatus(deviceId string, wantedStatus FastbootLockStatus) error {
	bootloaderCommand := t.lockBootloader
	if wantedStatus == Unlocked {
		bootloaderCommand = t.unlockBootloader
	}
	err := bootloaderCommand(deviceId)
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	lockStatus, err := t.GetBootloaderLockStatus(deviceId)
	if err != nil {
		return err
	}

	if lockStatus != wantedStatus {
		if wantedStatus == Unlocked {
			return fmt.Errorf("%w: %v", ErrorUnlockBootloader, deviceId)
		}
		return fmt.Errorf("%w: %v", ErrorLockBootloader, deviceId)
	}

	return nil
}

func (t *Tool) GetBootloaderLockStatus(deviceId string) (FastbootLockStatus, error) {
	unlocked, err := t.getVar("unlocked", deviceId)
	if err != nil {
		return Unknown, err
	}
	switch unlocked {
	case "yes":
		return Unlocked, nil
	case "no":
		return Locked, nil
	}
	return Unknown, fmt.Errorf("%w: %v", ErrorUnknownLockStatus, unlocked)
}

func (t *Tool) Reboot(deviceId string) error {
	_, err := t.command([]string{"-s", deviceId, "reboot"})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrorRebootFailure, err)
	}
	return nil
}

func (t *Tool) Name() platformtools.ToolName {
	return fastbootExecutable
}

func (t *Tool) command(args []string) ([]byte, error) {
	cmd := exec.Command(t.executable, args...)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrorCommandFailure, err)
	}
	return data, nil
}

func (t *Tool) lockBootloader(deviceId string) error {
	_, err := t.command([]string{"-s", deviceId, "flashing", "lock"})
	if err != nil {
		return err
	}
	return nil
}

func (t *Tool) unlockBootloader(deviceId string) error {
	_, err := t.command([]string{"-s", deviceId, "flashing", "unlock"})
	if err != nil {
		return err
	}
	return nil
}

func (t *Tool) getVar(prop, deviceId string) (string, error) {
	resp, err := t.command([]string{"-s", deviceId, "getvar", prop})
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(resp), "\n")
	for _, line := range lines {
		if strings.Contains(line, prop) {
			return strings.Trim(strings.Split(line, " ")[1], "\r"), nil
		}
	}
	return "", fmt.Errorf("var %v not found", prop)
}
