package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterLogger(t *testing.T) {
	modeName := "CacheClient"
	l := NewDefaultLogger(log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile))
	err := RegisterLogger(modeName, l)
	assert.Nil(t, err)
	modeName = "unknown"
	err = RegisterLogger(modeName, l)
	assert.NotNil(t, err)
}

func TestGetCommonLogger(t *testing.T) {
	modeName := "CacheClient"
	logger := GetCommonLogger(modeName)
	assert.NotNil(t, logger)
}

func TestCommonLogger_Debugf(t *testing.T) {
	modeName := "CacheClient"
	GetCommonLogger(modeName).Debugf("test log:%s", modeName)
}

func TestCommonLogger_Infof(t *testing.T) {
	modeName := "CacheClient"
	err := errors.New(fmt.Sprintf("err message"))
	GetCommonLogger(modeName).Infof("test log:%s, %+v", modeName, err)
}

func TestParseExceptionErrorMsg(t *testing.T) {
	modeName := "CacheClient"
	err := errors.New(fmt.Sprintf("err message"))
	log := &CommonLogger{}
	format := log.parseExceptionErrorMsg("test %s err:%v", modeName, err)
	assert.Equal(t, "test %s err:%v", format)
}
