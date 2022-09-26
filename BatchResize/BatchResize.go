package main

import (
	"flag"
	"fmt"
	"libs"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var (
	input       *string
	output      *string
	target_size *string
	threads     *int
	force_srgan *bool
	err_only    *bool
	model       *string
	failed      []string
)

func init() {
	input = flag.String("i", ".", "Input folder")
	output = flag.String("o", "resize_output", "Output folder")
	threads = flag.Int("t", 4, "Number of threads")
	target_size = flag.String("target_size", "w2500", "Destination size (README.md for more info)")
	force_srgan = flag.Bool("srgan", false, "Force resize using RealESRGAN for image has dimension larger than max")
	err_only = flag.Bool("err_only", false, "Only print error files")
	model = flag.String("model", "realesr-animevideov3", "Model for RealESRGAN")
	flag.Parse()
	*input = filepath.Clean(*input)
	*output = filepath.Clean(*output)
}

func resize(input_file, output_file string) error {
	config := *target_size
	mode := config[0:1]
	target_size := libs.StrToInt(config[1:])

	if mode == "r" {
		cmd := exec.Command("realesrgan-ncnn-vulkan", "-n", *model, "-i", input_file, "-o", output_file, "-s", fmt.Sprintf("%d", target_size))
		if err := cmd.Run(); err != nil {
			return err
		}
		return nil
	}

	w, h := libs.Dimension(input_file)
	if h == 0 || w == 0 {
		return fmt.Errorf("%s is not a valid image", input_file)
	}

	// Decide which edge of the image will be use to decide the ratio
	var source_size int
	if (mode == "a" && w > h) || mode == "w" {
		source_size = w
	} else if (mode == "a" && w < h) || mode == "h" {
		source_size = h
	}

	// Decide the ratio
	var ratio int
	if target_size < source_size || source_size*2 >= target_size {
		ratio = 2
	} else if source_size*3 >= target_size {
		ratio = 3
	} else {
		ratio = 4
	}
	if *model == "realesrgan-x4plus-anime" {
		ratio = 4
	}
	// Upscale with RealESRGAN
	upscaled_file := input_file + ".upscaled.png"
	if *force_srgan || source_size < target_size {
		srgan_cmd := exec.Command("realesrgan-ncnn-vulkan", "-i", input_file, "-o", upscaled_file, "-s", fmt.Sprintf("%d", ratio), "-n", *model)
		if err := srgan_cmd.Run(); err != nil {
			return err
		}
	} else {
		if err := libs.Copy(input_file, upscaled_file); err != nil {
			return err
		}
	}

	// Resize to max_size
	if source_size*ratio > target_size && target_size != 0 {
		var cmd_resize *exec.Cmd
		switch mode {
		case "w":
			cmd_resize = exec.Command("ffmpeg", "-i", upscaled_file, "-q:v", "2", "-vf", fmt.Sprintf(`scale='min(%d,iw)':-1`, target_size), output_file)
		case "h":
			cmd_resize = exec.Command("ffmpeg", "-i", upscaled_file, "-q:v", "2", "-vf", fmt.Sprintf(`scale='-1:min(%d,ih)'`, target_size), output_file)
		}
		if err := cmd_resize.Run(); err != nil {
			return err
		}
		os.Remove(upscaled_file)
	} else {
		if err := os.Rename(upscaled_file, output_file); err != nil {
			return err
		}
	}
	return nil
}

func startResize(input_file_list []string) {
	wg := new(sync.WaitGroup)
	files_queue := make(chan string)
	wg.Add(*threads)
	for i := 1; i <= *threads; i++ {
		go func() {
			for input_file := range files_queue {
				// Create path and set format for the output file
				out := libs.ReplaceIO(input_file, *input, *output)
				output_file := out[:len(out)-len(filepath.Ext(out))] + ".png"
				// Create output folder
				if _, err := os.Stat(filepath.Dir(output_file)); os.IsNotExist(err) {
					os.MkdirAll(filepath.Dir(output_file), 0755)
				}
				// Check if output file already exists
				if _, err := os.Stat(output_file); err == nil {
					libs.PrintErr(os.Stderr, "==> Already existed: %s\n", input_file)
					continue
				}
				if err := resize(input_file, output_file); err != nil {
					libs.PrintErr(os.Stderr, "==> Error: %s - %s\n", input_file, err)
					failed = append(failed, input_file)
				} else {
					if !*err_only {
						fmt.Printf("==> %s\n", input_file)
					}
				}

			}
			wg.Done()
		}()
	}
	for _, file := range input_file_list {
		files_queue <- file
	}
	close(files_queue)
	wg.Wait()
}

func main() {
	if libs.Rel(*input) == libs.Rel(*output) {
		libs.PrintErr(os.Stderr, "Input and output folder are the same, please use -o to specify output folder")
		os.Exit(1)
	}
	if !libs.IsDir(*input) {
		libs.PrintErr(os.Stderr, "Input folder must be a directory\n")
	}
	startResize(libs.ListFiles(*input, []string{".png", ".jpg", ".jpeg", ".webp"}, true, false))
	if len(failed) > 0 {
		libs.PrintErr(os.Stderr, "Failed to resize %d images\n", len(failed))
		for _, file := range failed {
			libs.PrintErr(os.Stderr, "%s\n", file)
		}
	}
}
