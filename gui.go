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
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"github.com/aospalliance/device-flasher/internal/udev"
	"image/color"
	"os"
	"strings"
)

var application = app.New()
var window = application.NewWindow("Android Factory Image Flasher")

func init() {
	application.Settings().SetTheme(theme.LightTheme())
	window.CenterOnScreen()
	window.Resize(fyne.Size{
		Width:  1024,
		Height: 768,
	})
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
	defer cleanup()
	gui(logger)
	window.ShowAndRun()
	return nil
}

func gui(logger *logrus.Logger) {
	window.SetContent(
		container.NewBorder(
			nil,
			container.NewHBox(
				layout.NewSpacer(),
				widget.NewButton("Next", func() {
					selection(logger)
				}),
			),
			nil,
			nil,
			container.NewGridWithColumns(2,
				container.NewCenter(
					container.NewVBox(
						canvas.NewText(version, color.Black),
						&canvas.Text{Color: color.Black, Text: "Android Factory Image Flasher", Alignment: fyne.TextAlignCenter, TextSize: 18},
					),
				),
				container.NewCenter(
					&canvas.Image{Resource: resourceAndroidSvg, FillMode: canvas.ImageFillOriginal},
				),
			),
		),
	)
}

func selection(logger *logrus.Logger) {
	selectedFile := canvas.NewText("", color.Black)
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
		container.NewBorder(
			nil,
			container.NewHBox(
				widget.NewButton("Back", func() {
					gui(logger)
				}),
				layout.NewSpacer(),
				canvas.NewText("1/3", color.Black),
				layout.NewSpacer(),
				nextButton,
			),
			nil,
			nil,
			container.NewCenter(
				container.NewVBox(
					&canvas.Text{Color: color.Black, Text: "Select Image", Alignment: fyne.TextAlignCenter, TextSize: 18},
					selectedFile,
					container.NewCenter(
						widget.NewButton("Select", func() {
							d := &dialog.FileDialog{}
							if !parallel {
								d = dialog.NewFileOpen(
									func(file fyne.URIReadCloser, err error) {
										if file != nil {
											path = strings.ReplaceAll(file.URI().String(), "file://", "")
											selectedFile.Text = path
											nextButton.Enable()
										}
									}, window)
							} else {
								d = dialog.NewFolderOpen(
									func(folder fyne.ListableURI, err error) {
										if folder != nil {
											path = strings.ReplaceAll(folder.String(), "file://", "")
											selectedFile.Text = path
											nextButton.Enable()
										}
									}, window)
							}
							wd, _ := os.Getwd()
							lister, _ := storage.ListerForURI(storage.NewFileURI(wd))
							d.SetLocation(lister)
							if !parallel {
								//TODO add other archive file extensions
								d.SetFilter(storage.NewExtensionFileFilter([]string{".zip", ".tar.xz", ".tgz"}))
							}
							d.Resize(window.Content().Size())
							d.Show()
						}),
					),
				),
			),
		),
	)
}

func preparation(logger *logrus.Logger) {
	window.SetContent(
		container.NewBorder(
			nil,
			container.NewHBox(
				widget.NewButton("Back", func() {
					selection(logger)
				}),
				layout.NewSpacer(),
				canvas.NewText("2/3", color.Black),
				layout.NewSpacer(),
				widget.NewButton("Next", func() {
					loading := dialog.NewProgressInfinite("Loading", "Setting up platform tools...", window)
					err := setupPlatformTools(logger)
					loading.Hide()
					if err != nil {
						dialog.ShowError(err, window)
						return
					}
					flashing(logger)
				}),
			),
			nil,
			&canvas.Image{Resource: resourceAndroidSvg, FillMode: canvas.ImageFillOriginal},
			container.NewCenter(
				container.NewVBox(
					&canvas.Text{Color: color.Black, Text: "Prepare your device", Alignment: fyne.TextAlignCenter, TextSize: 18},
					&canvas.Text{Color: color.Black, Text: "1. Connect to a wifi network and ensure that no SIM cards are installed", Alignment: fyne.TextAlignCenter, TextSize: 14},
					&canvas.Text{Color: color.Black, Text: "2. Enable Developer Options (Settings -> About Phone -> tap \"Build number\" 7 times)", Alignment: fyne.TextAlignCenter, TextSize: 14},
					&canvas.Text{Color: color.Black, Text: "3. Enable OEM Unlocking (Settings -> System -> Advanced -> Developer Options)", Alignment: fyne.TextAlignCenter, TextSize: 14},
					&canvas.Text{Color: color.Black, Text: "4. Enable USB debugging (Settings -> System -> Advanced -> Developer Options)", Alignment: fyne.TextAlignCenter, TextSize: 14},
					&canvas.Text{Color: color.Black, Text: "5. Plug the device into the computer and press \"Next\"", Alignment: fyne.TextAlignCenter, TextSize: 14},
					canvas.NewText("", color.Black),
					&canvas.Text{Color: color.Black, Text: "(Whenever prompted on the device, press \"OK\" on the device to allow USB debugging)", Alignment: fyne.TextAlignCenter, TextSize: 14},
				),
			),
		),
	)
}

