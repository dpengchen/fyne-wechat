package util

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

var application *fyne.App

func InitWindowUtil() *fyne.App {
	app := app.New()
	application = &app
	return application
}
