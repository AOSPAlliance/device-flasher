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
	"github.com/aospalliance/device-flasher/resources/strings"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"path/filepath"
	"strconv"
	str "strings"
)

var application = app.New()
var window = application.NewWindow(strings.Title + " " + version)

var splashScreen = container.NewPadded()
var infoColumn = container.NewGridWithRows(2)

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
	for _, line := range str.Split(text, "\n") {
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
				widget.NewButton(strings.Next, func() {
					selection(logger)
				}),
			),
			nil,
			nil,
			container.NewGridWithColumns(2,
				padding(5,
					container.NewVBox(
						layout.NewSpacer(),
						header(strings.Title),
						body(strings.Welcome),
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
	nextButton := widget.NewButton(strings.Next, func() {
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
		preparation(logger, step+1)
	})
	nextButton.Disable()
	window.SetContent(
		container.NewBorder(
			nil,
			container.NewHBox(
				widget.NewButton(strings.Back, func() {
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
						header(strings.SelectHeader),
						selectedFile,
						container.NewHBox(
							widget.NewButton(strings.Select, func() {
								d := dialog.NewFileOpen(
									func(file fyne.URIReadCloser, err error) {
										if file != nil {
											path = str.ReplaceAll(file.URI().String(), "file://", "")
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
	nextButtonLabel := strings.Next
	switch step {
	case 1:
		selection(logger)
		return
	case 2:
		heading = header(strings.PrepareDeviceHeader)
		text = body(strings.PrepareDeviceBody)
		screenshot = &canvas.Image{Resource: resources.ResourceSettingspanelPng, FillMode: canvas.ImageFillContain}
	case 3:
		heading = header(strings.DeveloperModeHeader)
		text = body(strings.DeveloperModeBody)
		screenshot = &canvas.Image{Resource: resources.ResourceEnabledevelopersettingsPng, FillMode: canvas.ImageFillContain}
	case 4:
		heading = header(strings.OemUnlockingHeader)
		text = body(strings.OemUnlockingBody)
		screenshot = &canvas.Image{Resource: resources.ResourceOemunlockingPng, FillMode: canvas.ImageFillContain}
	case 5:
		heading = header(strings.UsbDebuggingHeader)
		text = body(strings.UsbDebuggingBody)
		screenshot = &canvas.Image{Resource: resources.ResourceEnableusbdebuggingPng, FillMode: canvas.ImageFillContain}
	case 6:
		heading = header(strings.PlugDeviceHeader)
		text = body(strings.PlugDeviceBody)
		screenshot = &canvas.Image{Resource: resources.ResourceAllowusbdebuggingPng, FillMode: canvas.ImageFillContain}
		nextButtonLabel = strings.Flash
	case 7:
		loading := dialog.NewProgressInfinite(strings.Title, strings.SetupPlatformTools, window)
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
				widget.NewButton(strings.Back, func() {
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
	if str.Contains(line, strings.DeviceInstructions) {
		window.RequestFocus()
		codename := line[str.Index(line, "codename=")+len("codename=") : str.Index(line, ":")]
		line = str.Split(line, ": ")[1]
		if codename == "walleye" || codename == "jasmine_sprout" {
			line += strings.DeviceRebootRequired
		}
		bootloaderWarning := dialog.NewCustom(strings.HeadsUp, strings.Done, body(line), window)
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
	log := application.NewWindow(strings.Log)
	log.SetCloseIntercept(func() {
		log.Hide()
	})
	log.SetContent(
		container.NewBorder(
			nil,
			widget.NewButton(strings.Save + " " + strings.Log, func() {
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
		container.NewCenter(header(strings.InstallHeader)),
		progressBar,
		container.NewHBox(
			widget.NewButton(strings.View + " " + strings.Log, func() {
				log.Show()
			}),
		),
		body(strings.InstallBody),
		layout.NewSpacer(),
	)
	footer := container.NewHBox(
		widget.NewButton(strings.Back, func() {
			preparation(logger, step-1)
		}),
		layout.NewSpacer(),
		widget.NewLabel(strconv.Itoa(step)+"/7"),
		layout.NewSpacer(),
	)
	flashButton := widget.NewButton(strings.Retry, func() {
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
			err = flashDevices(logger)
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
				widget.NewButton(strings.Next, func() {
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
						header(strings.SuccessHeader),
						body(strings.SuccessBody),
						layout.NewSpacer(),
					),
				),
				splashScreen,
			),
		),
	)
}

func relock() {
	screenshot := &canvas.Image{Resource: resources.ResourceOemlockPng, FillMode: canvas.ImageFillContain}
	window.SetContent(
		container.NewBorder(
			nil,
			container.NewHBox(
				layout.NewSpacer(),
				widget.NewButton(strings.Finish, func() {
					window.Close()
				}),
			),
			nil,
			nil,
			container.NewGridWithColumns(2,
				padding(5,
					container.NewVBox(
						layout.NewSpacer(),
						header(strings.OemLockHeader),
						body(strings.OemLockBody),
						layout.NewSpacer(),
					),
				),
				padding(5, screenshot),
			),
		),
	)
}
