package sdk

import (
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/cache"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/logger"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/service"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/utils"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/assert"
)

var (
	accessKeyId     = os.Getenv("credentials_access_key_id")
	accessKeySecret = os.Getenv("credentials_access_secret")
)

func TestNewSecretCacheClient(t *testing.T) {
	client := NewSecretCacheClient()
	assert.Equal(t, defaultJsonTtlPropertyName, client.jsonTTLPropertyName)
	assert.Equal(t, utils.StageAcsCurrent, client.stage)
	assert.NotNil(t, client.secretTTLMap)
	assert.NotNil(t, client.scheduledMap)
	assert.Nil(t, client.secretManagerClient)
	assert.Nil(t, client.cacheHook)
	assert.Nil(t, client.refreshSecretStrategy)
	assert.Nil(t, client.cacheSecretStoreStrategy)
}

func TestSecretCacheClient_GetSecretInfo(t *testing.T) {
	regionId := "cn-hangzhou"
	jsonTTLPropertyName := "ttl"
	secretName := "cache_client"

	err := logger.RegisterLogger(utils.ModeName, logger.NewDefaultLogger(log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)))
	assert.Nil(t, err)

	client := &SecretManagerCacheClient{
		jsonTTLPropertyName:      jsonTTLPropertyName,
		stage:                    utils.StageAcsCurrent,
		secretManagerClient:      service.NewBaseSecretManagerClientBuilder().Standard().WithAccessKey(accessKeyId, accessKeySecret).WithRegion(regionId).Build(),
		cacheSecretStoreStrategy: cache.NewMemoryCacheSecretStoreStrategy(),
		refreshSecretStrategy:    service.NewDefaultRefreshSecretStrategy(jsonTTLPropertyName),
		cacheHook:                cache.NewDefaultSecretCacheHook(utils.StageAcsCurrent),
		secretTTLMap:             make(map[string]int64),
		scheduledMap:             cmap.New(),
		secretNameMtxMap:         make(map[string]*sync.Mutex),
	}

	err = client.Init()
	assert.Nil(t, err)

	info, err := client.GetSecretInfo(secretName)
	assert.Nil(t, err)
	assert.NotNil(t, info)
}

func TestSecretCacheClient_RefreshNow(t *testing.T) {
	regionId := "cn-hangzhou"
	jsonTTLPropertyName := "ttl"
	secretName := "cache_client"
	secretName1 := "cache_client_1"
	secretName2 := "cache_client_2"

	err := logger.RegisterLogger(utils.ModeName, logger.NewDefaultLogger(log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)))
	assert.Nil(t, err)

	client := &SecretManagerCacheClient{
		jsonTTLPropertyName:      jsonTTLPropertyName,
		stage:                    utils.StageAcsCurrent,
		secretManagerClient:      service.NewBaseSecretManagerClientBuilder().Standard().WithAccessKey(accessKeyId, accessKeySecret).WithRegion(regionId).Build(),
		cacheSecretStoreStrategy: cache.NewMemoryCacheSecretStoreStrategy(),
		refreshSecretStrategy:    service.NewDefaultRefreshSecretStrategy(jsonTTLPropertyName),
		cacheHook:                cache.NewDefaultSecretCacheHook(utils.StageAcsCurrent),
		secretTTLMap:             make(map[string]int64),
		scheduledMap:             cmap.New(),
		secretNameMtxMap:         make(map[string]*sync.Mutex),
	}

	client.secretTTLMap[secretName] = 10 * 1000
	client.secretTTLMap[secretName1] = 10 * 1000
	client.secretTTLMap[secretName2] = 10 * 1000

	err = client.Init()
	assert.Nil(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		i := i
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			var ok bool
			var err error
			if i == 0 {
				ok, err = client.RefreshNow(secretName)
			} else if i == 1 {
				ok, err = client.RefreshNow(secretName1)
			} else if i == 2 {
				ok, err = client.RefreshNow(secretName2)
			}
			assert.Nil(t, err)
			assert.Equal(t, true, ok)
			time.Sleep(1 * time.Minute)
		}(&wg)
	}
	wg.Wait()
}
