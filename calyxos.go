//+build GUI,CalyxOS

package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/aospalliance/device-flasher/resources/calyxos"
	"net/url"
)

var splashScreen = &canvas.Image{Resource: calyxos.ResourceCalyxOSPng, FillMode: canvas.ImageFillContain}
var redditUrl, _ = url.Parse("https://reddit.com/r/" + Vendor)
var redditLogo = &canvas.Image{Resource: calyxos.ResourceSnooSvg, FillMode: canvas.ImageFillContain}
var redditScreenshot = &canvas.Image{Resource: calyxos.ResourceRedditPng, FillMode: canvas.ImageFillContain}
var infoColumn *fyne.Container

func init() {
	infoColumn = container.NewGridWithRows(2,
		container.NewGridWithRows(3,
			layout.NewSpacer(),
			redditLogo,
			container.NewCenter(
				container.NewVBox(
					container.NewCenter(
						body("Check out our Reddit community at"),
					),
					container.NewCenter(
						widget.NewHyperlink(redditUrl.Host+redditUrl.Path, redditUrl),
					),
				),
			),
		),
		redditScreenshot,
	)
}
