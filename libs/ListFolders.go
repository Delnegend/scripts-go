package libs

import (
	"os"
	"path/filepath"
)

// list folders without using filepath.Walk
func ListFolders(path string, get_abs_path bool) []string {
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
	var folders []string
	file_infos, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, file_info := range file_infos {
		if file_info.IsDir() {
			folders = append(folders, file_info.Name())
		}
	}
	return folders
}
