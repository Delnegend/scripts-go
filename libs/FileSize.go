package libs

import (
	"os"
)

func FileSize(path string) int {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		panic(err)
	}
	return int(info.Size())
}