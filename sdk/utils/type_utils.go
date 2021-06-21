package utils

import (
	"errors"
	"strings"
)

func ParseString(obj interface{}) (string, error) {
	if obj == nil {
		return "", nil
	}
	if str, ok := obj.(string); ok {
		return str, nil
	} else {
		return "", errors.New("parse string type error")
	}
}

func ParseBool(obj interface{}) (bool, error) {
	if obj == nil {
		return false, nil
	}
	switch v := obj.(type) {
	case string:
		if strings.ToLower(v) == "true" {
			return true, nil
		} else if strings.ToLower(v) == "false" {
			return false, nil
		}
	case bool:
		return v, nil
	}
	return false, errors.New("parse bool failed")
}
