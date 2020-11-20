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
	"github.com/aospalliance/device-flasher/internal/udev"
	"github.com/aospalliance/device-flasher/resources"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var application = app.New()
var window = application.NewWindow(title + " " + version)

func padding(padding int, objects ...fyne.CanvasObject) *fyne.Container {
	//TODO find better way to pad
	padded := container.NewPadded(objects...)
	for i := 0; i < padding; i++ {
		padded = container.NewPadded(padded)
	}
	return padded
}

func header(text string) *fyne.Container {
	vbox := container.NewVBox()
	for _, line := range strings.Split(text, "\n") {
		vbox.Add(&canvas.Text{Color: fyne.CurrentApp().Settings().Theme().TextColor(), Text: line, TextSize: 32})
	}
	return vbox
}

func body(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	return label
}

func init() {
	window.SetMaster()
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
		//FIXME
		err := udev.Setup(logger, "gksudo", udev.DefaultUDevRules)
		if err != nil {
			return fmt.Errorf("failed to setup udev: %v", err)
		}
		cleanupPaths = append(cleanupPaths, udev.TempRulesFile)
	}
	return nil
}

func execute() error {
	defer cleanup()
	enableColorsStdout = false
	colorable.EnableColorsStdout(&enableColorsStdout)
	logger := logrus.New()
	logger.SetFormatter(&prefixed.TextFormatter{ForceFormatting: true})
	null, _ := os.Open(os.DevNull)
	logger.SetOutput(null)

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
				padding(5,
					container.NewVBox(
						layout.NewSpacer(),
						header(title),
						body("Get ready for the most private mobile operating system in the world"),
						layout.NewSpacer(),
					),
				),
				splashScreen,
			),
		),
	)
}

func selection(logger *logrus.Logger) {
	step := 1
	selectedFile := widget.NewLabel("")
	selectedFile.Hide()
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
				widget.NewLabel(strconv.Itoa(step)+"/7"),
				layout.NewSpacer(),
				nextButton,
			),
			nil,
			nil,
			container.NewGridWithColumns(2,
				padding(5,
					container.NewVBox(
						layout.NewSpacer(),
						header("Select the CalyxOS image"),
						selectedFile,
						container.NewHBox(
							widget.NewButton("Select", func() {
								d := dialog.NewFileOpen(
									func(file fyne.URIReadCloser, err error) {
										if file != nil {
											path = strings.ReplaceAll(file.URI().String(), "file://", "")
											selectedFile.Text = filepath.Base(path)
											selectedFile.Show()
											nextButton.Enable()
										}
									}, window)
								wd, _ := os.Getwd()
								lister, _ := storage.ListerForURI(storage.NewFileURI(wd))
								d.SetLocation(lister)
								//TODO add other archive file extensions
								d.SetFilter(storage.NewExtensionFileFilter([]string{".zip", ".tar.xz", ".tgz"}))
								d.Resize(window.Content().Size())
								d.Show()
							}),
						),
						layout.NewSpacer(),
					),
				),
				splashScreen,
			),
		),
	)
}

