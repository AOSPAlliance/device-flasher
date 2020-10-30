package imagediscovery

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Discover(discoverPath string) (map[string]string, error) {
	factoryImages := map[string]string{}
	discoverInfo, err := os.Stat(discoverPath)
	if err != nil {
		return nil, err
	}
	if discoverInfo.IsDir() {
		err := filepath.Walk(discoverPath, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			err = validate(info)
			if err != nil {
				return nil
			}
			codename, err := getCodename(info)
			if err != nil {
				return nil
			}
			if existing, ok := factoryImages[codename]; ok {
				return fmt.Errorf("duplicate factory image (%v) for codename=%v found: %v", existing, codename, path)
			}
			factoryImages[codename] = path
			return nil
		})
		if err != nil {
			return nil, err
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
	//TODO re-enable when we drop support for jasmine
	//if !strings.Contains(info.Name(), "factory") {
	//	return errors.New("missing factory in filename")
	//}
	return nil
}
