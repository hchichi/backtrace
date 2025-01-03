// backtrace-main/.github/bk/colors.go
package backtrace

import "fmt"

// HiThistle returns a bold thistle formatted string (#E8BBD0)
func HiThistle(text string) string {
	return fmt.Sprintf("\033[38;2;232;187;208m\033[01m%s\033[0m", text)
}

// HiGoldenrod returns a bold goldenrod formatted string (#CDAB53)
func HiGoldenrod(text string) string {
	return fmt.Sprintf("\033[38;2;205;171;83m\033[01m%s\033[0m", text)
}

// HiDarkOrange returns a bold dark orange formatted string (#FF8E1E)
func HiDarkOrange(text string) string {
	return fmt.Sprintf("\033[38;2;255;142;30m\033[01m%s\033[0m", text)
}

// HiChartreuse returns a bold chartreuse formatted string (#8EFF1E)
func HiChartreuse(text string) string {
	return fmt.Sprintf("\033[38;2;142;255;30m\033[01m%s\033[0m", text)
}

// HiSpringGreen returns a bold spring green formatted string (#1EFF8E)
func HiSpringGreen(text string) string {
	return fmt.Sprintf("\033[38;2;30;255;142m\033[01m%s\033[0m", text)
}

// HiMediumSeaGreen returns a bold medium sea green formatted string (#01A252)
func HiMediumSeaGreen(text string) string {
	return fmt.Sprintf("\033[38;2;1;162;82m\033[01m%s\033[0m", text)
}

// HiLightBlue returns a bold light blue formatted string (#B5E4F4)
func HiLightBlue(text string) string {
	return fmt.Sprintf("\033[38;2;181;228;244m\033[01m%s\033[0m", text)
}

// HiTurquoise returns a bold turquoise formatted string (#C8FAF4)
func HiTurquoise(text string) string {
	return fmt.Sprintf("\033[38;2;200;250;244m\033[01m%s\033[0m", text)
}

// HiSlateGray returns a bold slate gray formatted string (#70A598)
func HiSlateGray(text string) string {
	return fmt.Sprintf("\033[38;2;112;165;152m\033[01m%s\033[0m", text)
}

// HiOlive returns a bold olive formatted string (#87875F)
func HiOlive(text string) string {
	return fmt.Sprintf("\033[38;2;135;135;95m\033[01m%s\033[0m", text)
}

// HiIndianRed returns a bold Indian red formatted string (#AF5F5F)
func HiIndianRed(text string) string {
	return fmt.Sprintf("\033[38;2;175;95;95m\033[01m%s\033[0m", text)
}

// HiLightCoral returns a bold light coral formatted string (#F8747E)
func HiLightCoral(text string) string {
	return fmt.Sprintf("\033[38;2;248;116;126m\033[01m%s\033[0m", text)
}

// HiCrimson returns a bold crimson formatted string (#CD5C5C)
func HiCrimson(text string) string {
	return fmt.Sprintf("\033[38;2;205;92;92m\033[01m%s\033[0m", text)
}

// HiOliveDrab returns a bold olive drab formatted string (#88AA22)
func HiOliveDrab(text string) string {
	return fmt.Sprintf("\033[38;2;136;170;34m\033[01m%s\033[0m", text)
}

// HiTurquoiseBlue returns a bold turquoise blue formatted string (#27FFDF)
func HiTurquoiseBlue(text string) string {
	return fmt.Sprintf("\033[38;2;39;255;223m\033[01m%s\033[0m", text)
}

// HiFireBrick returns a bold fire brick formatted string (#CA4949)
func HiFireBrick(text string) string {
	return fmt.Sprintf("\033[38;2;202;73;73m\033[01m%s\033[0m", text)
}

// HiSeaGreen returns a bold sea green formatted string (#2D8F6F)
func HiSeaGreen(text string) string {
	return fmt.Sprintf("\033[38;2;45;143;111m\033[01m%s\033[0m", text)
}

// HiPeach returns a bold peach formatted string (#ec9255)
func HiPeach(text string) string {
	return fmt.Sprintf("\033[38;2;236;146;85m\033[01m%s\033[0m", text)
}
