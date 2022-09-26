package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"libs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	input_folder  *string
	output_folder *string
	create_log    *bool
	threads       *int
	keep_ext      *bool
	output_format *string
	err_only      *bool

	failed_files   []string
	skipped_files  []string
	original_size  int
	converted_size int
)

func init() {
	threads = flag.Int("t", 4, "Threads")
	input_folder = flag.String("i", ".", "Input folder")
	output_folder = flag.String("o", "", "Output folder")
	output_format = flag.String("f", "avif", "Output format: avif, cjxl, djxl, cwebp, gif2webp")
	keep_ext = flag.Bool("k", false, "Keep original extension")
	create_log = flag.Bool("log", false, "Write stderr to log file")
	err_only = flag.Bool("err_only", false, "Only print error files")
	flag.Parse()
	*input_folder = filepath.Clean(*input_folder)

	if libs.InArr(*output_format, []string{"avif", "cjxl", "djxl", "cwebp", "gif2webp"}) == "" {
		libs.PrintErr(os.Stderr, "Error: Invalid output format %s\n", *output_format)
		os.Exit(1)
	}

	if libs.Rel(*output_folder) == "." || libs.Rel(*output_folder) == "" {
		*output_folder = "output_" + *output_format
	} else {
		*output_folder = libs.Rel(*output_folder)
	}

	original_size = 0
	converted_size = 0
}

func convert_avif(file_in string, file_out string, mode string, log *os.File) error {
	var ext, enc, fallback, rep []string
	switch mode {
	case "image":
		ext = []string{"ffmpeg", "-i", file_in, "-strict", "-2", "-pix_fmt", "yuv444p10le", "-f", "yuv4mpegpipe", "-y", file_in + ".y4m"}
		enc = []string{"aomenc", "--codec=av1", "--allintra", "--i444", "--threads=4", "--bit-depth=10", "--max-q=63", "--min-q=0", "--end-usage=q", "--cq-level=25", "--cpu-used=6", "--enable-chroma-deltaq=1", "--qm-min=0", "--aq-mode=1", "--deltaq-mode=3", "--sharpness=2", "--enable-dnl-denoising=0", "--denoise-noise-level=5", "--tune=ssim", "--width={{ width }}", "--height={{ height }}", file_in + ".y4m", "--ivf", "-o", file_in + ".ivf"}
		rep = []string{"MP4Box", "-add-image", fmt.Sprintf("%s.ivf:primary", file_in), "-ab", "avif", "-ab", "miaf", "-new", file_out}
	case "animation":
		ext = []string{"ffmpeg", "-i", file_in, "-strict", "-2", "-pix_fmt", "yuv444p10le", "-f", "yuv4mpegpipe", "-y", file_in + ".y4m"}
		enc = []string{"aomenc", "--codec=av1", "-i444", "--threads=4", "--bit-depth=10", "--max-q=63", "--min-q=0", "--end-usage=q", "--cq-level=18", "--cpu-used=6", "--enable-chroma-deltaq=1", "--qm-min=0", "--aq-mode=1", "--enable-dnl-denoising=0", "--denoise-noise-level=5", "--tune=ssim", "--width={{ width }}", "--height={{ height }}", file_in + ".y4m", "--ivf", "-o", file_in + ".ivf"}
		rep = []string{"ffmpeg", "-i", file_in + ".ivf", "-c", "copy", "-map", "0", "-brand", "avis", "-f", "mp4", file_out}
	}
	// if mode == "image" {
	// } else if mode == "animation" {
	// }

	// extract to y4m
	if err := libs.ExecCmd(log, ext[0], ext[1:]...); err != nil {
		os.Remove(file_in + ".y4m")
		return errors.New("failed to extract")
	}
	w, h := libs.Dimension(file_in + ".y4m")

	// replace {{ width }} and {{ height }} in encoder command with real width and height
	for i, v := range enc {
		if strings.Contains(v, "{{ width }}") {
			enc[i] = strings.Replace(v, "{{ width }}", fmt.Sprintf("%d", w), -1)
		}
		if strings.Contains(v, "{{ height }}") {
			enc[i] = strings.Replace(v, "{{ height }}", fmt.Sprintf("%d", h), -1)
		}
	}

	// encode to ivf
	if err := libs.ExecCmd(log, enc[0], enc[1:]...); err != nil && len(fallback) <= 0 {
		os.Remove(file_in + ".y4m")
		os.Remove(file_in + ".ivy")
		return errors.New("failed to encode")
	} else if err == nil && len(fallback) > 0 {
		os.Remove(file_in + ".ivf")
		if err := libs.ExecCmd(log, fallback[0], fallback[1:]...); err != nil {
			os.Remove(file_in + ".y4m")
			os.Remove(file_in + ".ivf")
			return errors.New("failed to encode fallback")
		}
	}
	// repack to avif
	if err := libs.ExecCmd(log, rep[0], rep[1:]...); err != nil {
		os.Remove(file_in + ".y4m")
		os.Remove(file_in + ".ivf")
		os.Remove(file_out)
		return errors.New("failed to repack")
	}
	os.Remove(file_in + ".y4m")
	os.Remove(file_in + ".ivf")
	return nil
}

func convert_jxl(file_in string, file_out string, mode string, log *os.File) error {
	var cmd *exec.Cmd
	if mode == "compress" {
		cmd = exec.Command("cjxl", file_in, file_out, "-e", "8", "-q", "100", "--num_threads", "4")
	} else {
		cmd = exec.Command("djxl", file_in, file_out)
	}
	if *create_log {
		cmd.Stdout = log
		cmd.Stderr = log
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}
	if err := cmd.Run(); err != nil {
		return errors.New("failed to convert")
	}
	return nil
}

