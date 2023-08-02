package theme

import (
	"fyne.io/fyne/v2"
	fyneTheme "fyne.io/fyne/v2/theme"
	"image/color"
	"strings"
)
import _ "embed"

//go:embed assets/logo.ico
var logoBytes []byte

var Logo = &fyne.StaticResource{
	StaticName:    "icon.ico",
	StaticContent: logoBytes,
}

type SelectableTheme struct {
	Theme string
}

var _ fyne.Theme = (*SelectableTheme)(nil)

func (st SelectableTheme) Color(colorName fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if strings.Contains(st.Theme, "Light") {
		variant = fyneTheme.VariantLight
	} else {
		variant = fyneTheme.VariantDark
	}
	return fyneTheme.DefaultTheme().Color(colorName, variant)
}

func (st SelectableTheme) Icon(iconName fyne.ThemeIconName) fyne.Resource {
	return fyneTheme.DefaultTheme().Icon(iconName)
}

func (st SelectableTheme) Font(ts fyne.TextStyle) fyne.Resource {
	return fyneTheme.DefaultTheme().Font(ts)
}

func (st SelectableTheme) Size(sizeName fyne.ThemeSizeName) float32 {
	return fyneTheme.DefaultTheme().Size(sizeName)
}
