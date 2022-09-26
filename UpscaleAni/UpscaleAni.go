package main

import (
	"flag"
	"fmt"
	"libs"
	"os"
	"os/exec"
	"path/filepath"
	// "strings"
)

// input args
var (
	input       *string
	output      *string
	max         *string
	noti        *bool
	limit_frate *float64
	keep        *bool
	model       *string
)

func init() {
	input = flag.String("i", "", "Input file")
	output = flag.String("o", "", "Output file (available format: .mkv, .mp4, .gif, .webp, .mov)")
	max = flag.String("max", "h2160", "Max resolution <dimension><pixel>")
	noti = flag.Bool("noti", true, "Notify when done")
	limit_frate = flag.Float64("f", 0, "Change output framerate")
	keep = flag.Bool("k", true, "Keep upscaled frames after encoding")
	model = flag.String("model", "realesr-animevideov3", "Model name")
	flag.Parse()
	*input = libs.Rel(*input)
	if *output == "" {
		libs.PrintErr(os.Stderr, "Output file is required\n")
	} else {
		*output = libs.Rel(*output)
	}
}

func main() {
	var frate string
	if *limit_frate > 0 {
		frate = fmt.Sprintf("%.2f", *limit_frate)
	} else {
		frate = fmt.Sprintf("%.2f", libs.Framerate(*input))
	}
	fmt.Printf("Framerate: %s\n", frate)

	// Create folders
	frames := *input + "_frames"
	upscaled := *input + "_upscaled"
	if _, err := os.Stat(frames); os.IsNotExist(err) {
		os.Mkdir(frames, 0755)
	}
	if _, err := os.Stat(upscaled); os.IsNotExist(err) {
		os.Mkdir(upscaled, 0755)
	}
	if _, err := os.Stat(filepath.Dir(*output)); os.IsNotExist(err) {
		os.Mkdir(filepath.Dir(*output), 0755)
	}

	size_config := *max
	target_size := libs.StrToInt(size_config[1:])
	mode := size_config[0:1]
	w, h := libs.Dimension(*input)
	if h == 0 || w == 0 {
		libs.PrintErr(os.Stderr, "%s is not a valid video\n", *input)
		os.Exit(1)
	}
	var source_size int
	if mode == "w" {
		source_size = w
	} else if mode == "h" {
		source_size = h
	}
	var ratio string
	if target_size < source_size || source_size*2 >= target_size {
		ratio = "2"
	} else if source_size*3 >= target_size {
		ratio = "3"
	} else {
		ratio = "4"
	}
	if *model == "realesrgan-x4plus-anime" {
		ratio = "4"
	}
	fmt.Println("Extracting frames...")
	cmd := exec.Command("ffmpeg", "-i", *input, "-q:v", "2", filepath.Join(frames, "%04d.jpg"))
	if err := cmd.Run(); err != nil {
		libs.PrintErr(os.Stderr, "Error:%s\n", err)
		os.Exit(1)
	}

	fmt.Println("Upscaling frames...")
	upscale_cmd := exec.Command("realesrgan-ncnn-vulkan", "-i", frames, "-o", upscaled, "-s", ratio, "-f", "jpg", "-n", *model)
	if err := upscale_cmd.Run(); err != nil {
		libs.PrintErr(os.Stderr, "Error:%s\n", err)
		os.Exit(1)
	}

	fmt.Println("Merging frames...")
	var resize string
	if mode == "h" {
		resize = fmt.Sprintf(`scale='-1:min(%d,ih)'`, target_size)
	} else if mode == "w" {
		resize = fmt.Sprintf(`scale='min(%d,iw)':-1`, target_size)
	} else {
		libs.PrintErr(os.Stderr, "Error:%s is not a valid mode\n", mode)
		os.Exit(1)
	}
	merge_cmd_params := []string{"ffmpeg", "-i", filepath.Join(upscaled, "%04d.jpg"), "-r", frate, "-vf", resize}
	switch filepath.Ext(*output) {
	case ".gif":
		merge_cmd_params = append(merge_cmd_params, "-loop", "0")
	case ".webp":
		merge_cmd_params = append(merge_cmd_params, "-loop", "0", "-compression_level", "6")
	case ".mov":
		merge_cmd_params = append(merge_cmd_params, "-c:v", "prores_ks", "-profile:v", "4")
	default:
		merge_cmd_params = append(merge_cmd_params, "-f", "image2", "-r", frate, "-codec", "copy")
	}
	merge_cmd_params = append(merge_cmd_params, *output)
	merge_cmd := exec.Command(merge_cmd_params[0], merge_cmd_params[1:]...)
	if err := merge_cmd.Run(); err != nil {
		libs.PrintErr(os.Stderr, "Error:%s\n", err)
		os.Exit(1)
	}

	if !*keep {
		fmt.Println("Cleaning stuffs...")
		os.RemoveAll(upscaled)
	}
	os.RemoveAll(frames)
	if *noti {
		exec.Command("wscript", "D:\\sound.vbs").Run()
	}
}
