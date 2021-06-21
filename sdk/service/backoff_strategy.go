package service

import (
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/utils"
	"math"
)

// BackoffStrategy 规避重试策略接口
type BackoffStrategy interface {
	// 初始化策略
	Init() error

	// 获取规避等待时间，时间单位MS
	GetWaitTimeExponential(retryTimes int) int64
}

type FullJitterBackoffStrategy struct {
	//重试最大尝试次数
	RetryMaxAttempts int
	// 重试时间间隔，单位ms
	RetryInitialIntervalMills int64
	// 最大等待时间，单位ms
	Capacity int64
}

func NewFullJitterBackoffStrategy(retryMaxAttempts int, retryInitialIntervalMills int64, capacity int64) *FullJitterBackoffStrategy {
	return &FullJitterBackoffStrategy{
		RetryMaxAttempts:          retryMaxAttempts,
		RetryInitialIntervalMills: retryInitialIntervalMills,
		Capacity:                  capacity,
	}
}

func (fbs *FullJitterBackoffStrategy) Init() error {
	if fbs.RetryMaxAttempts == 0 {
		fbs.RetryMaxAttempts = utils.DefaultRetryMaxAttempts
	}
	if fbs.RetryInitialIntervalMills == 0 {
		fbs.RetryInitialIntervalMills = utils.DefaultRetryInitialIntervalMills
	}
	if fbs.Capacity == 0 {
		fbs.Capacity = utils.DefaultCapacity
	}
	return nil
}

func (fbs *FullJitterBackoffStrategy) GetWaitTimeExponential(retryTimes int) int64 {
	if retryTimes > fbs.RetryMaxAttempts {
		return -1
	}
	return int64(math.Min(float64(fbs.Capacity), math.Pow(2, float64(retryTimes))*float64(fbs.RetryInitialIntervalMills)))
}
