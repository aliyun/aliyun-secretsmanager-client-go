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

	"github.com/alibabacloud-go/tea/tea"
	transfersdk "github.com/aliyun/alibabacloud-dkms-transfer-go-sdk/sdk"
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
	regionInfos      []*models.RegionInfo
	credential       auth.Credential
	backoffStrategy  BackoffStrategy
	signer           auth.Signer
	dKmsConfigsMap   map[*models.RegionInfo]*models.DkmsConfig
	customConfigFile string
}

type defaultSecretManagerClient struct {
	*defaultSecretManagerClientBuilder
	clientMap map[*models.RegionInfo]interface{}
	clientMtx sync.Mutex
}

func NewBaseSecretManagerClientBuilder() *baseSecretManagerClientBuilder {
	return &baseSecretManagerClientBuilder{}
}

func NewDefaultSecretManagerClientBuilder() *defaultSecretManagerClientBuilder {
	return &defaultSecretManagerClientBuilder{
		dKmsConfigsMap: make(map[*models.RegionInfo]*models.DkmsConfig),
	}
}

func (base *baseSecretManagerClientBuilder) Standard() *defaultSecretManagerClientBuilder {
	return &defaultSecretManagerClientBuilder{
		dKmsConfigsMap: make(map[*models.RegionInfo]*models.DkmsConfig),
	}
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

func (dsb *defaultSecretManagerClientBuilder) AddDkmsConfig(dkmsConfig *models.DkmsConfig) *defaultSecretManagerClientBuilder {
	regionInfo := &models.RegionInfo{
		KmsType:  utils.DkmsType,
		RegionId: tea.StringValue(dkmsConfig.RegionId),
		Endpoint: tea.StringValue(dkmsConfig.Endpoint),
	}
	dsb.dKmsConfigsMap[regionInfo] = dkmsConfig
	dsb.AddRegionInfo(regionInfo)
	return dsb
}

func (dsb *defaultSecretManagerClientBuilder) WithCustomConfigFile(customConfigFile string) *defaultSecretManagerClientBuilder {
	dsb.customConfigFile = customConfigFile
	return dsb
}

func (dsb *defaultSecretManagerClientBuilder) Build() SecretManagerClient {
	return &defaultSecretManagerClient{
		defaultSecretManagerClientBuilder: dsb,
		clientMap:                         make(map[*models.RegionInfo]interface{}),
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
	err := dmc.initFromConfigFile()
	if err != nil {
		return err
	}
	err = dmc.initFromEnv()
	if err != nil {
		return err
	}
	if len(dmc.regionInfos) == 0 {
		return errors.New("the param[regionInfo] is needed")
	}
	if len(dmc.dKmsConfigsMap) == 0 && dmc.credential == nil {
		return errors.New("the param[credentials] is needed")
	}
	credential, ok := dmc.credential.(*models.ClientKeyCredential)
	if ok {
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
	if dmc.regionInfos!=nil && len(dmc.regionInfos)>1{
		dmc.regionInfos = dmc.sortRegionInfos(dmc.regionInfos)
	}
	return nil
}

func (dmc *defaultSecretManagerClient) GetSecretValue(req *kms.GetSecretValueRequest) (*kms.GetSecretValueResponse, error) {
	var results []*kms.GetSecretValueResponse
	var errs []error
	var wg sync.WaitGroup
	finished := int32(len(dmc.regionInfos))
	retryEnd := make(chan struct{})
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
		go func(wg *sync.WaitGroup, finished *int32, retryEnd <-chan struct{}) {
			if resp, err := dmc.retryGetSecretValue(request, regionInfo, retryEnd); err == nil {
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
		}(&wg, &finished, retryEnd)
	}
	dmc.waitTimeout(&wg, time.Duration(RequestWaitingTime)*time.Millisecond)
	close(retryEnd)
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
		switch c := client.(type) {
		case *kms.Client:
			c.Shutdown()
		case *transfersdk.KmsTransferClient:
			c.Shutdown()
		}
	}
	return nil
}

func (dmc *defaultSecretManagerClient) getSecretValue(regionInfo *models.RegionInfo, req *kms.GetSecretValueRequest) (*kms.GetSecretValueResponse, error) {
	client, err := dmc.getClient(regionInfo)
	if err != nil {
		return nil, err
	}
	switch c := client.(type) {
	case *kms.Client:
		response := kms.CreateGetSecretValueResponse()
		return response, utils.TransferErrorToClientError(c.DoActionWithSigner(req, response, dmc.signer))
	case *transfersdk.KmsTransferClient:
		response, err := c.GetSecretValue(req)
		return response, utils.TransferErrorToClientError(err)
	}
	return nil, errors.New("getClient unknown kms client type")
}

func (dmc *defaultSecretManagerClient) getClient(regionInfo *models.RegionInfo) (interface{}, error) {
	if client, ok := dmc.clientMap[regionInfo]; ok {
		return client, nil
	}
	dmc.clientMtx.Lock()
	defer dmc.clientMtx.Unlock()
	if client, ok := dmc.clientMap[regionInfo]; ok {
		return client, nil
	}
	if regionInfo.KmsType == utils.DkmsType {
		kmsTransferClient, err := dmc.buildDKmsTransferClient(regionInfo)
		if err != nil {
			return nil, err
		}
		dmc.clientMap[regionInfo] = kmsTransferClient
	} else {
		kmsClient, err := dmc.buildKmsClient(regionInfo)
		if err != nil {
			return nil, err
		}
		dmc.clientMap[regionInfo] = kmsClient
	}
	return dmc.clientMap[regionInfo], nil
}

func (dmc *defaultSecretManagerClient) buildKmsClient(regionInfo *models.RegionInfo) (*kms.Client, error) {
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
	return kmsClient, nil
}

func (dmc *defaultSecretManagerClient) buildDKmsTransferClient(regionInfo *models.RegionInfo) (*transfersdk.KmsTransferClient, error) {
	dkmsConfig, ok := dmc.dKmsConfigsMap[regionInfo]
	if !ok {
		return nil, errors.New("unrecognized regionId")
	}
	config := dkmsConfig.Config
	config.RegionId = tea.String(regionInfo.RegionId)
	config.Endpoint = tea.String(regionInfo.Endpoint)
	config.Password = dkmsConfig.Password
	kmsTransferClient, err := transfersdk.NewClientWithAccessKey(regionInfo.RegionId, utils.PretendAk, utils.PretendSk, config)
	if err != nil {
		return nil, err
	}
	kmsTransferClient.SetHTTPSInsecure(dkmsConfig.IgnoreSslCerts)
	if !dkmsConfig.IgnoreSslCerts {
		kmsTransferClient.SetVerify(dkmsConfig.CaCert)
	}
	kmsTransferClient.AppendUserAgent(UserAgentManager.GetUserAgent(), UserAgentManager.GetProjectVersion())
	return kmsTransferClient, nil
}

func (dmc *defaultSecretManagerClient) initFromConfigFile() error {
	credentialsProperties, err := utils.LoadCredentialsProperties(dmc.customConfigFile)
	if err != nil {
		return err
	}
	if credentialsProperties != nil {
		dmc.credential = credentialsProperties.Credential
		dmc.regionInfos = append(dmc.regionInfos, credentialsProperties.RegionInfoSlice...)
		for regionInfo, dkmsConfig := range credentialsProperties.DkmsConfigsMap {
			dmc.dKmsConfigsMap[regionInfo] = dkmsConfig
		}
	}
	return nil
}

func (dmc *defaultSecretManagerClient) initFromEnv() error {
	err := dmc.initCredentialFromEnv()
	if err != nil {
		return err
	}
	err = dmc.initDkmsInstancesFromEnv()
	if err != nil {
		return err
	}
	err = dmc.initKmsRegionsFromEnv()
	if err != nil {
		return err
	}
	return nil
}

func (dmc *defaultSecretManagerClient) initCredentialFromEnv() error {
	credentialsType := os.Getenv(utils.EnvCredentialsTypeKey)
	if credentialsType == "" {
		return nil
	}
	switch credentialsType {
	case "ak":
		accessKeyId := os.Getenv(utils.EnvCredentialsAccessKeyIdKey)
		err := dmc.checkEnvParam(accessKeyId, utils.EnvCredentialsAccessKeyIdKey)
		if err != nil {
			return err
		}
		accessSecret := os.Getenv(utils.EnvCredentialsAccessSecretKey)
		err = dmc.checkEnvParam(accessSecret, utils.EnvCredentialsAccessSecretKey)
		if err != nil {
			return err
		}
		dmc.credential = utils.CredentialsWithAccessKey(accessKeyId, accessSecret)
	case "token":
		tokenId := os.Getenv(utils.EnvCredentialsAccessTokenIdKey)
		err := dmc.checkEnvParam(tokenId, utils.EnvCredentialsAccessTokenIdKey)
		if err != nil {
			return err
		}
		token := os.Getenv(utils.EnvCredentialsAccessTokenKey)
		err = dmc.checkEnvParam(token, utils.EnvCredentialsAccessTokenKey)
		if err != nil {
			return err
		}
		dmc.credential = utils.CredentialsWithToken(tokenId, token)
	case "sts", "ram_role":
		accessKeyId := os.Getenv(utils.EnvCredentialsAccessKeyIdKey)
		err := dmc.checkEnvParam(accessKeyId, utils.EnvCredentialsAccessKeyIdKey)
		if err != nil {
			return err
		}
		accessSecret := os.Getenv(utils.EnvCredentialsAccessSecretKey)
		err = dmc.checkEnvParam(accessSecret, utils.EnvCredentialsAccessSecretKey)
		if err != nil {
			return err
		}
		roleSessionName := os.Getenv(utils.EnvCredentialsRoleSessionNameKey)
		err = dmc.checkEnvParam(roleSessionName, utils.EnvCredentialsRoleSessionNameKey)
		if err != nil {
			return err
		}
		roleArn := os.Getenv(utils.EnvCredentialsRoleArnKey)
		err = dmc.checkEnvParam(roleArn, utils.EnvCredentialsRoleArnKey)
		if err != nil {
			return err
		}
		policy := os.Getenv(utils.EnvCredentialsPolicyKey)
		dmc.credential = utils.CredentialsWithRamRoleArnOrSts(accessKeyId, accessSecret, roleSessionName, roleArn, policy)
	case "ecs_ram_role":
		roleName := os.Getenv(utils.EnvCredentialsRoleNameKey)
		err := dmc.checkEnvParam(roleName, utils.EnvCredentialsRoleNameKey)
		if err != nil {
			return err
		}
		dmc.credential = utils.CredentialsWithEcsRamRole(roleName)
	case "client_key":
		privateKeyPath := os.Getenv(utils.EnvClientKeyPrivateKeyPathNameKey)
		err := dmc.checkEnvParam(privateKeyPath, utils.EnvClientKeyPrivateKeyPathNameKey)
		if err != nil {
			return err
		}
		password, err := utils.GetPassword(nil, utils.EnvClientKeyPasswordFromEnvVariableNameKey, utils.PropertiesClientKeyPasswordFromFilePathName)
		if err != nil {
			return err
		}
		credential, signer, err := utils.LoadRsaKeyPairCredentialAndClientKeySigner(privateKeyPath, password)
		if err != nil {
			return err
		}
		dmc.credential = models.NewClientKeyCredential(signer, credential)
	default:
		return errors.New(fmt.Sprintf("env param[%s] is illegal", utils.EnvCredentialsTypeKey))
	}
	return nil
}

func (dmc *defaultSecretManagerClient) initKmsRegionsFromEnv() error {
	regionInfosJson := os.Getenv(utils.EnvCacheClientRegionIdKey)
	if regionInfosJson == "" {
		return nil
	}
	var regionInfos []map[string]interface{}
	err := json.Unmarshal([]byte(regionInfosJson), &regionInfos)
	if err != nil {
		return errors.New(fmt.Sprintf("env param[%s] is illegal, err: %v", utils.EnvCacheClientRegionIdKey, err))
	}
	for _, regionInfoMap := range regionInfos {
		regionId, err := utils.ParseString(regionInfoMap[utils.EnvRegionRegionIdNameKey])
		if err != nil {
			return err
		}
		endpoint, err := utils.ParseString(regionInfoMap[utils.EnvRegionEndpointNameKey])
		if err != nil {
			return err
		}
		var vpc bool
		if regionInfoMap[utils.EnvRegionVpcNameKey] == "" {
			vpc = false
		} else {
			vpc, err = utils.ParseBool(regionInfoMap[utils.EnvRegionVpcNameKey])
			if err != nil {
				return err
			}
		}
		dmc.regionInfos = append(dmc.regionInfos, models.NewRegionInfoWithKmsType(regionId, vpc, endpoint, utils.KmsType))
	}
	return nil
}

func (dmc *defaultSecretManagerClient) initDkmsInstancesFromEnv() error {
	configJson := os.Getenv(utils.CacheClientDkmsConfigInfoKey)
	if configJson == "" {
		return nil
	}
	var dkmsConfigs []*models.DkmsConfig
	err := json.Unmarshal([]byte(configJson), &dkmsConfigs)
	if err != nil {
		return errors.New(fmt.Sprintf("env param[%s] is illegal, err:%v", utils.CacheClientDkmsConfigInfoKey, err))
	}
	for _, dkmsConfig := range dkmsConfigs {
		if tea.StringValue(dkmsConfig.RegionId) == "" || tea.StringValue(dkmsConfig.Endpoint) == "" || tea.StringValue(dkmsConfig.ClientKeyFile) == "" {
			return errors.New("init env fail,cause of cache_client_dkms_config_info param[regionId or endpoint or clientKeyFile] is empty")
		}
		var password string
		if dkmsConfig.PasswordFromFilePath != "" {
			password, err = utils.ReadPasswordFile(dkmsConfig.PasswordFromFilePath)
		} else {
			password, err = utils.GetPassword(nil, dkmsConfig.PasswordFromEnvVariable, dkmsConfig.PasswordFromFilePathName)
		}
		if err != nil {
			return err
		}
		dkmsConfig.Password = tea.String(password)
		regionInfo := models.NewRegionInfoWithKmsType(
			tea.StringValue(dkmsConfig.RegionId),
			false,
			tea.StringValue(dkmsConfig.Endpoint),
			utils.DkmsType,
		)
		dmc.dKmsConfigsMap[regionInfo] = dkmsConfig
		dmc.regionInfos = append(dmc.regionInfos, regionInfo)
	}
	return nil
}

func (dmc *defaultSecretManagerClient) checkEnvParam(param, paramName string) error {
	if param == "" {
		return errors.New(fmt.Sprintf("env param[%s] is required", paramName))
	}
	return nil
}

func (dmc *defaultSecretManagerClient) retryGetSecretValue(req *kms.GetSecretValueRequest, regionInfo *models.RegionInfo, retryEnd <-chan struct{}) (*kms.GetSecretValueResponse, error) {
	retryTimes := 0
	for {
		select {
		case <-retryEnd:
			return nil, errors.New(fmt.Sprintf("action:retryGetSecretValue, retry end"))
		default:
			waitTimeExponential := dmc.backoffStrategy.GetWaitTimeExponential(retryTimes)
			if waitTimeExponential < 0 {
				return nil, errors.New(fmt.Sprintf("action:retryGetSecretValue, Times limit exceeded"))
			}

			time.Sleep(time.Duration(waitTimeExponential) * time.Millisecond)

			resp, err := dmc.getSecretValue(regionInfo, req)
			if err == nil {
				return resp, nil
			}
			logger.GetCommonLogger(utils.ModeName).Errorf("action:retryGetSecretValue, regionInfo:%+v, %+v", regionInfo, err)
			if !utils.JudgeNeedRecoveryException(err) {
				return nil, err
			}
			retryTimes += 1
		}
	}
}

func (dmc *defaultSecretManagerClient) waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()
	select {
	case <-done:
		return false
	case <-time.After(timeout):
		return true
	}
}
