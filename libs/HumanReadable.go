package libs

import(
	// "strings"
	"fmt"
	"math"
)

// Human readable size
func Hsize(input_ int) string {
	input := float64(input_)
	for _, unit := range []string{"B", "KB", "MB", "GB", "TB"} {
		if input < 1024 {
			return fmt.Sprintf("%.2f %s", float64(input), unit)
		}
		input /= 1024
	}
	return fmt.Sprintf("%.2f TB", float64(input))
}

// Human readable time
func Htime(miliseconds float64) string {
	if miliseconds >= 60*60*1000 {
		return fmt.Sprintf("%dh%dm%ds", int(miliseconds/3600000), int(math.Mod(miliseconds,3600000)/60000), int(math.Mod(miliseconds,60000)/1000))
	}
	if miliseconds >= 60*1000 {
		return fmt.Sprintf("%dm%ds", int(miliseconds/60000), int(math.Mod(miliseconds,60000)/1000))
	}
	if miliseconds >= 1000 {
		return fmt.Sprintf("%.2fs", float64(miliseconds/1000))
	}
	return fmt.Sprintf("%.0fms", miliseconds)
}