package utils

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
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

func WriteLineToFile(filename, content string) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	writer.WriteString(content)
	writer.WriteString("\n")
	writer.Flush()
	return nil
}

func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		fmt.Println(err)
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Println(localAddr.String())
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}
