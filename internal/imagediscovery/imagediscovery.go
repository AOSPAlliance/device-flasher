package imagediscovery

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func Discover(discoverPath string) (map[string]string, error) {
	factoryImages := map[string]string{}
	discoverInfo, err := os.Stat(discoverPath)
	if err != nil {
		return nil, err
	}
	if discoverInfo.IsDir() {
		discoverDir, err := ioutil.ReadDir(discoverPath)
		if err != nil {
			return nil, err
		}
		for _, file := range discoverDir {
			err = validate(file)
			if err != nil {
				continue
			}
			codename, err := getCodename(file)
			if err != nil {
				continue
			}
			if existing, ok := factoryImages[codename]; ok {
				return nil, fmt.Errorf("duplicate factory image (%v) for codename=%v found: %v", existing, codename, discoverPath)
			}
			factoryImages[codename] = discoverPath + string(os.PathSeparator) + file.Name()
		}
	} else {
		err = validate(discoverInfo)
		if err != nil {
			return nil, err
		}
		codename, err := getCodename(discoverInfo)
		if err != nil {
			return nil, err
		}
		factoryImages[codename] = discoverPath
	}

	return factoryImages, nil
}

func getCodename(info os.FileInfo) (string, error) {
	codename := strings.Split(info.Name(), "-")
	if len(codename) <= 0 {
		return "", errors.New("unable to parse codename")
	}
	return codename[0], nil
}

func validate(info os.FileInfo) error {
	if !strings.Contains(info.Name(), "factory") {
		return errors.New("missing factory in filename")
	}
	return nil
}
