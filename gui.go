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
	"github.com/aospalliance/device-flasher/internal/udev"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"image/color"
	"os"
	"strconv"
	"strings"
)

var application = app.New()
var window = application.NewWindow("Android Factory Image Flasher")

func init() {
	application.Settings().SetTheme(theme.LightTheme())
	window.SetMaster()
	window.CenterOnScreen()
	window.Resize(fyne.Size{
		Width:  1280,
		Height: 720,
	})
	window.SetCloseIntercept(func() {
		cleanup()
		os.Exit(0)
	})
}

func setupUdev(logger *logrus.Logger) error {
	// setup udev if running linux
	if hostOS == "linux" {
		//FIXME
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
						&canvas.Text{Color: color.Black, Text: "Android Factory Image Flasher", Alignment: fyne.TextAlignCenter, TextSize: 24},
						&canvas.Text{Color: color.Black, Text: version, TextSize: 24},
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
	step := 1
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
		preparation(logger, step+1)
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
				canvas.NewText(strconv.Itoa(step)+"/7", color.Black),
				layout.NewSpacer(),
				nextButton,
			),
			nil,
			nil,
			container.NewCenter(
				container.NewVBox(
					&canvas.Text{Color: color.Black, Text: "Select Image", Alignment: fyne.TextAlignCenter, TextSize: 24},
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

func preparation(logger *logrus.Logger, step int) {
	text := &canvas.Text{}
	screenshot := &canvas.Image{}
	nextButtonLabel := "Next"
	switch step {
	case 1:
		selection(logger)
		return
	case 2:
		text = &canvas.Text{Color: color.Black, Text: "Connect to a wifi network and ensure that no SIM cards are installed", TextSize: 14}
		screenshot = &canvas.Image{Resource: resourceWifisimPng, FillMode: canvas.ImageFillContain}
	case 3:
		text = &canvas.Text{Color: color.Black, Text: "Enable Developer Options (Settings -> About Phone -> tap \"Build number\" 7 times)", TextSize: 14}
		screenshot = &canvas.Image{Resource: resourceBuildnumberPng, FillMode: canvas.ImageFillContain}
	case 4:
		text = &canvas.Text{Color: color.Black, Text: "Enable OEM Unlocking (Settings -> System -> Advanced -> Developer Options)", TextSize: 14}
		screenshot = &canvas.Image{Resource: resourceOemunlockingPng, FillMode: canvas.ImageFillContain}
	case 5:
		text = &canvas.Text{Color: color.Black, Text: "Enable USB debugging (Settings -> System -> Advanced -> Developer Options)", TextSize: 14}
		screenshot = &canvas.Image{Resource: resourceEnableusbdebuggingPng, FillMode: canvas.ImageFillContain}
	case 6:
		text = &canvas.Text{Color: color.Black, Text: "Plug the device into the computer and allow USB debugging", TextSize: 14}
		screenshot = &canvas.Image{Resource: resourceAllowusbdebuggingPng, FillMode: canvas.ImageFillContain}
		nextButtonLabel = "Flash"
	case 7:
		loading := dialog.NewProgressInfinite("Loading", "Setting up platform tools...", window)
		err := setupPlatformTools(logger)
		loading.Hide()
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		flashing(logger)
		return
	}
	screenshot.SetMinSize(fyne.Size{
		Width:  480,
		Height: 0,
	})
	window.SetContent(
		container.NewBorder(
			nil,
			container.NewHBox(
				widget.NewButton("Back", func() {
					go func() { preparation(logger, step-1) }()
				}),
				layout.NewSpacer(),
				canvas.NewText(strconv.Itoa(step)+"/7", color.Black),
				layout.NewSpacer(),
				widget.NewButton(nextButtonLabel, func() {
					go func() { preparation(logger, step+1) }()
				}),
			),
			nil,
			nil,
			container.NewGridWithColumns(2,
				container.NewCenter(
					text,
				),
				screenshot,
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
		window.RequestFocus()
		codename := line[strings.Index(line, "codename=")+len("codename=") : strings.Index(line, ":")]
		line = strings.Split(line, ": ")[1]
		screenshot := &canvas.Image{}
		if strings.Contains(line, " lock ") {
			screenshot = &canvas.Image{Resource: resourceBootloaderlockingJpg, FillMode: canvas.ImageFillContain}
		} else {
			screenshot = &canvas.Image{Resource: resourceBootloaderunlockingJpg, FillMode: canvas.ImageFillContain}
		}
		screenshot.SetMinSize(fyne.Size{
			Width:  360,
			Height: 0,
		})
		if codename == "walleye" || codename == "jasmine_sprout" {
			line += "Note: your device will reboot\n\nAdditional steps:\n1. Disconnect the cable and power the device off\n2. Press and hold the volume down and power keys to boot the device into fastboot mode\n3. Reconnect the cable"
		}
		bootloaderWarning := dialog.NewCustom("Warning", "OK", container.NewBorder(nil, nil, nil, screenshot, container.NewCenter(widget.NewLabel(line))), window)
		bootloaderWarning.Resize(window.Content().Size())
		bootloaderWarning.Show()
	}
	return len(p), nil
}

func flashing(logger *logrus.Logger) {
	step := 7
	enableColorsStdout = false
	colorable.EnableColorsStdout(&enableColorsStdout)
	logger.SetFormatter(&prefixed.TextFormatter{ForceFormatting: true})
	textGrid := widget.NewTextGrid()
	scroll := container.NewVScroll(textGrid)
	progressBar := widget.NewProgressBar()
	log := application.NewWindow("Log")
	log.SetCloseIntercept(func() {
		log.Hide()
	})
	log.SetContent(
		container.NewBorder(
			nil,
			widget.NewButton("Save Log", func() {
				d := dialog.NewFileSave(func(file fyne.URIWriteCloser, err error) {
					if file != nil {
						_, err = file.Write([]byte(textGrid.Text()))
						if err != nil {
							dialog.ShowError(err, log)
						}
					}
				}, log)
				d.Resize(log.Content().Size())
				wd, _ := os.Getwd()
				lister, _ := storage.ListerForURI(storage.NewFileURI(wd))
				d.SetLocation(lister)
				d.SetFilter(storage.NewExtensionFileFilter([]string{".log"}))
				d.Show()
			}),
			nil,
			nil,
			scroll,
		),
	)
	log.Resize(window.Content().Size())
	installColumn := container.NewVBox(
		&canvas.Text{Color: color.Black, Text: "Installing...", Alignment: fyne.TextAlignCenter, TextSize: 24},
		progressBar,
		widget.NewButton("View Log", func() {
			log.Show()
		}),
	)
	installPadding := layout.NewSpacer()
	padding := layout.NewSpacer()
	infoColumn := container.NewVBox(
		&canvas.Text{Color: color.Black, Text: "Need help?", Alignment: fyne.TextAlignCenter, TextSize: 24},
	)
	footer := container.NewHBox(
		widget.NewButton("Back", func() {
			preparation(logger, step-1)
		}),
		layout.NewSpacer(),
		canvas.NewText(strconv.Itoa(step)+"/7", color.Black),
		layout.NewSpacer(),
	)
	flashButton := widget.NewButton("Retry", func() {
		go func() {
			textGrid.SetText("")
			progressBar.SetValue(0)
			installColumn.Show()
			installPadding.Show()
			footer.Hide()
			err := deviceDiscovery(logger)
			if err != nil {
				dialog.ShowError(err, window)
				installColumn.Hide()
				installPadding.Hide()
				footer.Show()
				return
			}
			err = factoryImageExtraction(logger)
			if err != nil {
				dialog.ShowError(err, window)
				installColumn.Hide()
				installPadding.Hide()
				footer.Show()
				return
			}
			err = flashDevices(logger)
			if err != nil {
				dialog.ShowError(err, window)
				installColumn.Hide()
				installPadding.Hide()
				footer.Show()
				return
			}
			success()
		}()
	})
	footer.Add(flashButton)
	logger.SetOutput(&scrollableTextGridWriter{textGrid, scroll, progressBar})
	flashButton.Tapped(new(fyne.PointEvent))
	window.SetContent(
		container.NewBorder(
			nil,
			footer,
			nil,
			nil,
			container.NewVBox(
				layout.NewSpacer(),
				container.NewHBox(
					installPadding,
					installColumn,
					padding,
					infoColumn,
					padding,
				),
				layout.NewSpacer(),
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
	screenshot := &canvas.Image{Resource: resourceOemlockingPng, FillMode: canvas.ImageFillContain}
	screenshot.SetMinSize(fyne.Size{
		Width:  480,
		Height: 0,
	})
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
			nil,
			container.NewGridWithColumns(2,
				container.NewCenter(
					container.NewVBox(
						&canvas.Text{Color: color.Black, Text: "Re-enable OEM lock", Alignment: fyne.TextAlignCenter, TextSize: 24},
						&canvas.Text{Color: color.Black, Text: "Disable OEM Unlocking (Settings -> System -> Advanced -> Developer Options)", Alignment: fyne.TextAlignCenter, TextSize: 14},
					),
				),
				screenshot,
			),
		),
	)
}