func preparation(logger *logrus.Logger, step int) {
	heading := &fyne.Container{}
	text := &widget.Label{}
	screenshot := &canvas.Image{}
	nextButtonLabel := "Next"
	switch step {
	case 1:
		selection(logger)
		return
	case 2:
		heading = header("Prepare Your Device")
		text = body("1. Connect to a wifi network\n2. Remove your SIM card")
		screenshot = &canvas.Image{Resource: resources.ResourceSettingsPanelPng, FillMode: canvas.ImageFillContain}
	case 3:
		heading = header("Enable Developer Mode")
		text = body("1. Go to Settings > About Phone\n2. Tap \"Build number\" 7 times")
		screenshot = &canvas.Image{Resource: resources.ResourceDeveloperModePng, FillMode: canvas.ImageFillContain}
	case 4:
		heading = header("Enable OEM Unlocking")
		text = body("1. Go to Settings > System > Advanced > Developer Options\n2. Tap the toggle labelled \"OEM Unlocking\" to enable it\n3. Press \"Enable\" on the \"Allow OEM unlocking?\" prompt")
		screenshot = &canvas.Image{Resource: resources.ResourceOEMUnlockingPng, FillMode: canvas.ImageFillContain}
	case 5:
		heading = header("Enable USB debugging")
		text = body("1. Go to Settings > System > Advanced > Developer Options\n2. Tap the toggle labelled \"USB debugging\" to enable it\n3. Press \"OK\" on the \"Allow USB debugging?\" prompt")
		screenshot = &canvas.Image{Resource: resources.ResourceEnableUSBDebuggingPng, FillMode: canvas.ImageFillContain}
	case 6:
		heading = header("Connect to Your Computer")
		text = body("1. Plug the device into the computer\n2. Press \"Allow\" on the \"Allow USB debugging?\" prompt")
		screenshot = &canvas.Image{Resource: resources.ResourceUSBDebuggingPng, FillMode: canvas.ImageFillContain}
		nextButtonLabel = "Flash"
	case 7:
		loading := dialog.NewProgressInfinite(title, "Setting up platform tools...", window)
		err := setupPlatformTools(logger)
		loading.Hide()
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		flashing(logger)
		return
	}
	screenshotContainer := padding(5, screenshot)
	window.SetContent(
		container.NewBorder(
			nil,
			container.NewHBox(
				widget.NewButton("Back", func() {
					go func() { preparation(logger, step-1) }()
				}),
				layout.NewSpacer(),
				widget.NewLabel(strconv.Itoa(step)+"/7"),
				layout.NewSpacer(),
				widget.NewButton(nextButtonLabel, func() {
					go func() { preparation(logger, step+1) }()
				}),
			),
			nil,
			nil,
			container.NewGridWithColumns(2,
				padding(5,
					container.NewVBox(
						layout.NewSpacer(),
						heading,
						text,
						layout.NewSpacer(),
					),
				),
				screenshotContainer,
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
		if codename == "walleye" || codename == "jasmine_sprout" {
			line += "Your device will reboot!\n\nWhen your phone reboots you'll need to complete the following steps:\n1. Disconnect the cable and power the device off\n2. Press and hold the volume down and power keys to boot the device into fastboot mode\n3. Reconnect the cable to your device"
		}
		bootloaderWarning := dialog.NewCustom("Hey, heads up...", "Done", body(line), window)
		bootloaderWarning.Resize(window.Content().Size())
		bootloaderWarning.Show()
	}
	return len(p), nil
}

func flashing(logger *logrus.Logger) {
	step := 7
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
		layout.NewSpacer(),
		container.NewCenter(header("Installing "+Vendor)),
		progressBar,
		container.NewHBox(
			widget.NewButton("View Log", func() {
				log.Show()
			}),
		),
		body("* Do not interact with your device unless asked to\n* Do not unplug your device"),
		layout.NewSpacer(),
	)
	footer := container.NewHBox(
		widget.NewButton("Back", func() {
			preparation(logger, step-1)
		}),
		layout.NewSpacer(),
		widget.NewLabel(strconv.Itoa(step)+"/7"),
		layout.NewSpacer(),
	)
	flashButton := widget.NewButton("Retry", func() {
		go func() {
			textGrid.SetText("")
			progressBar.SetValue(0)
			installColumn.Show()
			footer.Hide()
			err := deviceDiscovery(logger)
			if err != nil {
				dialog.ShowError(err, window)
				installColumn.Hide()
				footer.Show()
				return
			}
			err = factoryImageExtraction(logger)
			if err != nil {
				dialog.ShowError(err, window)
				installColumn.Hide()
				footer.Show()
				return
			}
			//err = flashDevices(logger)
			if err != nil {
				dialog.ShowError(err, window)
				installColumn.Hide()
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
			infoColumn,
			installColumn,
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
			container.NewGridWithColumns(2,
				padding(5,
					container.NewVBox(
						layout.NewSpacer(),
						container.NewHBox(&canvas.Image{Resource: resources.ResourceSuccessSvg, FillMode: canvas.ImageFillOriginal}),
						header("You've successfully installed\n"+Vendor+"!"),
						body("It's now safe to unplug your device"),
						layout.NewSpacer(),
					),
				),
				splashScreen,
			),
		),
	)
}

func relock() {
	screenshot := &canvas.Image{Resource: resources.ResourceOEMLockPng, FillMode: canvas.ImageFillContain}
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
				padding(5,
					container.NewVBox(
						layout.NewSpacer(),
						header("Re-enable OEM lock"),
						body("1. Go to Settings > System > Advanced > Developer Options\n2. Tap the toggle labelled \"OEM Unlocking\" to disable it"),
						layout.NewSpacer(),
					),
				),
				padding(5, screenshot),
			),
		),
	)
}
