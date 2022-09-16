package sdk

import (
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/logger"
	"log"
	"os"
	"testing"
	"time"

	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/cache"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/service"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/utils"

	"github.com/stretchr/testify/assert"
)

func TestSecretCacheClientBuilder_Build(t *testing.T) {
	regionId := "cn-hangzhou"
	cacheSecretPath := "secrets"
	salt := "1234abcd"
	jsonTTLPropertyName := "ttl"
	secretName := "cache_client"

	client, err := NewSecretCacheClientBuilder(service.NewDefaultSecretManagerClientBuilder().WithAccessKey(accessKeyId, accessKeySecret).WithRegion(regionId).WithBackoffStrategy(&service.FullJitterBackoffStrategy{RetryMaxAttempts: 3, RetryInitialIntervalMills: 2000, Capacity: 10000}).Build()).WithCacheSecretStrategy(cache.NewFileCacheSecretStoreStrategy(cacheSecretPath, true, salt)).WithRefreshSecretStrategy(service.NewDefaultRefreshSecretStrategy(jsonTTLPropertyName)).WithCacheStage(utils.StageAcsCurrent).WithLogger(logger.NewDefaultLogger(log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile))).WithSecretTTL(secretName, 1*60*1000).Build()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	info, err := client.GetSecretInfo(secretName)
	println("info :", info.SecretValue)
	time.Sleep(1000000 * time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, info)
}

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	assert.Nil(t, err)
	secretInfo, err := client.GetSecretInfo("cache_client")
	assert.Nil(t, err)
	println("secretInfo:", secretInfo.SecretValue)
}
func TestNewSecretCacheClientBuilder(t *testing.T) {
	client, _ := NewSecretCacheClientBuilder(service.NewDefaultSecretManagerClientBuilder().WithAccessKey(accessKeyId, accessKeySecret).Build()).Build()
	secretInfo, err := client.GetSecretInfo("cache_client")
	assert.Nil(t, err)
	println("secretInfo:", secretInfo.SecretValue)
}
