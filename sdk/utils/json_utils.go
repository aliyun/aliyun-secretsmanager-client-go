package utils

import (
	"encoding/json"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/logger"
	"io/ioutil"
	"os"
)

func ReadJsonObject(filePath, fileName string, out interface{}) error {
	jsonFile, err := os.Open(filePath + string(os.PathSeparator) + fileName)
	if err != nil {
		return err
	}
	defer func() {
		e := jsonFile.Close()
		if e != nil {
			logger.GetCommonLogger(ModeName).Errorf(e.Error())
		}
	}()
	byteVale, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(byteVale, out)
	if err != nil {
		return err
	}
	return nil
}

func WriteJsonObject(filePath, fileName string, in interface{}) error {
	if !FileExists(filePath, "") {
		err := os.MkdirAll(filePath, os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}
	}
	jsonFile, err := os.OpenFile(filePath+string(os.PathSeparator)+fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func() {
		e := jsonFile.Close()
		if e != nil {
			logger.GetCommonLogger(ModeName).Errorf(e.Error())
		}
	}()
	byteVale, err := json.Marshal(in)
	if err != nil {
		return err
	}
	_, err = jsonFile.Write(byteVale)
	if err != nil {
		return err
	}
	return nil
}

func FileExists(filePath, fileName string) bool {
	if _, err := os.Stat(filePath + string(os.PathSeparator) + fileName); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func FileDelete(filePath, fileName string) error {
	err := os.Remove(filePath + string(os.PathSeparator) + fileName)
	if err != nil {
		return err
	}
	return nil
}
