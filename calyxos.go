//+build CalyxOS

package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
	"net/url"
)

const vendor = "CalyxOS"

var redditUrl, _ = url.Parse("https://reddit.com/r/" + vendor)
var redditLogo = &canvas.Image{Resource: resourceRedditSvg, FillMode: canvas.ImageFillContain}
//var redditScreenshot = &canvas.Image{Resource: , FillMode: canvas.ImageFillContain}
var infoColumn *fyne.Container

func init() {
	redditLogo.SetMinSize(fyne.Size{
		Width:  100,
		Height: 0,
	})
	//redditScreenshot.SetMinSize(fyne.Size{
	//	Width:  100,
	//	Height: 0,
	//})
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
			//redditScreenshot,
		),
	)
}
