package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/logger"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/models"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/utils"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/kms"
)

const (
	// 默认请求等待时间
	RequestWaitingTime = 10 * 60 * 1000
)

type SecretManagerClient interface {
	// 初始化Client
	Init() error

	// 获取指定凭据信息
	GetSecretValue(req *kms.GetSecretValueRequest) (*kms.GetSecretValueResponse, error)

	// 关闭Client
	Close() error
}

type baseSecretManagerClientBuilder struct {
}

type defaultSecretManagerClientBuilder struct {
	baseSecretManagerClientBuilder
	regionInfos     []*models.RegionInfo
	credential      auth.Credential
	backoffStrategy BackoffStrategy
	signer          auth.Signer
}

type defaultSecretManagerClient struct {
	*defaultSecretManagerClientBuilder
	clientMap map[string]*kms.Client
	clientMtx sync.Mutex
}

func NewBaseSecretManagerClientBuilder() *baseSecretManagerClientBuilder {
	return &baseSecretManagerClientBuilder{}
}

func NewDefaultSecretManagerClientBuilder() *defaultSecretManagerClientBuilder {
	return &defaultSecretManagerClientBuilder{}
}

func (base *baseSecretManagerClientBuilder) Standard() *defaultSecretManagerClientBuilder {
	return &defaultSecretManagerClientBuilder{}
}

func (dsb *defaultSecretManagerClientBuilder) WithToken(tokenId, token string) *defaultSecretManagerClientBuilder {
	dsb.credential = utils.CredentialsWithToken(tokenId, token)
	return dsb
}

func (dsb *defaultSecretManagerClientBuilder) WithAccessKey(accessKeyId, accessKeySecret string) *defaultSecretManagerClientBuilder {
	dsb.credential = utils.CredentialsWithAccessKey(accessKeyId, accessKeySecret)
	return dsb
}

func (dsb *defaultSecretManagerClientBuilder) WithCredentials(credential auth.Credential) *defaultSecretManagerClientBuilder {
	dsb.credential = credential
	return dsb
}

// 指定多个调用地域Id
func (dsb *defaultSecretManagerClientBuilder) WithRegion(regionIds ...string) *defaultSecretManagerClientBuilder {
	for _, regionId := range regionIds {
		dsb.AddRegionInfo(&models.RegionInfo{RegionId: regionId})
	}
	return dsb
}

// 指定调用地域信息
func (dsb *defaultSecretManagerClientBuilder) AddRegionInfo(regionInfo *models.RegionInfo) *defaultSecretManagerClientBuilder {
	dsb.regionInfos = append(dsb.regionInfos, regionInfo)
	return dsb
}

func (dsb *defaultSecretManagerClientBuilder) WithBackoffStrategy(backoffStrategy BackoffStrategy) *defaultSecretManagerClientBuilder {
	dsb.backoffStrategy = backoffStrategy
	return dsb
}

func (dsb *defaultSecretManagerClientBuilder) Build() SecretManagerClient {
	return &defaultSecretManagerClient{
		defaultSecretManagerClientBuilder: dsb,
		clientMap:                         make(map[string]*kms.Client),
	}
}

// 指定调用地域Id
func (dsb *defaultSecretManagerClientBuilder) addRegion(regionId string) *defaultSecretManagerClientBuilder {
	return dsb.AddRegionInfo(&models.RegionInfo{RegionId: regionId})
}

