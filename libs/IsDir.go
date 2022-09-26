package libs

import (
	"os"
)

func IsDir(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fileInfo, err2 := file.Stat()
	if err2 != nil {
		panic(err)
	}
	return fileInfo.IsDir()
}
