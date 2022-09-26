package libs

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func ListFiles(path string, ext []string, recursive bool, get_abs_path bool) []string {
	cwd, _ := os.Getwd()
	abs_path, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	rel_path, err2 := filepath.Rel(cwd, abs_path)
	if err2 != nil {
		panic(err2)
	}
	if get_abs_path {
		path = abs_path
	} else {
		path = rel_path
	}
	var files []string
	filepath.WalkDir(path, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}
		if info.IsDir() {
			if !recursive {
				return filepath.SkipDir
			}
		}
		if len(ext) == 0 {
			files = append(files, path)
		} else if inArr(ext, strings.ToLower(filepath.Ext(path))) {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func inArr(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}
