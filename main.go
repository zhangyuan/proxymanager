package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const AppName = "Proxy Helper"

func main() {
	a := app.New()
	w := a.NewWindow(AppName)
	w.SetContent(widget.NewLabel("Hello World!"))

	if desk, ok := a.(desktop.App); ok {
		m := fyne.NewMenu(AppName,
			fyne.NewMenuItem("Show", func() {
				w.Show()
				log.Println("Tapped show")
			}))
		desk.SetSystemTrayMenu(m)
	}

	w.SetCloseIntercept(func() {
		w.Hide()
	})

	w.ShowAndRun()
}
