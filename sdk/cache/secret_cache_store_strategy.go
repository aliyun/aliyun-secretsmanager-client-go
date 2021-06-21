package cache

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/models"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/utils"
	mapset "github.com/deckarep/golang-set"
	cmap "github.com/orcaman/concurrent-map"
)

const (
	JsonFileNamePrefix = "stage_"
	JsonFileNameSuffix = ".json"
)

// SecretCacheStoreStrategy 缓存secret策略
type SecretCacheStoreStrategy interface {
	// 初始化凭据缓存
	Init() error

	// 缓存secret信息
	StoreSecret(cacheSecretInfo *models.CacheSecretInfo) error

	// 获取secret缓存信息
	GetCacheSecretInfo(secretName string) (*models.CacheSecretInfo, error)

	// 关闭，释放资源
	Close() error
}

type FileCacheSecretStoreStrategy struct {
	// 缓存凭据文件路径
	CacheSecretPath string
	// 首次启动时候是否允许从文件进行加载，true为允许
	ReloadOnStart bool
	//加解密过程中使用的salt
	Salt               string
	ReloadedSet        mapset.Set
	CacheSecretInfoMap cmap.ConcurrentMap
}

type MemoryCacheSecretStoreStrategy struct {
	CacheSecretInfoMap cmap.ConcurrentMap
}

func NewFileCacheSecretStoreStrategy(cacheSecretPath string, reloadOnStart bool, salt string) *FileCacheSecretStoreStrategy {
	return &FileCacheSecretStoreStrategy{
		CacheSecretPath:    cacheSecretPath,
		ReloadOnStart:      reloadOnStart,
		Salt:               salt,
		ReloadedSet:        mapset.NewSet(),
		CacheSecretInfoMap: cmap.New(),
	}
}

func NewMemoryCacheSecretStoreStrategy() *MemoryCacheSecretStoreStrategy {
	return &MemoryCacheSecretStoreStrategy{
		CacheSecretInfoMap: cmap.New(),
	}
}

func (fs *FileCacheSecretStoreStrategy) Init() error {
	if fs.CacheSecretPath == "" {
		fs.CacheSecretPath = "."
	}
	if fs.Salt == "" {
		return errors.New("the argument salt must not be empty")
	}
	return nil
}

func (fs *FileCacheSecretStoreStrategy) StoreSecret(cacheSecretInfo *models.CacheSecretInfo) error {
	memoryCacheSecretInfo := cacheSecretInfo.Clone()
	fileCacheSecretInfo := cacheSecretInfo.Clone()
	secretInfo := fileCacheSecretInfo.SecretInfo
	secretValue := secretInfo.SecretValue
	key, err := fs.generateRandomKey()
	if err != nil {
		return err
	}
	encryptedValue, err := fs.encryptSecretValue(secretValue, key)
	if err != nil {
		return err
	}
	secretInfo.SecretValue = encryptedValue
	fileName := strings.ToLower(JsonFileNamePrefix + cacheSecretInfo.Stage + JsonFileNameSuffix)
	cacheSecretPath := fs.CacheSecretPath + string(os.PathSeparator) + secretInfo.SecretName
	if utils.FileExists(cacheSecretPath, fileName) {
		err = utils.FileDelete(cacheSecretPath, fileName)
		if err != nil {
			return err
		}
	}
	err = utils.WriteJsonObject(cacheSecretPath, fileName, fileCacheSecretInfo)
	if err != nil {
		return err
	}
	fs.CacheSecretInfoMap.Set(secretInfo.SecretName, memoryCacheSecretInfo)
	fs.ReloadedSet.Add(cacheSecretInfo.SecretInfo.SecretName)
	return nil
}

