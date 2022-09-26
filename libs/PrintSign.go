package libs

import (
	"fmt"
	"os"
	"strings"

	"github.com/gookit/color"
)

func r(input string, times int) string {
	return strings.Repeat(input, times)
}

func PrintSign(sign string, size string) {
	slen := len(sign)
	fmt.Println()
	switch size {
	case "main":
		color.Cyanln(r("=", slen+8))
		color.Cyanln("=", r(" ", slen+4), "=")
		color.Cyanln("=  ", sign, "  =")
		color.Cyanln("=", r(" ", slen+4), "=")
		color.Cyanln(r("=", slen+8))
	case "small":
		color.Greenf("%s\n%s\n%s\n", r("=", slen), sign, r("=", slen))
	}
}

func PrintErr(output *os.File, format string, a ...any) {
	// custom_style := color.New(color.Red, color.BgLightYellow, color.OpBold)
	custom_style := color.New(color.Red, color.OpBold, color.OpFastBlink)
	fmt.Fprint(output, custom_style.Sprintf(format, a...))
}
