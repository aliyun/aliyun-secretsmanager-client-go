package logger

import (
	"errors"
	"fmt"
	sdkerr "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	cmap "github.com/orcaman/concurrent-map"
	"log"
	"os"
)

type DefaultLogger struct {
	logger *log.Logger
}

func NewDefaultLogger(logger *log.Logger) *DefaultLogger {
	return &DefaultLogger{logger: logger}
}

var defaultLogger *CommonLogger
var commonLoggerMap cmap.ConcurrentMap
var allowModes map[string]struct{}

func init() {
	commonLoggerMap = cmap.New()
	allowModes = make(map[string]struct{})
	allowModes["CacheClient"] = struct{}{}
	defaultLogger = &CommonLogger{
		wrapper:  NewDefaultLogger(log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)),
		modeName: "",
	}
}

type CommonLogger struct {
	wrapper  Wrapper
	modeName string
}

func RegisterLogger(modeName string, wrapper Wrapper) error {
	if _, ok := allowModes[modeName]; !ok {
		return errors.New(fmt.Sprintf("regist log err: the modeName [%s] is invalid", modeName))
	}
	commonLoggerMap.Set(modeName, &CommonLogger{wrapper: wrapper, modeName: modeName})
	return nil
}

func GetCommonLogger(modeName string) *CommonLogger {
	if modeName == "" {
		return defaultLogger
	}
	if !commonLoggerMap.Has(modeName) {
		defaultLogger.Errorf("the modeName [%s] not register, instead of default logger", modeName)
		return defaultLogger
	}
	if v, ok := commonLoggerMap.Get(modeName); ok {
		if logger, ok := v.(*CommonLogger); ok {
			return logger
		} else {
			defaultLogger.Errorf("the logger type err, instead of default logger")
			return defaultLogger
		}
	} else {
		defaultLogger.Errorf("the modeName [%s] not register, instead of default logger", modeName)
		return defaultLogger
	}
}

func IsRegistered(modeName string) bool {
	return commonLoggerMap.Has(modeName)
}

func (d DefaultLogger) Flush() {
}

func (d DefaultLogger) Tracef(format string, params ...interface{}) {
	f := "[Trace] " + format
	d.logger.Printf(f, params...)
}

func (d DefaultLogger) Infof(format string, params ...interface{}) {
	f := "[Info] " + format
	d.logger.Printf(f, params...)
}

func (d DefaultLogger) Debugf(format string, params ...interface{}) {
	f := "[Debug] " + format
	d.logger.Printf(f, params...)
}

func (d DefaultLogger) Warnf(format string, params ...interface{}) {
	f := "[Warn] " + format
	d.logger.Printf(f, params...)
}

func (d DefaultLogger) Errorf(format string, params ...interface{}) {
	f := "[Error] " + format
	d.logger.Printf(f, params...)
}

func (cl *CommonLogger) Flush() {
	cl.wrapper.Flush()
}

func (cl *CommonLogger) Tracef(format string, params ...interface{}) {
	cl.wrapper.Tracef(cl.parseExceptionErrorMsg(format, params...), params...)
}

func (cl *CommonLogger) Infof(format string, params ...interface{}) {
	cl.wrapper.Infof(cl.parseExceptionErrorMsg(format, params...), params...)
}

func (cl *CommonLogger) Debugf(format string, params ...interface{}) {
	cl.wrapper.Debugf(cl.parseExceptionErrorMsg(format, params...), params...)
}

func (cl *CommonLogger) Warnf(format string, params ...interface{}) {
	cl.wrapper.Warnf(cl.parseExceptionErrorMsg(format, params...), params...)
}

func (cl *CommonLogger) Errorf(format string, params ...interface{}) {
	cl.wrapper.Errorf(cl.parseExceptionErrorMsg(format, params...), params...)
}

func (cl *CommonLogger) parseExceptionErrorMsg(format string, params ...interface{}) string {
	if len(params) > 0 {
		param := params[len(params)-1]
		switch err := param.(type) {
		case *sdkerr.ClientError:
			format = format + fmt.Sprintf("\tmodeName:%s\terrorCode:%s\terrMsg:%s\tcaused by:%s", cl.modeName, err.ErrorCode(), err.Message(), err.OriginError().Error())
		case *sdkerr.ServerError:
			format = format + fmt.Sprintf("\tmodeName:%s\terrorCode:%s\terrMsg:%s\terrorRecommend:%s\trequestId:%s", cl.modeName, err.ErrorCode(), err.Message(), err.Recommend(), err.RequestId())
		}
	}
	return format
}