func (dsb *defaultSecretManagerClientBuilder) sortRegionInfos(regionInfos []*models.RegionInfo) []*models.RegionInfo {
	var regionInfoResp []*models.RegionInfo
	var regionInfoExtends []*models.RegionInfoExtend
	var wg sync.WaitGroup
	for _, regionInfo := range regionInfos {
		wg.Add(1)
		regionInfo := regionInfo
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			var pingDelay float64
			regionInfoExtend := &models.RegionInfoExtend{
				RegionInfo: regionInfo,
			}
			if regionInfo.Endpoint != "" {
				pingDelay = utils.Ping(regionInfo.Endpoint)
			} else if regionInfo.Vpc {
				pingDelay = utils.Ping(utils.GetVpcEndpoint(regionInfo.RegionId))
			} else {
				pingDelay = utils.Ping(utils.GetEndpoint(regionInfo.RegionId))
			}
			if pingDelay >= 0 {
				regionInfoExtend.Escaped = pingDelay
			} else {
				regionInfoExtend.Escaped = math.MaxFloat64
			}
			regionInfoExtend.Reachable = pingDelay >= 0
			regionInfoExtends = append(regionInfoExtends, regionInfoExtend)
		}(&wg)
	}
	wg.Wait()
	// 注意>go1.8才有sort.Slice
	sort.Slice(regionInfoExtends, func(i, j int) bool {
		return regionInfoExtends[i].Escaped < regionInfoExtends[j].Escaped
	})
	for _, regionInfoExtend := range regionInfoExtends {
		regionInfoResp = append(regionInfoResp, regionInfoExtend.RegionInfo)
	}
	return regionInfoResp
}

func (dmc *defaultSecretManagerClient) Init() error {
	err := dmc.initProperties()
	if err != nil {
		return err
	}
	err = dmc.initEnv()
	if err != nil {
		return err
	}
	credential, yes := dmc.credential.(*models.ClientKeyCredential)
	if yes {
		dmc.signer = credential.Signer
		dmc.credential = credential.Credential
	}
	UserAgentManager.RegisterUserAgent(utils.UserAgentOfSecretsManagerGolang, 0, utils.ProjectVersion)
	if dmc.backoffStrategy == nil {
		dmc.backoffStrategy = &FullJitterBackoffStrategy{}
	}
	err = dmc.backoffStrategy.Init()
	if err != nil {
		return err
	}
	dmc.regionInfos = dmc.sortRegionInfos(dmc.regionInfos)
	return nil
}

func (dmc *defaultSecretManagerClient) GetSecretValue(req *kms.GetSecretValueRequest) (*kms.GetSecretValueResponse, error) {
	var results []*kms.GetSecretValueResponse
	var errs []error
	var wg sync.WaitGroup
	finished := int32(len(dmc.regionInfos))
	for i, regionInfo := range dmc.regionInfos {
		if i == 0 {
			resp, err := dmc.getSecretValue(regionInfo, req)
			if err == nil {
				return resp, nil
			}
			logger.GetCommonLogger(utils.ModeName).Errorf("action:getSecretValue, regionInfo:%+v, %+v", regionInfo, err)
			if !utils.JudgeNeedRecoveryException(err) {
				return nil, err
			}
			wg.Add(1)
		}
		regionInfo := regionInfo
		// kms sdk判断domain为空以后通过region获取endpoint写到req结构domain字段里
		// 后续domain有值以后sdk不再修改了，所以会导致所有region使用同一个endpoint
		// 因此，要想访问不同的region，这里需要重新创建request
		request := kms.CreateGetSecretValueRequest()
		request.Scheme = "https"
		request.SecretName = req.SecretName
		request.VersionStage = req.VersionStage
		request.FetchExtendedConfig = requests.NewBoolean(true)
		go func(wg *sync.WaitGroup, finished *int32) {
			if resp, err := dmc.retryGetSecretValue(request, regionInfo); err == nil {
				results = append(results, resp)
				wg.Done()
			} else {
				errs = append(errs, err)
				for {
					val := atomic.LoadInt32(finished)
					if atomic.CompareAndSwapInt32(finished, val, val-1) {
						break
					}
				}
				if atomic.LoadInt32(finished) == 0 {
					wg.Done()
				}
			}
		}(&wg, &finished)
	}
	dmc.waitTimeout(&wg, time.Duration(RequestWaitingTime)*time.Millisecond)
	if len(results) == 0 {
		var errStr string
		for _, err := range errs {
			errStr += fmt.Sprintf("%+v;", err)
		}
		return nil, errors.New(fmt.Sprintf("action:retryGetSecretValueTask:%s", errStr))
	}
	return results[0], nil
}

