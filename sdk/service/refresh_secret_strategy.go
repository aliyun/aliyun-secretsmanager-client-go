package service

import (
	"encoding/json"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/logger"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/models"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/utils"
	"time"
)

// 刷新Secret的策略
type RefreshSecretStrategy interface {
	//  初始化刷新策略
	Init() error

	// 获取下一次secret刷新执行的时间
	GetNextExecuteTime(secretName string, ttl, offsetTimestamp int64) int64

	// 通过secret信息解析下一次secret刷新执行的时间
	ParseNextExecuteTime(cacheSecretInfo *models.CacheSecretInfo) int64

	// 根据凭据信息解析轮转时间间隔，单位MS
	ParseTTL(secretInfo *models.SecretInfo) int64

	// 关闭，释放资源
	Close() error
}

type defaultRefreshSecretStrategy struct {
	jsonTTLPropertyName string
}

func NewDefaultRefreshSecretStrategy(jsonTTLPropertyName string) RefreshSecretStrategy {
	return &defaultRefreshSecretStrategy{
		jsonTTLPropertyName: jsonTTLPropertyName,
	}
}

func (drs *defaultRefreshSecretStrategy) Init() error {
	return nil
}

func (drs *defaultRefreshSecretStrategy) GetNextExecuteTime(secretName string, ttl, offsetTimestamp int64) int64 {
	now := time.Now().UnixNano() / 1e6
	if ttl+offsetTimestamp > now {
		return ttl + offsetTimestamp
	} else {
		return now + ttl
	}
}

func (drs *defaultRefreshSecretStrategy) ParseNextExecuteTime(cacheSecretInfo *models.CacheSecretInfo) int64 {
	ttl := drs.ParseTTL(cacheSecretInfo.SecretInfo)
	if ttl <= 0 {
		return ttl
	}
	return drs.GetNextExecuteTime(cacheSecretInfo.SecretInfo.SecretName, ttl, cacheSecretInfo.RefreshTimestamp)
}

func (drs *defaultRefreshSecretStrategy) ParseTTL(secretInfo *models.SecretInfo) int64 {
	if drs.jsonTTLPropertyName == "" {
		return -1
	}
	var secretValue map[string]interface{}
	err := json.Unmarshal([]byte(secretInfo.SecretValue), &secretValue)
	if err != nil {
		logger.GetCommonLogger(utils.ModeName).Errorf("ParseTTL:%s", err.Error())
		return -1
	}
	if ttl, ok := secretValue[drs.jsonTTLPropertyName]; ok {
		if v, okk := ttl.(int64); okk {
			return v
		}
	}
	return -1
}

func (drs *defaultRefreshSecretStrategy) Close() error {
	return nil
}
