package main

import (
	_ "font-fyne/font"
	"font-fyne/gui"
	_ "font-fyne/os"
	"font-fyne/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

const H = 600
const W = 400
const App = "NavicatWatcher"

func main() {
	application := app.NewWithID(App)
	application.Preferences().SetBool("__MainWindowInit__", false)
	w := gui.InitMainWindow()
	w.SetTitle(App)
	w.SetIcon(theme.Logo)
	w.CenterOnScreen()
	w.SetMaster()
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(W, H))
	application.Preferences().SetBool("__MainWindowInit__", true)
	w.ShowAndRun()
}
