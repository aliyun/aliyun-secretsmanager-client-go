package utils

import (
	"bufio"
	"io"
	"os"
	"strings"
)

func LoadProperties(fileName string) (map[string]string, error) {
	properties := make(map[string]string)
	if !FileExist(fileName) {
		return nil, nil
	}
	srcFile, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	defer srcFile.Close()
	if err != nil {
		return nil, err
	}
	srcReader := bufio.NewReader(srcFile)
	for {
		str, err := srcReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		if 0 == len(str) || str == "\n" || strings.HasPrefix(strings.TrimSpace(str), "#") {
			continue
		}
		if strings.HasSuffix(str, "\r\n") {
			str = strings.Replace(str, "\r\n", "", -1)
		} else if strings.HasSuffix(str, "\n") {
			str = strings.Replace(str, "\n", "", -1)
		}
		property := strings.Split(str, "=")
		properties[strings.TrimSpace(property[0])] = strings.TrimSpace(property[1])
	}
	return properties, nil
}

func FileExist(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
