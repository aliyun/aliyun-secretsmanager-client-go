package cache

import (
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/models"
	"time"
)

type SecretCacheHook interface {
	// 初始化Hook
	Init() error

	// 将secret对象转化为Cache secret对象
	Put(o *models.SecretInfo) (*models.CacheSecretInfo, error)

	// 将Cache secret对象转化为secret对象
	Get(cachedObject *models.CacheSecretInfo) (*models.SecretInfo, error)

	// RecoveryGetSecret
	RecoveryGetSecret(secretName string) (*models.SecretInfo, error)

	// 关闭，释放资源
	Close() error
}

func NewDefaultSecretCacheHook(stage string) SecretCacheHook {
	return &defaultSecretCacheHook{stage: stage}
}

// 默认hook,不做特殊操作
type defaultSecretCacheHook struct {
	// 缓存的凭据Version Stage
	stage string
}

func (dch *defaultSecretCacheHook) Init() error {
	// do something
	return nil
}

func (dch *defaultSecretCacheHook) Put(o *models.SecretInfo) (*models.CacheSecretInfo, error) {
	return &models.CacheSecretInfo{
		SecretInfo:       o,
		Stage:            dch.stage,
		RefreshTimestamp: time.Now().UnixNano() / 1e6,
	}, nil
}

func (dch *defaultSecretCacheHook) Get(cachedObject *models.CacheSecretInfo) (*models.SecretInfo, error) {
	return cachedObject.SecretInfo, nil
}

func (dch *defaultSecretCacheHook) RecoveryGetSecret(secretName string) (*models.SecretInfo, error) {
	return nil, nil
}

func (dch *defaultSecretCacheHook) Close() error {
	return nil
}