func (fs *FileCacheSecretStoreStrategy) GetCacheSecretInfo(secretName string) (*models.CacheSecretInfo, error) {
	if !fs.ReloadOnStart && !fs.ReloadedSet.Contains(secretName) {
		return nil, errors.New(fmt.Sprintf("reloadedSet can't find [%s] key", secretName))
	}
	if cacheSecretInfoI, ok := fs.CacheSecretInfoMap.Get(secretName); ok {
		if cacheSecretInfo, okk := cacheSecretInfoI.(*models.CacheSecretInfo); okk {
			return cacheSecretInfo, nil
		} else {
			return nil, errors.New(fmt.Sprintf("CacheSecretInfoMap unknown type, expect: *models.CacheSecretInfo"))
		}
	}
	fileName := strings.ToLower(JsonFileNamePrefix + utils.StageAcsCurrent + JsonFileNameSuffix)
	cacheSecretPath := fs.CacheSecretPath + string(os.PathSeparator) + secretName
	var cacheSecretInfo *models.CacheSecretInfo
	err := utils.ReadJsonObject(cacheSecretPath, fileName, &cacheSecretInfo)
	if err != nil {
		return nil, err
	}
	secretInfo := cacheSecretInfo.SecretInfo
	secretValue, err := fs.decryptSecretValue(secretInfo.SecretValue)
	if err != nil {
		return nil, err
	}
	secretInfo.SecretValue = secretValue
	fs.CacheSecretInfoMap.Set(secretInfo.SecretName, cacheSecretInfo)
	return cacheSecretInfo, nil
}

func (fs *FileCacheSecretStoreStrategy) encryptSecretValue(secretValue string, key []byte) (string, error) {
	iv := make([]byte, utils.IvLength)
	_, err := rand.Read(iv)
	if err != nil {
		return "", err
	}
	encrypted := []byte(utils.Aes256CbcModeKey)
	encrypted = append(append(encrypted, key...), iv...)
	cipherData, err := utils.EncryptAes256Cbc([]byte(secretValue), key, iv, []byte(fs.Salt))
	if err != nil {
		return "", err
	}
	encrypted = append(encrypted, cipherData...)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func (fs *FileCacheSecretStoreStrategy) decryptSecretValue(secretValue string) (string, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(secretValue)
	if err != nil {
		return "", err
	}
	key := decodeBytes[len(utils.Aes256CbcModeKey) : len(utils.Aes256CbcModeKey)+utils.RandomKeyLength]
	iv := decodeBytes[len(utils.Aes256CbcModeKey)+utils.RandomKeyLength : len(utils.Aes256CbcModeKey)+utils.RandomKeyLength+utils.IvLength]
	cipherData := decodeBytes[len(utils.Aes256CbcModeKey)+utils.RandomKeyLength+utils.IvLength:]
	return utils.DecryptAes256Cbc(cipherData, key, iv, []byte(fs.Salt))
}

func (fs *FileCacheSecretStoreStrategy) generateRandomKey() ([]byte, error) {
	key := make([]byte, utils.RandomKeyLength)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (fs *FileCacheSecretStoreStrategy) Close() error {
	return nil
}

func (ms *MemoryCacheSecretStoreStrategy) Init() error {
	// do something
	return nil
}

func (ms *MemoryCacheSecretStoreStrategy) StoreSecret(cacheSecretInfo *models.CacheSecretInfo) error {
	ms.CacheSecretInfoMap.Set(cacheSecretInfo.SecretInfo.SecretName, cacheSecretInfo)
	return nil
}

func (ms *MemoryCacheSecretStoreStrategy) GetCacheSecretInfo(secretName string) (*models.CacheSecretInfo, error) {
	if cacheSecretInfoI, ok := ms.CacheSecretInfoMap.Get(secretName); ok {
		if cacheSecretInfo, okk := cacheSecretInfoI.(*models.CacheSecretInfo); okk {
			return cacheSecretInfo, nil
		} else {
			return nil, errors.New(fmt.Sprintf("invalid type [CacheSecretInfo]"))
		}
	}
	return nil, errors.New(fmt.Sprintf("invalid cacheSecretInfoMap key [%s]", secretName))
}

func (ms *MemoryCacheSecretStoreStrategy) Close() error {
	return nil
}
