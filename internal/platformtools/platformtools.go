package platformtools

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/mholt/archiver/v3"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
)

const (
	DefaultBaseURI = "https://dl.google.com/android/repository"
)

type PlatformToolsPath string

type Config struct {
	CacheDir             string
	BaseURI              string
	HttpClient           *http.Client
	HostOS               string
	ToolsVersion         SupportedVersion
	DestinationDirectory string
	Logger               *logrus.Logger
}

type PlatformTools struct {
	cacheDir         string
	httpClient       *http.Client
	hostOS           string
	downloadURI      string
	sha256           string
	workingDirectory string
	zipFile          string
	path             string
	logger           *logrus.Logger
}

func New(config *Config) (*PlatformTools, error) {
	workingDirectory := config.DestinationDirectory
	cacheDir := config.CacheDir
	zipFile := fmt.Sprintf("%v%v%v", cacheDir, string(os.PathSeparator), "platform-tools.zip")
	path := fmt.Sprintf("%v%vplatform-tools", workingDirectory, string(os.PathSeparator))
	download := Downloads[config.ToolsVersion][SupportedHostOS(config.HostOS)]

	platformTools := &PlatformTools{
		cacheDir:         cacheDir,
		path:             path,
		httpClient:       config.HttpClient,
		downloadURI:      fmt.Sprintf(download.TemplateURL, DefaultBaseURI),
		sha256:           download.CheckSum,
		workingDirectory: workingDirectory,
		zipFile:          zipFile,
		logger:           config.Logger,
	}

	err := platformTools.initialize()
	if err != nil {
		return nil, err
	}

	return platformTools, nil
}

func (p *PlatformTools) initialize() error {
	logger := p.logger.WithFields(logrus.Fields{
		"cacheDir":    p.cacheDir,
		"downloadURI": p.downloadURI,
		"sha256":      p.sha256,
		"zipFile":     p.zipFile,
	})

	_, err := os.Stat(p.zipFile)
	if err != nil {
		logger.Debug("starting tools download")
		err := p.download()
		if err != nil {
			return err
		}
	}

	logger.Debug("starting tools verify")
	err = p.verify()
	if err != nil {
		os.Remove(p.zipFile)
		return err
	}

	logger.Debug("starting tools extract")
	err = p.extract()
	if err != nil {
		return err
	}

	return nil
}

func (p *PlatformTools) Path() PlatformToolsPath {
	return PlatformToolsPath(p.path)
}

func (p *PlatformTools) download() error {
	p.logger.Debugf("making directory %v", p.workingDirectory)
	_ = os.Mkdir(p.workingDirectory, os.ModePerm)

	p.logger.Debugf("creating file %v", p.zipFile)
	out, err := os.Create(p.zipFile)
	if err != nil {
		return err
	}
	defer out.Close()

	p.logger.Infof("downloading %v", p.downloadURI)
	resp, err := p.httpClient.Get(p.downloadURI)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad download status from %v: %v", p.zipFile, resp.Status)
	}

	p.logger.Debugf("copying data to %v", p.zipFile)
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (p *PlatformTools) extract() error {
	p.logger.Debugf("unzipping %v to directory %v", p.zipFile, p.workingDirectory)
	return archiver.Unarchive(p.zipFile, p.workingDirectory)
}

func (p *PlatformTools) verify() error {
	p.logger.Debugf("verifying sha256 for %v = %v", p.zipFile, p.sha256)
	f, err := os.Open(p.zipFile)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	sum := hex.EncodeToString(h.Sum(nil))
	if p.sha256 != sum {
		return fmt.Errorf("expected sha256 to be %v but got %v", p.sha256, sum)
	}
	return nil
}
