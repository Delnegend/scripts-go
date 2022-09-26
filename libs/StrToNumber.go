package libs

import (
	"strconv"
)

func StrToInt(str string) int {
	i, _ := strconv.Atoi(str)
	return i
}
func StrToFloat(str string) float64 {
	f, _ := strconv.ParseFloat(str, 64)
	return f
}
