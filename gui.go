// +build GUI

package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/storage"
	"fyne.io/fyne/widget"
	"image"
	"os"
	"strings"
)

const GUI = true

func gui() {
	a := app.New()
	w := a.NewWindow("Android Factory Image Flasher")
	selectedFile := canvas.NewText(path, image.White)
	flashButton := widget.NewButton("Flash", func() {
		w.Close()
	})
	flashButton.Disable()
	w.SetContent(
		container.NewVBox(
			container.NewHBox(
				widget.NewButton("Select", func() {
					w.Resize(fyne.Size{
						Width:  400,
						Height: 400,
					})
					d := &dialog.FileDialog{}
					d = dialog.NewFileOpen(
						func(file fyne.URIReadCloser, err error) {
							if file != nil {
								path = strings.ReplaceAll(file.URI().String(), "file://", "")
								selectedFile.Text = path
								flashButton.Enable()
							}
							w.Resize(fyne.Size{
								Width:  400,
								Height: 100,
							})
						}, w)
					wd, _ := os.Getwd()
					lister, _ := storage.ListerForURI(storage.NewFileURI(wd))
					d.SetLocation(lister)
					if !parallel {
						//TODO add other archive file extensions
						d.SetFilter(storage.NewExtensionFileFilter([]string{".zip", ".tar.xz", ".tgz"}))
					}
					d.Show()
				}),
				selectedFile,
			),
			layout.NewSpacer(),
			flashButton,
		),
	)
	w.Resize(fyne.Size{
		Width:  400,
		Height: 100,
	})
	w.SetCloseIntercept(func() {
		os.Exit(0)
	})
	w.ShowAndRun()
}