func (dmc *defaultSecretManagerClient) Close() error {
	for _, client := range dmc.clientMap {
		client.Shutdown()
	}
	return nil
}

func (dmc *defaultSecretManagerClient) getSecretValue(regionInfo *models.RegionInfo, req *kms.GetSecretValueRequest) (*kms.GetSecretValueResponse, error) {
	client, err := dmc.getClient(regionInfo)
	if err != nil {
		return nil, err
	}
	response := kms.CreateGetSecretValueResponse()
	return response, client.DoActionWithSigner(req, response, dmc.signer)
}

func (dmc *defaultSecretManagerClient) getClient(regionInfo *models.RegionInfo) (*kms.Client, error) {
	if client, ok := dmc.clientMap[regionInfo.RegionId]; ok {
		return client, nil
	}
	dmc.clientMtx.Lock()
	defer dmc.clientMtx.Unlock()
	if client, ok := dmc.clientMap[regionInfo.RegionId]; ok {
		return client, nil
	}
	config := sdk.NewConfig()
	kmsClient, err := kms.NewClientWithOptions(regionInfo.RegionId, config, dmc.credential)
	if err != nil {
		return nil, err
	}
	if regionInfo.Endpoint != "" {
		kmsClient.Domain = regionInfo.Endpoint
	} else if regionInfo.Vpc {
		kmsClient.Domain = utils.GetVpcEndpoint(regionInfo.RegionId)
	}
	kmsClient.SetHTTPSInsecure(true)
	kmsClient.AppendUserAgent(UserAgentManager.GetUserAgent(), UserAgentManager.GetProjectVersion())
	dmc.clientMap[regionInfo.RegionId] = kmsClient
	return kmsClient, nil
}