func convert_webp(file_in string, file_out string, mode string, log *os.File) error {
	var cmd *exec.Cmd
	switch mode {
	case "cwebp":
		cmd = exec.Command("cwebp", "-q", "80", "-f", "60", "-m", "6", "-mt", "-o", file_out, "--", file_in)
	case "gif2webp":
		cmd = exec.Command("gif2webp", "-q", "80", "-f", "60", "-m", "6", "-lossy", "-mt", "-o", file_out, "--", file_in)
	}
	if *create_log {
		cmd.Stdout = log
		cmd.Stderr = log
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}
	if err := cmd.Run(); err != nil {
		return errors.New("failed to convert")
	}
	return nil
}

func start_convert(file_list []string, output_ext string, mode string, f func(string, string, string, *os.File) error, log *os.File) {
	wg := new(sync.WaitGroup)
	queue_list := make(chan string)
	wg.Add(*threads)
	for i := 1; i <= *threads; i++ {
		go func() {
			for input_file := range queue_list {
				var name string
				if *keep_ext {
					name = input_file
				} else if !*keep_ext {
					name = strings.TrimSuffix(input_file, filepath.Ext(input_file))
				}
				output_dir := filepath.Dir(libs.ReplaceIO(input_file, *input_folder, *output_folder))
				if _, err := os.Stat(output_dir); os.IsNotExist(err) {
					os.MkdirAll(output_dir, 0755)
				}
				output_file := filepath.Join(output_dir, filepath.Base(name)+output_ext)
				if _, err := os.Stat(output_file); err == nil {
					libs.PrintErr(os.Stderr, "==> Already existed: %s\n", input_file)
					skipped_files = append(skipped_files, input_file)
					continue
				}

				if err := f(input_file, output_file, mode, log); err == nil {
					original_size = original_size + libs.FileSize(input_file)
					converted_size = converted_size + libs.FileSize(output_file)
					if !*err_only {
						fmt.Printf("==> %s\n", libs.Rel(input_file))
					}
				} else {
					failed_files = append(failed_files, input_file)
					libs.PrintErr(os.Stderr, "==> Error: %s - %s\n", input_file, err.Error())
				}
			}
			wg.Done()
		}()
	}
	for _, file := range file_list {
		queue_list <- file
	}
	close(queue_list)
	wg.Wait()
}

func main() {
	// Still avif images and animated avif requires completely different commands to encode so we need to process them separately. For other transcoders we might get away with a single command.
	var images []string
	var animations []string
	var media []string

	// Determine which types of files to convert according to output format
	switch *output_format {
	case "avif":
		images = libs.ListFiles(*input_folder, []string{".jpg", ".jpeg", ".png", ".bmp", ".tif", ".tiff", ".webp", ".gif"}, true, false)
		animations = libs.ListFiles(*input_folder, []string{".gif", ".mp4", ".webm"}, true, false)
		media = append(images, animations...)
	case "cjxl":
		media = libs.ListFiles(*input_folder, []string{".jpg", ".jpeg", ".png", ".bmp", ".tif", ".tiff"}, true, false)
	case "djxl":
		media = libs.ListFiles(*input_folder, []string{".jxl"}, true, false)
	case "cwebp":
		media = libs.ListFiles(*input_folder, []string{".jpg", ".jpeg", ".png", ".bmp", ".tif", ".tiff"}, true, false)
	case "gif2webp":
		media = libs.ListFiles(*input_folder, []string{".gif"}, true, false)
	}

	var log *os.File
	if *create_log {
		log, _ = os.Create(fmt.Sprintf("%s.log", time.Now().Format("2006-01-02-15-04-05")))
	} else

	// Check for duplicates file names. Eg: file.jpg and file.png when keep extension is false with both be converted to file.avif, causing conflict and will overwrite each other.
	if *output_format != "djxl" && !*keep_ext {
		duplicated := func(file_list []string) []string {
			array := make([]string, len(file_list))
			copy(array, file_list)
			var dupls []string
			for i := 0; i < len(array); i++ {
				for j := i; j < len(array); j++ {
					if strings.TrimSuffix(array[i], filepath.Ext(array[i])) == strings.TrimSuffix(array[j], filepath.Ext(array[j])) && (i != j) {
						if libs.InArr(array[i], dupls) != "" {
							dupls = append(dupls, array[i])
						}
						if libs.InArr(array[j], dupls) != "" {
							dupls = append(dupls, array[j])
						}
					}
				}
			}
			return dupls
		}(media)
		if len(duplicated) > 0 {
			fmt.Println("Error: Found duplicated file names:")
			for _, file := range duplicated {
				fmt.Println("\t", file)
			}
			os.Exit(1)
		}
	}

	start_time := time.Now().UnixNano() / 1000000

	switch *output_format {
	case "avif":
		start_convert(images, ".avif", "image", convert_avif, log)
		start_convert(animations, ".avif", "animation", convert_avif, log)
	case "cjxl":
		start_convert(media, ".jxl", "compress", convert_jxl, log)
	case "djxl":
		start_convert(media, ".png", "decompress", convert_jxl, log)
	case "cwebp":
		start_convert(media, ".webp", "cwebp", convert_webp, log)
	case "gif2webp":
		start_convert(media, ".webp", "gif2webp", convert_webp, log)
	}

	if log != nil {
		log.Close()
	}

	if len(failed_files) > 0 {
		fmt.Fprintln(os.Stderr, "")
		libs.PrintErr(os.Stderr, "%d files were failed to convert to %s.\n", len(failed_files), *output_format)
	}

	if !*err_only {
		fmt.Println()
	}
	fmt.Println(libs.ReportResult(len(media), original_size, converted_size, start_time, time.Now().UnixNano()/1000000))
}
