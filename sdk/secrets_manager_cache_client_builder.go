package sdk

import (
	"log"
	"os"

	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/cache"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/logger"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/service"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/utils"
)

type SecretCacheClientBuilder struct {
	secretCacheClient *SecretManagerCacheClient
}

// 构建一个Secret Cache client
func NewClient() (*SecretManagerCacheClient, error) {
	builder := &SecretCacheClientBuilder{}
	return builder.Build()
}

// 根据指定的Secret Manager Client构建一个Cache client Builder
func NewSecretCacheClientBuilder(client service.SecretManagerClient) *SecretCacheClientBuilder {
	builder := &SecretCacheClientBuilder{}
	builder.buildSecretCacheClient()
	builder.secretCacheClient.secretManagerClient = client
	return builder
}

// 设定指定凭据名称的凭据TTL
func (scb *SecretCacheClientBuilder) WithSecretTTL(secretName string, ttl int64) *SecretCacheClientBuilder {
	scb.buildSecretCacheClient()
	scb.secretCacheClient.secretTTLMap[secretName] = ttl
	return scb
}

// 设定secret value解析TTL字段名称
func (scb *SecretCacheClientBuilder) WithParseJSONTTL(jsonTTLPropertyName string) *SecretCacheClientBuilder {
	scb.buildSecretCacheClient()
	scb.secretCacheClient.jsonTTLPropertyName = jsonTTLPropertyName
	return scb
}

// 设定secret刷新策略
func (scb *SecretCacheClientBuilder) WithRefreshSecretStrategy(refreshSecretStrategy service.RefreshSecretStrategy) *SecretCacheClientBuilder {
	scb.buildSecretCacheClient()
	scb.secretCacheClient.refreshSecretStrategy = refreshSecretStrategy
	return scb
}

// 设定secret缓存策略
func (scb *SecretCacheClientBuilder) WithCacheSecretStrategy(cacheSecretStrategy cache.SecretCacheStoreStrategy) *SecretCacheClientBuilder {
	scb.buildSecretCacheClient()
	scb.secretCacheClient.cacheSecretStoreStrategy = cacheSecretStrategy
	return scb
}

// 指定凭据Version stage
func (scb *SecretCacheClientBuilder) WithCacheStage(stage string) *SecretCacheClientBuilder {
	scb.buildSecretCacheClient()
	scb.secretCacheClient.stage = stage
	return scb
}

// 指定凭据Cache Hook
func (scb *SecretCacheClientBuilder) WithSecretCacheHook(hook cache.SecretCacheHook) *SecretCacheClientBuilder {
	scb.buildSecretCacheClient()
	scb.secretCacheClient.cacheHook = hook
	return scb
}

// 指定输出日志
func (scb *SecretCacheClientBuilder) WithLogger(l logger.Wrapper) *SecretCacheClientBuilder {
	err := logger.RegisterLogger(utils.ModeName, l)
	if err != nil {
		logger.GetCommonLogger("").Errorf(err.Error())
	}
	return scb
}

// 构建Cache Client对象
func (scb *SecretCacheClientBuilder) Build() (*SecretManagerCacheClient, error) {
	scb.buildSecretCacheClient()
	if !logger.IsRegistered(utils.ModeName) {
		err := logger.RegisterLogger(utils.ModeName, logger.NewDefaultLogger(log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)))
		if err != nil {
			return nil, err
		}
	}
	if scb.secretCacheClient.secretManagerClient == nil {
		scb.secretCacheClient.secretManagerClient = service.NewBaseSecretManagerClientBuilder().Standard().Build()
	}
	if scb.secretCacheClient.cacheSecretStoreStrategy == nil {
		scb.secretCacheClient.cacheSecretStoreStrategy = cache.NewMemoryCacheSecretStoreStrategy()
	}
	if scb.secretCacheClient.refreshSecretStrategy == nil {
		scb.secretCacheClient.refreshSecretStrategy = service.NewDefaultRefreshSecretStrategy(scb.secretCacheClient.jsonTTLPropertyName)
	}
	if scb.secretCacheClient.cacheHook == nil {
		scb.secretCacheClient.cacheHook = cache.NewDefaultSecretCacheHook(scb.secretCacheClient.stage)
	}
	err := scb.secretCacheClient.Init()
	if err != nil {
		return nil, err
	}
	logger.GetCommonLogger(utils.ModeName).Infof("SecretCacheClientBuilder build success")
	return scb.secretCacheClient, nil
}

func (scb *SecretCacheClientBuilder) buildSecretCacheClient() {
	if scb.secretCacheClient == nil {
		scb.secretCacheClient = NewSecretCacheClient()
	}
}