type scrollableTextGridWriter struct {
	*widget.TextGrid
	*container.Scroll
	*widget.ProgressBar
}

func (textGridWriter *scrollableTextGridWriter) Write(p []byte) (n int, err error) {
	cells := make([]widget.TextGridCell, len(p))
	line := string(p[:])
	for j, rune := range line {
		cells[j] = widget.TextGridCell{Rune: rune}
	}
	textGridWriter.TextGrid.Rows = append(textGridWriter.Rows, widget.TextGridRow{Cells: cells})
	textGridWriter.TextGrid.Refresh()
	textGridWriter.Scroll.ScrollToBottom()
	if flashableDevices != nil {
		textGridWriter.ProgressBar.SetValue(textGridWriter.ProgressBar.Value + (0.0075 / float64(len(flashableDevices))))
	}
	if strings.Contains(line, "Please") {
		codename := line[strings.Index(line, "codename=")+len("codename=") : strings.Index(line, ":")]
		line = strings.Split(line, ": ")[1]
		if codename == "walleye" || codename == "jasmine_sprout" {
			line += "\n Once device boots, disconnect its cable and power it off"
			line += "\n Then, press volume down + power to boot it into fastboot mode, and connect the cable again"
		}
		dialog.ShowInformation("Warning", line, window)
	}
	return len(p), nil
}

func flashing(logger *logrus.Logger) {
	enableColorsStdout = false
	colorable.EnableColorsStdout(&enableColorsStdout)
	logger.SetFormatter(&prefixed.TextFormatter{ForceFormatting: true})
	textGrid := widget.NewTextGrid()
	scroll := container.NewVScroll(textGrid)
	progressBar := widget.NewProgressBar()
	progress := container.NewGridWithColumns(2,
		container.NewCenter(
			container.NewVBox(
				&canvas.Text{Color: color.Black, Text: "Installing...", Alignment: fyne.TextAlignCenter, TextSize: 18},
				progressBar,
			),
		),
		container.NewCenter(
			container.NewVBox(
				&canvas.Text{Color: color.Black, Text: "Did you know?", Alignment: fyne.TextAlignCenter, TextSize: 16},
			),
		),
	)
	footer := container.NewHBox(
		widget.NewButton("Back", func() {
			preparation(logger)
		}),
		layout.NewSpacer(),
		canvas.NewText("3/3", color.Black),
		layout.NewSpacer(),
	)
	flashButton := widget.NewButton("Flash", func() {
		go func() {
			textGrid.SetText("")
			progressBar.SetValue(0)
			progress.Show()
			footer.Hide()
			err := deviceDiscovery(logger)
			if err != nil {
				dialog.ShowError(err, window)
				footer.Show()
				progress.Hide()
				return
			}
			err = factoryImageExtraction(logger)
			if err != nil {
				dialog.ShowError(err, window)
				footer.Show()
				progress.Hide()
				return
			}
			err = flashDevices(logger)
			if err != nil {
				dialog.ShowError(err, window)
				footer.Show()
				progress.Hide()
				return
			}
			success()
		}()
	})
	footer.Add(flashButton)
	scroll.Hide()
	progress.Hide()
	logger.SetOutput(&scrollableTextGridWriter{textGrid, scroll, progressBar})
	window.SetContent(
		container.NewBorder(
			widget.NewButton("Log", func() {
				scroll.Show()
				window.CenterOnScreen()
			}),
			footer,
			nil,
			nil,
			container.NewGridWithRows(2,
				scroll,
				progress,
			),
		),
	)
}

func success() {
	window.SetContent(
		container.NewBorder(
			nil,
			container.NewHBox(
				layout.NewSpacer(),
				widget.NewButton("Next", func() {
					relock()
				}),
			),
			nil,
			nil,
			container.NewCenter(
				container.NewVBox(
					&canvas.Image{Resource: resourceSuccessSvg, FillMode: canvas.ImageFillOriginal},
					&canvas.Text{Color: color.Black, Text: "Flashing Complete", Alignment: fyne.TextAlignCenter, TextSize: 24},
					&canvas.Text{Color: color.Black, Text: "You can now unplug your device", Alignment: fyne.TextAlignCenter, TextSize: 14},
				),
			),
		),
	)
}

func relock() {
	window.SetContent(
		container.NewBorder(
			nil,
			container.NewHBox(
				layout.NewSpacer(),
				widget.NewButton("Finish", func() {
					window.Close()
				}),
			),
			nil,
			&canvas.Image{Resource: resourceAndroidSvg, FillMode: canvas.ImageFillOriginal},
			container.NewCenter(
				container.NewVBox(
					&canvas.Text{Color: color.Black, Text: "Re-enable OEM lock", Alignment: fyne.TextAlignCenter, TextSize: 18},
					&canvas.Text{Color: color.Black, Text: "Disable OEM Unlocking (Settings -> System -> Advanced -> Developer Options)", Alignment: fyne.TextAlignCenter, TextSize: 14},
				),
			),
		),
	)
}
