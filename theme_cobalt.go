// Code generated by fyne-theme-generator

package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type cobaltTheme struct{}

func (cobaltTheme) Color(c fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch c {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x15, G: 0x15, B: 0x15, A: 0xff}
	case theme.ColorNameButton:
		return color.NRGBA{R: 0x18, G: 0x18, B: 0x18, A: 0xfc}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x26, G: 0x26, B: 0x26, A: 0xff}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x42}
	case theme.ColorNameError:
		return color.NRGBA{R: 0xf5, G: 0x12, B: 0x1, A: 0xff}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0x6d, G: 0x6b, B: 0x6b, A: 0xed}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case theme.ColorNameHover:
		return color.NRGBA{R: 0x57, G: 0x59, B: 0x5b, A: 0xff}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x2b}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0xb2, G: 0xb2, B: 0xb2, A: 0xff}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xe2}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0xbc, G: 0xb8, B: 0xbb, A: 0xff}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0xf1, G: 0xef, B: 0xef, A: 0x99}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0xa, G: 0x8, B: 0x8, A: 0x88}
	default:
		return theme.DefaultTheme().Color(c, v)
	}
}

func (cobaltTheme) Font(s fyne.TextStyle) fyne.Resource {
	if s.Monospace {
		return theme.DefaultTheme().Font(s)
	}
	if s.Bold {
		if s.Italic {
			return theme.DefaultTheme().Font(s)
		}
		return theme.DefaultTheme().Font(s)
	}
	if s.Italic {
		return theme.DefaultTheme().Font(s)
	}
	return theme.DefaultTheme().Font(s)
}

func (cobaltTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (cobaltTheme) Size(s fyne.ThemeSizeName) float32 {
	switch s {
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameScrollBar:
		return 16
	case theme.SizeNameScrollBarSmall:
		return 3
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 14
	case theme.SizeNameInputBorder:
		return 2
	default:
		return theme.DefaultTheme().Size(s)
	}
}
