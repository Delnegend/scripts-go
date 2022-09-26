package main

import (
	"flag"
	"fmt"
	"libs"
	"os"
	"os/exec"
	"path/filepath"

	// "runtime"
	"strings"
	// "sync"
)

var (
	ext    *string
	format *string
)

func init() {
	format = flag.String("f", ".7z", "Format of output file, .7z or .zip")
	ext = flag.String("e", "*", "Extension to be included in output file, * for all, or a list of extensions including the dot")
	flag.Parse()
}

func compress(path string, format string, ext_list_ string) {
	os.Chdir(path)
	compress_cmd := []string{"7z", "a", "-bt", "-t" + format[1:], "-mx=9", "-r", filepath.Join("../", path+format)}
	ext_list := strings.Split(ext_list_, " ")
	if ext_list[0] == "*" {
		compress_cmd = append(compress_cmd, "*.*")
	} else {
		compress_cmd = append(compress_cmd, func(exts []string) []string {
			var ret []string
			for _, ext := range exts {
				ret = append(ret, "*"+ext)
			}
			return ret
		}(ext_list)...)
	}
	cmd := exec.Command(compress_cmd[0], compress_cmd[1:]...)
	if err := cmd.Run(); err != nil {
		libs.PrintErr(os.Stderr, "Error: %s\n%s\n", path, err)
	}
	fmt.Printf("==> %s\n", path)
	os.Chdir("..")
}

func main() {
	folders_to_compress := libs.ListFolders(".", false)
	for _, folder := range folders_to_compress {
		compress(folder, *format, *ext)
	}
}
