//+build GUI,CalyxOS

package main

import (
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/aospalliance/device-flasher/resources/calyxos"
	"net/url"
)

var redditUrl, _ = url.Parse("https://reddit.com/r/" + Vendor)
var redditLogo = &canvas.Image{Resource: calyxos.ResourceSnooSvg, FillMode: canvas.ImageFillContain}
var redditScreenshot = &canvas.Image{Resource: calyxos.ResourceRedditPng, FillMode: canvas.ImageFillContain}

func init() {
	splashScreen = padding(5, &canvas.Image{Resource: calyxos.ResourceCalyxOSPng, FillMode: canvas.ImageFillContain})
	infoColumn = container.NewGridWithRows(2,
		container.NewGridWithRows(3,
			layout.NewSpacer(),
			redditLogo,
			container.NewCenter(
				container.NewVBox(
					container.NewCenter(
						widget.NewLabel("Check out our Reddit community at"),
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
