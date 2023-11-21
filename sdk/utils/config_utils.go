package utils

import (
	"bufio"
	"os"
	"strings"
)

func LoadProperties(fileName string) (map[string]string, error) {
	properties := make(map[string]string)
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		str := scanner.Text()
		if strings.HasPrefix(str, "#") {
			continue
		}
		strings.TrimSpace(str)
		if str != "" {
			property := strings.Split(str, "=")
			if len(property) == 2 {
				properties[strings.TrimSpace(property[0])] = strings.TrimSpace(property[1])
			}
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return properties, nil
}

func FileExist(_path string) (bool, error) {
	_, err := os.Stat(_path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
