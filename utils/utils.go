package utils

import (
	"os"
)

func ListFiles(dir string) ([]string, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}

	files, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, v := range files {
		// fmt.Println(v.Name(), v.IsDir())
		if !v.IsDir() {
			names = append(names, v.Name())
		}
	}
	return names, nil
}
