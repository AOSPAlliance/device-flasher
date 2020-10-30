package factoryimage

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/mholt/archiver/v3"
	"github.com/sirupsen/logrus"
	"gitlab.com/calyxos/device-flasher/internal/device"
	"gitlab.com/calyxos/device-flasher/internal/platformtools"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	ErrorFailedToFlash = errors.New("failed to flash device")
	ErrorValidation    = errors.New("failed to validate image for device")
)

type Config struct {
	HostOS           string
	ImagePath        string
	WorkingDirectory string
	Logger           *logrus.Logger
}

type FactoryImage struct {
	hostOS             string
	extractedDirectory string
	flashAllScript     string
	ImagePath          string
	workingDirectory   string
	logger             *logrus.Logger
	IsExtracted        bool
}

func New(config *Config) *FactoryImage {
	return &FactoryImage{
		hostOS:           config.HostOS,
		workingDirectory: config.WorkingDirectory,
		ImagePath:        config.ImagePath,
		logger:           config.Logger,
	}
}

func (f *FactoryImage) FlashAll(device *device.Device, platformToolsPath platformtools.PlatformToolsPath) error {
	pathEnvironmentVariable := "PATH"
	if f.hostOS == "windows" {
		pathEnvironmentVariable = "Path"
	}
	path := os.Getenv(pathEnvironmentVariable)
	pathWithPlatformTools := string(platformToolsPath) + string(os.PathListSeparator) + path

	flashAll := fmt.Sprintf(".%v%v", string(os.PathSeparator), f.flashAllScript)
	f.logger.WithFields(logrus.Fields{
		"flashAll": flashAll,
	}).Debug("running flash all script on device")
	flashCmd := exec.Command(flashAll)
	flashCmd.Dir = f.extractedDirectory
	flashCmd.Env = os.Environ()
	flashCmd.Env = append(flashCmd.Env, fmt.Sprintf("%v=%v", pathEnvironmentVariable, pathWithPlatformTools))
	flashCmd.Env = append(flashCmd.Env, fmt.Sprintf("ANDROID_SERIAL=%v", device.ID))

	cmdStdoutReader, err := flashCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrorFailedToFlash, err)
	}
	cmdStderrReader, err := flashCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrorFailedToFlash, err)
	}

	scannerStdout := bufio.NewScanner(cmdStdoutReader)
	go func() {
		for scannerStdout.Scan() {
			f.logger.Infof("%v | %s", device.String(), scannerStdout.Text())
		}
	}()
	scannerStderr := bufio.NewScanner(cmdStderrReader)
	go func() {
		for scannerStderr.Scan() {
			f.logger.Infof("%v | %s", device.String(), scannerStderr.Text())
		}
	}()

	err = flashCmd.Start()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrorFailedToFlash, err)
	}

	err = flashCmd.Wait()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrorFailedToFlash, err)
	}

	return nil
}

func (f *FactoryImage) Validate(deviceCodename device.Codename) error {
	f.logger.WithFields(logrus.Fields{
		"deviceCodename": deviceCodename,
	}).Debug("running factory image validation")
	if _, err := os.Stat(f.ImagePath); os.IsNotExist(err) {
		return fmt.Errorf("%w: %v", ErrorValidation, err)
	}
	if !strings.Contains(f.ImagePath, strings.ToLower(string(deviceCodename))) {
		return fmt.Errorf("%w: image filename should contain device codename %v", ErrorValidation, deviceCodename)
	}
	//TODO re-enable when we drop support for jasmine
	//if !strings.Contains(f.ImagePath, "factory") {
	//	return fmt.Errorf("%w: image filename should contain 'factory'", ErrorValidation)
	//}
	return nil
}

func (f *FactoryImage) Extract() error {
	if f.IsExtracted {
		f.logger.Debugf("already extracted %v to %v", f.ImagePath, f.workingDirectory)
		return nil
	}

	err := f.extract()
	if err != nil {
		return err
	}

	err = f.postExtractValidations()
	if err != nil {
		return err
	}

	f.IsExtracted = true
	return nil
}

func (f *FactoryImage) extract() error {
	f.logger.Infof("Extracting factory image: %v", f.ImagePath)
	err := archiver.Unarchive(f.ImagePath, f.workingDirectory)
	if err != nil {
		return err
	}
	return nil
}

func (f *FactoryImage) postExtractValidations() error {
	f.flashAllScript = "flash"
	if filepath.Base(f.ImagePath) == "jasmine_global_images_V9.6.17.0.ODIMIFE_20181108.0000.00_8.1_1c60295d1c.tgz" {
		f.flashAllScript += "_"
	} else {
		f.flashAllScript += "-"
	}
	if f.hostOS == "windows" {
		f.flashAllScript += "all.bat"
	} else {
		f.flashAllScript += "all.sh"
	}

	// TODO this can probably be simplified
	files, err := ioutil.ReadDir(f.workingDirectory)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			_, err := os.Stat(f.workingDirectory + file.Name() + string(os.PathSeparator) + f.flashAllScript)
			if err != nil {
				f.extractedDirectory = f.workingDirectory + string(os.PathSeparator) + file.Name()
			}
		}
	}
	if f.extractedDirectory == "" {
		return fmt.Errorf("unable to find %v in directory %v", f.flashAllScript, f.workingDirectory)
	}

	f.logger.WithFields(logrus.Fields{
		"flashAllScript":     f.flashAllScript,
		"extractedDirectory": f.extractedDirectory,
	}).Debug("validated extracted factory image")

	return nil
}
