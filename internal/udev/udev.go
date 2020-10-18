package udev

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
)

const (
	RulesPath = "/etc/udev/rules.d/"
	RulesFile = "98-device-flasher.rules"
)

type UDevRule struct {
	DeviceName   string
	AttrIDVendor string
}

type UDevRules struct {
	Rules []UDevRule
}

var (
	DefaultUDevRules = UDevRules{
		[]UDevRule{
			{
				"Google",
				"18d1",
			},
			{
				"Xiaomi",
				"2717",
			},
		},
	}
)

func (u *UDevRules) Output() ([]byte, error) {
	tmpl, err := template.New("template").Parse(`
{{range $i, $r := .}}
# {{.DeviceName}}
SUBSYSTEM=="usb", ATTR{idVendor}=="{{.AttrIDVendor}}", GROUP="sudo"
{{end}}`)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, u.Rules)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Setup(logger *logrus.Logger, udevRules UDevRules) error {
	_, err := os.Stat(RulesPath + RulesFile)
	if os.IsNotExist(err) {
		logger.Info("setting up udev - this will require elevated privileges and may prompt for password")

		// setup udev rules path
		logger.Infof("udev: creating %v", RulesPath)
		cmd := exec.Command("sudo", "mkdir", "-p", RulesPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to mkdir %v: err=%v output=%v", RulesPath, err, string(output))
		}

		// write udev rules
		udevRulesOutput, err := udevRules.Output()
		if err != nil {
			return err
		}
		logger.Debug("udev: rules=%v", udevRulesOutput)
		err = ioutil.WriteFile(RulesFile, udevRulesOutput, 0644)
		if err != nil {
			return err
		}
		logger.Infof("udev: writing rules to %v/%v", RulesPath, RulesFile)
		cmd = exec.Command("sudo", "cp", RulesFile, RulesPath)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to copy %v to %v: err=%v output=%v", RulesFile, RulesPath, err, string(output))
		}

		// reload udev rules
		logger.Info("udev: reloading rules with udevadm")
		cmd = exec.Command("sudo", "udevadm", "control", "--reload-rules")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to reload udev rules: err=%v output=%v", err, string(output))
		}
	}
	return nil
}
