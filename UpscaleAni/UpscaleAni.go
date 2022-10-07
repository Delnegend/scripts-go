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
	input   *string
	output  *string
	max     *string
	cleanup *bool
	model   *string
	force   *bool
)

func init() {
	input = flag.String("i", "", "Input file")
	output = flag.String("o", "", "Output file (available format: .mkv, .mp4, .gif, .webp, .mov)")
	max = flag.String("max", "h2160", "Max resolution <dimension><pixel>")
	cleanup = flag.Bool("cleanup", false, "Delete extracted frames and upscaled frames")
	model = flag.String("model", "realesr-animevideov3", "Model name")
	force = flag.Bool("force", false, "Force upscale")
	flag.Parse()
	*input = libs.Rel(*input)
	if *output == "" {
		libs.PrintErr(os.Stderr, "Output file is required\n")
		os.Exit(1)
	} else {
		*output = libs.Rel(*output)
	}
}

func upscale(input string, output string, width int, height int, side string, max int) error {
	// Get framerate
	frate := fmt.Sprintf("%.2f", libs.Framerate(input))
	fmt.Printf("Framerate: %s\n", frate)

	// Create folders
	frames := input + "_frames"
	upscaled := input + "_upscaled"
	if _, err := os.Stat(frames); os.IsNotExist(err) {
		os.Mkdir(frames, 0755)
	}
	if _, err := os.Stat(upscaled); os.IsNotExist(err) {
		os.Mkdir(upscaled, 0755)
	}
	if _, err := os.Stat(filepath.Dir(output)); os.IsNotExist(err) {
		os.Mkdir(filepath.Dir(output), 0755)
	}

	// Extract frames
	fmt.Println("Extracting frames...")
	cmd := exec.Command("ffmpeg", "-i", input, "-q:v", "2", filepath.Join(frames, "%04d.jpg"))
	// cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		libs.PrintErr(os.Stderr, "Error:%s\n", err)
		os.Exit(1)
	}

	// Calculate scale factor for realESRGAN
	// region
	source_size := height
	if (side == "a" && width > height) || side == "w" {
		source_size = width
	}
	var ratio int
	if source_size*2 >= max {
		ratio = 2
	} else if source_size*3 >= max {
		ratio = 3
	} else {
		ratio = 4
	}
	if *model == "realesrgan-x4plus" || *model == "realesrgan-x4plus-anime" || *model == "realesrnet-x4plus" {
		ratio = 4
	}
	// endregion

	// Upscale frames
	fmt.Println("Upscaling frames...")
	upscale_cmd := exec.Command("realesrgan-ncnn-vulkan", "-i", frames, "-o", upscaled, "-s", fmt.Sprintf("%d", ratio), "-f", "jpg", "-n", *model)
	if err := upscale_cmd.Run(); err != nil {
		return err
	}

	// Merging frames
	fmt.Println("Merging frames...")
	if err := ffmpeg_encode(filepath.Join(upscaled, "%04d.jpg"), output, frate, side, fmt.Sprintf("%d", max)); err != nil {
		return err
	}

	// Clean up
	if *cleanup {
		fmt.Println("Cleaning stuffs...")
		os.RemoveAll(upscaled)
		os.RemoveAll(frames)
	}

	return nil
}

func ffmpeg_encode(input, output, framerate, side, max string) error {
	ffmpeg_resize := fmt.Sprintf(`scale='-1:min(%s,ih)'`, max)
	if side == "w" {
		ffmpeg_resize = fmt.Sprintf(`scale='min(%s,iw)':-1`, max)
	}
	var merge_params []string
	if framerate == "0" {
		merge_params = []string{"ffmpeg", "-i", input, "-vf", ffmpeg_resize}
	} else {
		merge_params = []string{"ffmpeg", "-i", input, "-r", framerate, "-vf", ffmpeg_resize}
	}
	switch filepath.Ext(output) {
	case ".gif":
		merge_params = append(merge_params, "-loop", "0", output)
	case ".webp":
		merge_params = append(merge_params, "-loop", "0", "-compression_level", "6", "-quality", "80", output)
	case ".mov":
		merge_params = append(merge_params, "-c:v", "prores_ks", "-profile:v", "4", output)
	default:
		merge_params = append(merge_params, "-f", "image2", "-codec", "copy", output)
	}
	merge_cmd := exec.Command(merge_params[0], merge_params[1:]...)
	// merge_cmd.Stderr = os.Stderr
	// merge_cmd.Stdout = os.Stdout
	if err := merge_cmd.Run(); err != nil {
		return err
	}
	return nil
}

func main() {
	// Get max config resolution, source file resolution
	size_config := *max
	max_output_size := libs.StrToInt(size_config[1:])
	side_to_resize := size_config[0:1]
	w, h := libs.Dimension(*input)
	if h == 0 || w == 0 {
		libs.PrintErr(os.Stderr, "%s is not a valid video\n", *input)
		os.Exit(1)
	}
	if !(side_to_resize == "h" && (max_output_size <= h)) || (side_to_resize == "w" && (max_output_size <= w)) || *force {
		// Upscale function
		fmt.Println("Upscaling...")
		if err := upscale(*input, *output, w, h, side_to_resize, max_output_size); err != nil {
			libs.PrintErr(os.Stderr, "Error:%s\n", err)
			os.Exit(1)
		}
	} else {
		// Downscale function
		fmt.Println("Downscaling...")
		if err := ffmpeg_encode(*input, *output, "0", side_to_resize, fmt.Sprintf("%d", max_output_size)); err != nil {
			libs.PrintErr(os.Stderr, "Error:%s\n", err)
		}
	}
}
