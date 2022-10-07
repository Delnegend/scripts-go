package libs

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Getting dimensions of a media file, returns width, height | 0, 0 if error
func Dimension(path string) (int, int) {
	if InArr(strings.ToLower(filepath.Ext(path)), []string{".jpg", ".jpeg", ".png", ".gif"}) != "" {
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		img, _, err := image.DecodeConfig(file)
		if err != nil {
			panic(err)
		}
		return img.Width, img.Height
	}
	out, err := exec.Command("ffprobe", "-v", "error", "-select_streams", "v", "-show_entries", "stream=width,height", "-of", "csv=p=0:s=x", path).Output()
	if err != nil {
		panic(err)
	}
	dim := strings.Split(string(out), "x")
	return StrToInt(strings.TrimSpace(dim[0])), StrToInt(strings.TrimSpace(dim[1]))
}
