package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var (
	colBg      = color.NRGBA{0x0F, 0x11, 0x15, 0xFF}
	colCard    = color.NRGBA{0x17, 0x1A, 0x21, 0xFF}
	colBorder  = color.NRGBA{0x26, 0x2B, 0x36, 0xFF}
	colText    = color.NRGBA{0xE6, 0xE8, 0xEE, 0xFF}
	colMuted   = color.NRGBA{0x8A, 0x90, 0xA0, 0xFF}
	colAccent  = color.NRGBA{0x7C, 0x5C, 0xFF, 0xFF}
	colSuccess = color.NRGBA{0x3D, 0xDC, 0x97, 0xFF}
	colError   = color.NRGBA{0xFF, 0x5C, 0x7A, 0xFF}
)

// 다크 고정 + 보라 포인트 컬러 테마.
type modernTheme struct{}

func (modernTheme) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return colBg
	case theme.ColorNameButton, theme.ColorNameInputBackground, theme.ColorNameOverlayBackground, theme.ColorNameMenuBackground:
		return colCard
	case theme.ColorNameForeground:
		return colText
	case theme.ColorNamePlaceHolder:
		return colMuted
	case theme.ColorNamePrimary, theme.ColorNameFocus:
		return colAccent
	case theme.ColorNameHover:
		return color.NRGBA{0x2A, 0x30, 0x40, 0xFF}
	case theme.ColorNameSeparator: // 목록 구분선 숨김(카드로 구분)
		return color.Transparent
	case theme.ColorNameInputBorder:
		return colBorder
	case theme.ColorNameSuccess:
		return colSuccess
	case theme.ColorNameError:
		return colError
	}
	return theme.DefaultTheme().Color(n, theme.VariantDark)
}

func (modernTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (modernTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }

func (modernTheme) Size(n fyne.ThemeSizeName) float32 {
	switch n {
	case theme.SizeNameInputRadius, theme.SizeNameSelectionRadius:
		return 8
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameText:
		return 14
	}
	return theme.DefaultTheme().Size(n)
}
