package utils

import (
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func Ping(host string) float64 {
	var args string
	var pattern string
	if runtime.GOOS == "windows" {
		args = "-n"
		pattern = "=(\\d+)(.\\d+)?ms"
	} else {
		args = "-c"
		pattern = "=(\\d+)(.\\d+)? ms"
	}
	out, err := exec.Command("ping", args, "1", host).Output()
	if err != nil {
		return -1
	}
	outStr := string(out)
	if strings.Count(outStr, "=") == 0 {
		return -1
	}
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return -1
	}
	matches := regex.FindStringSubmatch(outStr)
	if matches == nil {
		return -1
	}
	if len(matches) == 3 {
		t, err := strconv.ParseFloat(matches[1]+matches[2], 64)
		if err != nil {
			return -1
		}
		return t
	} else if len(matches) == 2 {
		t, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return -1
		}
		return t
	}
	return -1
}
