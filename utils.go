package main

import (
	"fmt"
	"github.com/aospalliance/device-flasher/internal/platformtools"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func getToolsVersion(path string) platformtools.SupportedVersion {
	// TODO: jasmine specific hack
	toolsVersion := platformtools.Version_30_0_4
	if strings.Contains(path, "jasmine") {
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
