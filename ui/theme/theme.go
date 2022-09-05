package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type Theme struct {
}

func (t *Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorRed:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorOrange:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorYellow:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorGreen:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorBlue:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorPurple:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorBrown:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorGray:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNameBackground:
		return &color.RGBA{R: 247, G: 252, B: 249, A: 255} //
	case theme.ColorNameButton:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNameDisabledButton:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNameDisabled:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNameError:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNameFocus:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNameForeground:
		return &color.RGBA{R: 98, G: 107, B: 130, A: 255} //
	case theme.ColorNameHover:
		return &color.RGBA{R: 230, G: 230, B: 230, A: 255} //
	case theme.ColorNameInputBackground:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNamePlaceHolder:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNamePressed:
		return &color.RGBA{R: 128, G: 128, B: 128, A: 255} //
	case theme.ColorNamePrimary:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNameScrollBar:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNameSelection:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNameShadow:
		return &color.RGBA{R: 162, G: 179, B: 219, A: 255} //
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *Theme) Font(textStyle fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(textStyle)
}

func (t *Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *Theme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameCaptionText:
		return theme.DefaultTheme().Size(name)
	case theme.SizeNameInlineIcon:
		return theme.DefaultTheme().Size(name)
	case theme.SizeNamePadding:
		return theme.DefaultTheme().Size(name)
	case theme.SizeNameScrollBar:
		return theme.DefaultTheme().Size(name)
	case theme.SizeNameScrollBarSmall:
		return theme.DefaultTheme().Size(name)
	case theme.SizeNameSeparatorThickness:
		return theme.DefaultTheme().Size(name)
	case theme.SizeNameText:
		return theme.DefaultTheme().Size(name)
	case theme.SizeNameHeadingText:
		return theme.DefaultTheme().Size(name) * 0.75
	case theme.SizeNameSubHeadingText:
		return theme.DefaultTheme().Size(name) * 0.75
	case theme.SizeNameInputBorder:
		return theme.DefaultTheme().Size(name)
	default:
		return theme.DefaultTheme().Size(name)
	}
}
