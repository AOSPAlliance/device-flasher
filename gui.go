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
		cleanup()
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
			return
		}
		err = imageDiscovery(logger)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		err = setupUdev(logger)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		preparation(logger)
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
					&canvas.Text{Color: color.White, Text: "5. Plug the device into a computer and press \"Next\". Whenever prompted, press \"OK\" on the device to allow USB debugging", Alignment: fyne.TextAlignCenter, TextSize: 14},
					layout.NewSpacer(),
				),
				&canvas.Image{Resource: resourceAndroidPng, FillMode: canvas.ImageFillOriginal},
			),
			container.NewHBox(
				layout.NewSpacer(),
				widget.NewButton("Back", func() {
					selection(logger)
				}),
				widget.NewButton("Next", func() {
					err := setupPlatformTools(logger)
					if err != nil {
						dialog.ShowError(err, window)
						return
					}
					flashing(logger)
				}),
			),
		),
	)
}

type scrollableTextGridWriter struct {
	*container.Scroll
	*widget.TextGrid
}

func (textGridWriter *scrollableTextGridWriter) Write(p []byte) (n int, err error) {
	cells := make([]widget.TextGridCell, len(p))
	for j, r := range string(p[:]) {
		cells[j] = widget.TextGridCell{Rune: r}
	}
	textGridWriter.TextGrid.Rows = append(textGridWriter.Rows, widget.TextGridRow{Cells: cells})
	textGridWriter.TextGrid.Refresh()
	textGridWriter.Scroll.ScrollToBottom()
	return len(p), nil
}

func flashing(logger *logrus.Logger) {
	textGrid := widget.NewTextGrid()
	scroll := container.NewVScroll(textGrid)
	logger.SetOutput(&scrollableTextGridWriter{scroll, textGrid})
	window.SetContent(
		container.NewMax(
			scroll,
			container.NewVBox(
				layout.NewSpacer(),
				container.NewHBox(
					layout.NewSpacer(),
					widget.NewButton("Back", func() {
						preparation(logger)
					}),
					widget.NewButton("Flash", func() {
						err := deviceDiscovery(logger)
						if err != nil {
							dialog.ShowError(err, window)
							return
						}
						err = factoryImageExtraction(logger)
						if err != nil {
							dialog.NewError(err, window)
							return
						}
						err = flashDevices(logger)
						if err != nil {
							dialog.NewError(err, window)
							return
						}
						success()
					}),
				),
			),
		),
	)
}

func success() {
	window.SetContent(
		container.NewCenter(
			&canvas.Text{Color: color.White, Text: "Flashing Complete", Alignment: fyne.TextAlignCenter, TextSize: 24},
		),
	)
}