func (dmc *defaultSecretManagerClient) initProperties() error {
	if dmc.credential == nil {
		credentialsProperties, err := utils.LoadCredentialsProperties("")
		if err != nil {
			return err
		}
		if credentialsProperties != nil {
			dmc.credential = credentialsProperties.Credential
			dmc.regionInfos = credentialsProperties.RegionInfoSlice
		}
	}
	return nil
}
func (dmc *defaultSecretManagerClient) initEnv() error {
	if dmc.credential == nil {
		credentialsType := os.Getenv(utils.EnvCredentialsTypeKey)
		if credentialsType == "" {
			return errors.New(fmt.Sprintf("env param[%s] is required", utils.EnvCredentialsTypeKey))
		}
		accessKeyId := os.Getenv(utils.EnvCredentialsAccessKeyIdKey)
		if accessKeyId == "" {
			return errors.New(fmt.Sprintf("env param[%s] is required", utils.EnvCredentialsAccessKeyIdKey))
		}
		accessSecret := os.Getenv(utils.EnvCredentialsAccessSecretKey)
		if accessSecret == "" {
			return errors.New(fmt.Sprintf("env param[%s] is required", utils.EnvCredentialsAccessSecretKey))
		}
		switch credentialsType {
		case "ak":
			dmc.credential = utils.CredentialsWithAccessKey(accessKeyId, accessSecret)
			break
		case "token":
			tokenId := os.Getenv(utils.EnvCredentialsAccessTokenIdKey)
			if tokenId == "" {
				return errors.New(fmt.Sprintf("env param[%s] is required", utils.EnvCredentialsAccessTokenIdKey))
			}
			token := os.Getenv(utils.EnvCredentialsAccessTokenKey)
			if token == "" {
				return errors.New(fmt.Sprintf("env param[%s] is required", utils.EnvCredentialsAccessTokenKey))
			}
			dmc.credential = utils.CredentialsWithToken(tokenId, token)
			break
		case "sts":
		case "ram_role":
			roleSessionName := os.Getenv(utils.EnvCredentialsRoleSessionNameKey)
			if roleSessionName == "" {
				return errors.New(fmt.Sprintf("env param[%s] is required", utils.EnvCredentialsRoleSessionNameKey))
			}
			roleArn := os.Getenv(utils.EnvCredentialsRoleArnKey)
			if roleArn == "" {
				return errors.New(fmt.Sprintf("env param[%s] is required", utils.EnvCredentialsRoleArnKey))
			}
			policy := os.Getenv(utils.EnvCredentialsPolicyKey)
			dmc.credential = utils.CredentialsWithRamRoleArnOrSts(accessKeyId, accessSecret, roleSessionName, roleArn, policy)
			break
		case "ecs_ram_role":
			roleName := os.Getenv(utils.EnvCredentialsRoleNameKey)
			if roleName == "" {
				return errors.New(fmt.Sprintf("env param[%s] is required", utils.EnvCredentialsRoleNameKey))
			}
			dmc.credential = utils.CredentialsWithEcsRamRole(roleName)
			break
		case "client_key":
			privateKeyPath := os.Getenv(utils.EnvClientKeyPrivateKeyPathNameKey)
			if privateKeyPath == "" {
				return errors.New(fmt.Sprintf("env param[%s] is required", utils.EnvClientKeyPrivateKeyPathNameKey))
			}
			password, err := utils.GetPassword(nil)
			if err != nil {
				return err
			}
			credential, signer, err := utils.LoadRsaKeyPairCredentialAndClientKeySigner(privateKeyPath, password)
			if err != nil {
				return err
			}
			dmc.credential = models.NewClientKeyCredential(signer, credential)
			break
		default:
			return errors.New(fmt.Sprintf("env param[%s] is illegal", utils.EnvCredentialsTypeKey))
		}
		if dmc.credential != nil {
			regionInfoJson := os.Getenv(utils.EnvCacheClientRegionIdKey)
			if regionInfoJson == "" {
				return errors.New(fmt.Sprintf("env param[%s] is required", utils.EnvCacheClientRegionIdKey))
			}
			var regionInfoList []map[string]interface{}
			err := json.Unmarshal([]byte(regionInfoJson), &regionInfoList)
			if err != nil {
				return err
			}
			for _, regionInfoMap := range regionInfoList {
				regionId, err := utils.ParseString(regionInfoMap[utils.EnvRegionRegionIdNameKey])
				if err != nil {
					return err
				}
				endpoint, err := utils.ParseString(regionInfoMap[utils.EnvRegionEndpointNameKey])
				if err != nil {
					return err
				}
				vpc, err := utils.ParseBool(regionInfoMap[utils.EnvRegionVpcNameKey])
				if err != nil {
					return err
				}
				dmc.regionInfos = append(dmc.regionInfos, &models.RegionInfo{RegionId: regionId, Endpoint: endpoint, Vpc: vpc})
			}
		}
	}
	return nil
}

func (dmc *defaultSecretManagerClient) retryGetSecretValue(req *kms.GetSecretValueRequest, regionInfo *models.RegionInfo) (*kms.GetSecretValueResponse, error) {
	retryTimes := 0
	for {
		// todo: 这里需要增加退出机制防止无限重试

		waitTimeExponential := dmc.backoffStrategy.GetWaitTimeExponential(retryTimes)
		if waitTimeExponential < 0 {
			return nil, errors.New(fmt.Sprintf("Times limit exceeded"))
		}

		time.Sleep(time.Duration(waitTimeExponential) * time.Millisecond)

		resp, err := dmc.getSecretValue(regionInfo, req)
		if err == nil {
			return resp, nil
		}
		logger.GetCommonLogger(utils.ModeName).Errorf("action:getSecretValue, regionInfo:%+v, %+v", regionInfo, err)
		if !utils.JudgeNeedRecoveryException(err) {
			return nil, err
		}
		retryTimes += 1
	}
}

func (dmc *defaultSecretManagerClient) waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()
	select {
	case <-done:
		return false
	case <-time.After(timeout):
		return true
	}
}
