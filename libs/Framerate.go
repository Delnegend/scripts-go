package libs

import (
	"os/exec"
	"strings"
)

func Framerate(file string) float64 {
	raw_fr, _ := exec.Command("ffprobe", "-v", "0", "-of", "csv=p=0", "-select_streams", "v:0", "-show_entries", "stream=r_frame_rate", file).Output()
	math_fr := strings.Split(string(raw_fr), "/")
	return StrToFloat(strings.TrimSpace(math_fr[0])) / StrToFloat(strings.TrimSpace(math_fr[1]))
}
