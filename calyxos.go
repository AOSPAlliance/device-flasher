//+build CalyxOS

package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
	"github.com/aospalliance/device-flasher/resources/calyxos"
	"net/url"
)

var splashScreen = &canvas.Image{Resource: calyxos.ResourceCalyxOSSvg, FillMode: canvas.ImageFillContain}
var redditUrl, _ = url.Parse("https://reddit.com/r/" + Vendor)
var redditLogo = &canvas.Image{Resource: calyxos.ResourceSnooSvg, FillMode: canvas.ImageFillContain}
var redditScreenshot = &canvas.Image{Resource: calyxos.ResourceRedditSvg, FillMode: canvas.ImageFillContain}
var infoColumn *fyne.Container

func init() {
	splashScreen.SetMinSize(fyne.Size{
		Width:  480,
		Height: 0,
	})
	redditLogo.SetMinSize(fyne.Size{
		Width:  100,
		Height: 0,
	})
	redditScreenshot.SetMinSize(fyne.Size{
		Width:  100,
		Height: 0,
	})
	infoColumn = container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		container.NewGridWithRows(3,
			redditLogo,
			container.NewVBox(
				container.NewCenter(
					widget.NewLabel("Check out our Reddit community at"),
				),
				container.NewCenter(
					widget.NewHyperlink(redditUrl.Host+redditUrl.Path, redditUrl),
				),
			),
			redditScreenshot,
		),
	)
}
