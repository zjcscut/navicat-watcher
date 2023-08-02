package gui

import (
	"font-fyne/navicat"
	pm "font-fyne/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/pkg/browser"
	"golang.org/x/sys/windows"
	"strconv"
)

const DefaultTheme = "Dark"
const DefaultLink = "https://blog.vlts.cn"
const PwdMask = "********************"

var servers = make(map[string]*navicat.Server)
var serverNameList = binding.NewStringList()

func InitMainWindow() fyne.Window {
	app := fyne.CurrentApp()
	w := app.NewWindow("")
	if desk, ok := app.(desktop.App); ok {
		m := fyne.NewMenu("System Menu",
			fyne.NewMenuItem("Hide", func() {
				w.Hide()
			}),
			fyne.NewMenuItem("Show", func() {
				w.Show()
			}),
			fyne.NewMenuItem("About", func() {
				go syncOpenBrowser(DefaultLink)
			}),
		)
		desk.SetSystemTrayMenu(m)
		desk.SetSystemTrayIcon(pm.Logo)
	}
	vbox := container.NewVBox()
	w.SetContent(vbox)
	version := app.Metadata().Version
	versionGroup := widget.NewFormItem("Version", widget.NewLabelWithStyle(version, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	themeGroup := widget.NewRadioGroup([]string{
		"Dark",
		"Light",
	}, func(st string) {
		app.Settings().SetTheme(&pm.SelectableTheme{Theme: st})
	})
	themeGroup.Required = true
	themeGroup.Horizontal = true
	themeGroup.SetSelected(DefaultTheme)
	themeForm := widget.NewFormItem("Theme", themeGroup)
	sourceBtn := widget.NewButtonWithIcon("Fetch Source Code", theme.FolderOpenIcon(), func() {
		go syncOpenBrowser(DefaultLink)
	})
	metaForm := container.NewVBox(widget.NewForm(versionGroup), widget.NewForm(themeForm), sourceBtn)
	metaCard := widget.NewCard("Meta", "meta info", metaForm)
	serverList := widget.NewListWithData(
		serverNameList,
		func() fyne.CanvasObject {
			label := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			toolbar := widget.NewToolbar(
				widget.NewToolbarAction(theme.SearchIcon(), func() {}),
				widget.NewToolbarSeparator(),
			)
			return container.NewHBox(
				toolbar,
				container.NewMax(label),
			)
		},
		func(item binding.DataItem, o fyne.CanvasObject) {
			toolbar := o.(*fyne.Container).Objects[0].(*widget.Toolbar)
			labelContainer := o.(*fyne.Container).Objects[1].(*fyne.Container)
			label := labelContainer.Objects[0].(*widget.Label)
			content, err := item.(binding.String).Get()
			if err == nil {
				label.SetText(content)
				toolbar.Items[0] = widget.NewToolbarAction(theme.SearchIcon(), func() {
					server, ok := servers[content]
					if ok {
						info := app.NewWindow("")
						info.SetTitle(server.Path)
						info.SetIcon(pm.Logo)
						var items []fyne.CanvasObject
						items = append(items, newNavicatInfoColumn("Name", server.Path, server.Path, info)...)
						items = append(items, newNavicatInfoColumn("Host", server.Host, server.Host, info)...)
						sv := strconv.FormatUint(server.ServerVersion, 10)
						items = append(items, newNavicatInfoColumn("Version", sv, sv, info)...)
						p := strconv.FormatUint(server.Port, 10)
						items = append(items, newNavicatInfoColumn("Port", p, p, info)...)
						items = append(items, newNavicatInfoColumn("Username", server.UserName, server.UserName, info)...)
						hwd := server.HighVersionPassword
						if len(server.HighVersionPassword) > 0 {
							hwd = PwdMask
						}
						lwd := server.LowVersionPassword
						if len(server.LowVersionPassword) > 0 {
							lwd = PwdMask
						}
						items = append(items, newNavicatInfoColumn("Hwd", hwd, server.HighVersionPassword, info)...)
						items = append(items, newNavicatInfoColumn("Lwd", lwd, server.LowVersionPassword, info)...)
						info.SetContent(container.New(layout.NewFormLayout(), items...))
						info.SetFixedSize(true)
						info.CenterOnScreen()
						info.Show()
					}
				})
			}
			toolbar.Refresh()
		})
	serverListScroll := container.NewVScroll(serverList)
	serverListScroll.SetMinSize(fyne.NewSize(0, 350))
	serverCard := widget.NewCard("Server List", "navicat server conf list", serverListScroll)
	vbox.Add(metaCard)
	vbox.Add(serverCard)
	loadBtn := widget.NewButtonWithIcon("Load Navicat Conf", theme.ContentRedoIcon(), func() {
		reloadAllNavicatServers()
	})
	vbox.Add(loadBtn)
	return w
}

func reloadAllNavicatServers() {
	navicatServers, err := navicat.GetNavicatServers()
	if err == nil {
		var serverPathSlice []string
		for _, server := range navicatServers {
			servers[server.Path] = server
			serverPathSlice = append(serverPathSlice, server.Path)
		}
		if len(serverPathSlice) > 0 {
			err := serverNameList.Set(serverPathSlice)
			if err != nil {
				panic(err)
			}
		}
	}
}

func newNavicatInfoColumn(label string, displayInfo string, copyInfo string, iw fyne.Window) []fyne.CanvasObject {
	return []fyne.CanvasObject{container.NewHBox(container.New(layout.NewFormLayout(), widget.NewLabel(label),
		widget.NewLabel(displayInfo))),
		widget.NewToolbar(widget.NewToolbarSeparator(),
			widget.NewToolbarAction(theme.ContentCopyIcon(), func() { iw.Clipboard().SetContent(copyInfo) }))}
}

func syncOpenBrowser(url string) {
	// 优先使用Chrome打开URL
	err := syncOpenChromeBrowser(url)
	if err != nil {
		syncOpenDefaultBrowser(url)
	}
}

// syncOpenChromeBrowser 同步基于Chrome打开链接
func syncOpenChromeBrowser(url string) error {
	// 如果使用APP模式,可以下面的参数可以这样: windows.StringToUTF16Ptr("--app="+url)
	return windows.ShellExecute(0,
		windows.StringToUTF16Ptr("open"),
		windows.StringToUTF16Ptr("chrome.exe"),
		windows.StringToUTF16Ptr(url), nil, windows.SW_SHOWNORMAL)
}

func syncOpenDefaultBrowser(url string) {
	_ = browser.OpenURL(url)
}
