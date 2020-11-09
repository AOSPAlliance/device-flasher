// +build GUI

package main

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/storage"
	"fyne.io/fyne/widget"
	"github.com/sirupsen/logrus"
	"gitlab.com/calyxos/device-flasher/internal/udev"
	"image/color"
	"os"
	"strings"
)

var window = app.New().NewWindow("Android Factory Image Flasher")

func init() {
	window.SetCloseIntercept(func() {
		os.Exit(0)
	})
}

func setupUdev(logger *logrus.Logger) error {
	// setup udev if running linux
	if hostOS == "linux" {
		err := udev.Setup(logger, "gksudo", udev.DefaultUDevRules)
		if err != nil {
			return fmt.Errorf("failed to setup udev: %v", err)
		}
		cleanupPaths = append(cleanupPaths, udev.TempRulesFile)
	}
	return nil
}

func execute(logger *logrus.Logger) error {
	gui(logger)
	window.ShowAndRun()
	return nil
}

func gui(logger *logrus.Logger) {
	window.SetContent(
		container.NewVBox(
			container.NewGridWithColumns(2,
				&canvas.Text{Color: color.White, Text: "Android Factory Image Flasher", Alignment: fyne.TextAlignCenter, TextSize: 18},
				&canvas.Image{Resource: resourceAndroidPng, FillMode: canvas.ImageFillOriginal},
			),
			container.NewHBox(
				layout.NewSpacer(),
				widget.NewButton("Next", func() {
					selection(logger)
				}),
			),
		),
	)
}

func selection(logger *logrus.Logger) {
	selectedFile := canvas.NewText("", color.White)
	nextButton := widget.NewButton("Next", func() {
		err := pathValidation()
		if err != nil {
			dialog.ShowError(err, window)
		}
		err = imageDiscovery(logger)
		if err != nil {
			dialog.ShowError(err, window)
		}
		err = setupUdev(logger)
		if err != nil {
			dialog.ShowError(err, window)
		}
		preparation(logger)
		err = setupPlatformTools(logger)
		if err != nil {
			dialog.ShowError(err, window)
		}
		err = deviceDiscovery(logger)
		if err != nil {
			dialog.ShowError(err, window)
		}
	})
	nextButton.Disable()
	window.SetContent(
		container.NewVBox(
			container.NewHBox(
				widget.NewButton("Select", func() {
					d := dialog.NewFileOpen(
						func(file fyne.URIReadCloser, err error) {
							if file != nil {
								path = strings.ReplaceAll(file.URI().String(), "file://", "")
								selectedFile.Text = path
								nextButton.Enable()
							}
						}, window)
					wd, _ := os.Getwd()
					lister, _ := storage.ListerForURI(storage.NewFileURI(wd))
					d.SetLocation(lister)
					//TODO add other archive file extensions
					d.SetFilter(storage.NewExtensionFileFilter([]string{".zip", ".tar.xz", ".tgz"}))
					d.Show()
				}),
				selectedFile,
			),
			layout.NewSpacer(),
			nextButton,
		),
	)
}

func preparation(logger *logrus.Logger) {
	window.SetContent(
		container.NewVBox(
			container.NewGridWithColumns(2,
				container.NewVBox(
					layout.NewSpacer(),
					&canvas.Text{Color: color.White, Text: "Prepare your device", Alignment: fyne.TextAlignCenter, TextSize: 18},
					&canvas.Text{Color: color.White, Text: "1. Connect to a wifi network and ensure that no SIM cards are installed", Alignment: fyne.TextAlignCenter, TextSize: 14},
					&canvas.Text{Color: color.White, Text: "2. Enable Developer Options (Settings -> About Phone -> tap \"Build number\" 7 times)", Alignment: fyne.TextAlignCenter, TextSize: 14},
					&canvas.Text{Color: color.White, Text: "3. Enable OEM Unlocking (Settings -> System -> Advanced -> Developer Options)", Alignment: fyne.TextAlignCenter, TextSize: 14},
					&canvas.Text{Color: color.White, Text: "4. Enable USB debugging (Settings -> System -> Advanced -> Developer Options)", Alignment: fyne.TextAlignCenter, TextSize: 14},
					&canvas.Text{Color: color.White, Text: "5. Plug the device into a computer and, when prompted, press \"OK\" to allow USB debugging", Alignment: fyne.TextAlignCenter, TextSize: 14},
					layout.NewSpacer(),
				),
				&canvas.Image{Resource: resourceAndroidPng, FillMode: canvas.ImageFillOriginal},
			),
			container.NewHBox(
				layout.NewSpacer(),
				widget.NewButton("Next", func() {
					err := factoryImageExtraction(logger)
					if err != nil {
						dialog.NewError(err, window)
					}
					err = flashDevices(logger)
					if err != nil {
						dialog.NewError(err, window)
					}
				}),
			),
		),
	)
}
