package main

import (
	"flag"
	"fmt"
	"github.com/gookit/color"
	"libs"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	input           *string
	force_srgan     *bool
	source_format   *string
	no_archive      *bool
	target_size     *string
	format          *string
	single          *bool
	model           *string
	upscale_threads *string
)

func init() {
	input = flag.String("i", ".", "Input folder")
	source_format = flag.String("sf", "jxl", "Source format is 'jxl/webp' or 'png/jpg'")
	no_archive = flag.Bool("no_archive", false, "Only process files, not archive")
	format = flag.String("format", "webp", "Format of processed files: 'avif', 'webp'")
	target_size = flag.String("max", "w2000", "Max size of image in pixel, 0 to disable")
	force_srgan = flag.Bool("srgan", false, "Force resize using RealESRGAN even if image has dimension larger than max")
	single = flag.Bool("single", false, "Treat the input folder as a single pack")
	model = flag.String("model", "realesr-animevideov3", "Model for RealESRGAN")
	upscale_threads = flag.String("t", "4", "Number of threads for BatchResize")
	flag.Parse()
	if *source_format != "jxl" && *source_format != "webp" && *source_format != "png" {
		*source_format = "png"
	}
	*input = filepath.Clean(*input)
}

func stage_upscale(pack_folder string) string {
	var resize_cmd *exec.Cmd
	upscaled_folder := pack_folder + "_upscaled"
	if os.Mkdir(upscaled_folder, os.ModePerm) != nil {
		fmt.Println("Error creating folder", upscaled_folder)
		os.Exit(1)
	}
	input_cmd := []string{"-i", pack_folder, "-o", upscaled_folder, "-t", *upscale_threads ,"-target_size", *target_size, "-err_only", "-model", *model}
	if *force_srgan {
		input_cmd = append(input_cmd, "-srgan")
	}
	resize_cmd = exec.Command("BatchResize", input_cmd...)
	resize_cmd.Stdout = os.Stdout
	resize_cmd.Stderr = os.Stderr
	if err := resize_cmd.Run(); err != nil {
		fmt.Println(err)
	}
	return upscaled_folder
}

func stage_upscale_ani(pack_folder string, upscaled_folder string, formats []string) string {
	for _, file := range libs.ListFiles(pack_folder, formats, true, false) {
		output_file := libs.ReplaceIO(file[:len(file)-len(filepath.Ext(file))], pack_folder, upscaled_folder) + ".webp"
		upscale_cmd := exec.Command("UpscaleAni", "-i", file, "-o", output_file, "-target_size", "h720")
		upscale_cmd.Stdout = os.Stdout
		upscale_cmd.Stderr = os.Stderr
		fmt.Println("==>", file)
		if err := upscale_cmd.Run(); err != nil {
			libs.PrintErr(os.Stderr, "==> Error: %s\n", file)
		}
	}
	return upscaled_folder
}

func stage_transcode(input_folder, output_folder, format string) {
	encode_cmd := exec.Command("BatchConvert", "-i", input_folder, "-o", output_folder, "-f", format, "-err_only")
	encode_cmd.Stdout = os.Stdout
	encode_cmd.Stderr = os.Stderr
	if err := encode_cmd.Run(); err != nil {
		fmt.Println(err)
	}
}

func stage_compress(pack string, name string, compress_type string, included []string) {
	cwd, _ := os.Getwd()
	os.Chdir(pack)
	archive_cmd := exec.Command("7z", append([]string{"a", "-bt", "-t" + compress_type, "-mx=9", "-r", filepath.Join("../", name+"."+compress_type)}, included...)...)
	if err := archive_cmd.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("==> %s\n", pack)
	os.Chdir(cwd)
}

func process(pack string) {
	libs.PrintSign(pack, "main")

	color.Greenf("\nStage 0: Decode JXL to PNG\n")
	if *source_format == "jxl" {
		stage_transcode(pack, pack, "djxl")
	} else {
		fmt.Println("Skipped")
	}
	color.Greenf("\nStage 1: Encode to JXL\n")
	if *source_format == "jxl" || *source_format == "webp" {
		fmt.Println("Skipped")
	} else {
		stage_transcode(pack, pack, "cjxl")
	}

	color.Greenf("\nStage 2: Upscaling images\n")
	upscale_folder := stage_upscale(pack)

	color.Greenf("\nStage 3: Upscaling animations\n")
	transcoded_folder := pack + "_transcoded" // animations get upscale and encode directly into webp without going through a transcode step so we can just throw the output file into the transcoded folder
	stage_upscale_ani(pack, transcoded_folder, []string{".mp4", ".mkv", ".webm", ".gif"})

	color.Greenf("\nStage 4: Transcode\n")
	switch *format {
	case "webp":
		stage_transcode(upscale_folder, transcoded_folder, "cwebp")
	case "avif":
		stage_transcode(upscale_folder, transcoded_folder, "avif")
	}

	color.Greenf("\nStage 5: Compress\n")
	if *no_archive || *single {
		fmt.Println("Skipped")
	} else {
		if *source_format != "jxl" {
			stage_compress(pack, pack, "7z", []string{"*.mp4", "*.webp", "*.gif", "*.jxl"})
		}
		stage_compress(transcoded_folder, pack, "zip", []string{"*.avif", "*.webp"})
	}

	color.Greenf("\nStage 6: Cleanup\n")
	if *no_archive || *single {
		fmt.Println("Skipped")
		if *source_format != "jxl" {
			jxl_folder := pack + "_jxl"
			os.Mkdir(jxl_folder, os.ModePerm)
			for _, file := range libs.ListFiles(pack, []string{".jxl"}, true, false) {
				os.Rename(file, libs.ReplaceIO(file, pack, jxl_folder))
			}
		}
	} else {
		os.RemoveAll(transcoded_folder)
	}
	os.RemoveAll(upscale_folder)
}

func main() {
	if !libs.IsDir(*input) {
		libs.PrintErr(os.Stderr, "Input must be a folder\n")
		os.Exit(1)
	}

	os.Stdout.Write([]byte("\033[H\033[2J")) // clear the console
	fmt.Printf("Source format: %s | Create archive: %t | Target format: %s\n", *source_format, !*no_archive, *format)

	if !*single {
		os.Chdir(*input)
		for _, pack := range libs.ListFolders(".", false) {
			process(pack)
		}
	} else {
		process(*input)
	}
	exec.Command("wscript", "D:\\sound.vbs").Run()
}
