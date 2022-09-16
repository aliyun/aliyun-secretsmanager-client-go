package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/models"
	"io/ioutil"
	"strings"
)

func checkConfigParam(param, paramName string) error {
	if param == "" {
		return errors.New(fmt.Sprintf("credentials config missing required parameters[%s]", paramName))
	}
	return nil
}

func LoadCredentialsProperties(fileName string) (*models.CredentialsProperties, error) {
	if fileName == "" {
		fileName = DefaultConfigName
	}
	configMap, err := LoadProperties(fileName)
	if err != nil {
		return nil, err
	}
	if configMap != nil && len(configMap) > 0 {
		credentialsProperties := &models.CredentialsProperties{
			DkmsConfigsMap: make(map[*models.RegionInfo]*models.DkmsConfig),
		}
		err = initDefaultConfig(configMap, credentialsProperties)
		if err != nil {
			return nil, err
		}
		err = initSecretsRegions(configMap, credentialsProperties)
		if err != nil {
			return nil, err
		}
		err = initCredentials(configMap, credentialsProperties)
		if err != nil {
			return nil, err
		}
		initSecretNames(configMap, credentialsProperties)
		return credentialsProperties, nil
	}
	return nil, nil
}

func initDefaultConfig(configMap map[string]string, credentialsProperties *models.CredentialsProperties) error {
	credentialsProperties.PrivateKeyPath = configMap[EnvClientKeyPrivateKeyPathNameKey]
	password, _ := GetPassword(configMap, EnvClientKeyPasswordFromEnvVariableNameKey, PropertiesClientKeyPasswordFromFilePathName)
	credentialsProperties.Password = password
	return nil
}

func initSecretNames(configMap map[string]string, credentialsProperties *models.CredentialsProperties) {
	secretNames := configMap[PropertiesSecretNamesKey]
	if secretNames != "" {
		credentialsProperties.SecretNameSlice = append(credentialsProperties.SecretNameSlice, strings.Split(secretNames, ",")...)
	}
}

func initSecretsRegions(configMap map[string]string, credentialsProperties *models.CredentialsProperties) error {
	var regionInfoSlice []*models.RegionInfo
	err := initDkmsInstances(configMap, &regionInfoSlice, credentialsProperties)
	if err != nil {
		return err
	}
	err = initKmsRegions(configMap, &regionInfoSlice)
	if err != nil {
		return err
	}
	credentialsProperties.RegionInfoSlice = append(credentialsProperties.RegionInfoSlice, regionInfoSlice...)
	return nil
}

func initKmsRegions(configMap map[string]string, regionInfoSlice *[]*models.RegionInfo) error {
	regionInfoJson := configMap[EnvCacheClientRegionIdKey]
	if regionInfoJson == "" {
		return nil
	}
	var regionInfos []map[string]interface{}
	err := json.Unmarshal([]byte(regionInfoJson), &regionInfos)
	if err != nil {
		return errors.New(fmt.Sprintf("credentials config param[%s] is illegal, err:%v", EnvCacheClientRegionIdKey, err))
	}
	for _, regionInfoMap := range regionInfos {
		regionId, err := ParseString(regionInfoMap[EnvRegionRegionIdNameKey])
		if err != nil {
			return err
		}
		endpoint, err := ParseString(regionInfoMap[EnvRegionEndpointNameKey])
		if err != nil {
			return err
		}
		var vpc bool
		if regionInfoMap[EnvRegionVpcNameKey] == "" {
			vpc = false
		} else {
			vpc, err = ParseBool(regionInfoMap[EnvRegionVpcNameKey])
			if err != nil {
				return err
			}
		}
		*regionInfoSlice = append(*regionInfoSlice, models.NewRegionInfoWithKmsType(regionId, vpc, endpoint, KmsType))
	}
	return nil
}

func initDkmsInstances(configMap map[string]string, regionInfoSlice *[]*models.RegionInfo, credentialsProperties *models.CredentialsProperties) error {
	configJson := configMap[CacheClientDkmsConfigInfoKey]
	if configJson == "" {
		return nil
	}
	var dkmsConfigs []*models.DkmsConfig
	err := json.Unmarshal([]byte(configJson), &dkmsConfigs)
	if err != nil {
		return errors.New(fmt.Sprintf("credentials config param[%s] is illegal, err:%v", CacheClientDkmsConfigInfoKey, err))
	}
	for _, dkmsConfig := range dkmsConfigs {
		if tea.StringValue(dkmsConfig.RegionId) == "" || tea.StringValue(dkmsConfig.Endpoint) == "" || tea.StringValue(dkmsConfig.ClientKeyFile) == "" {
			return errors.New("init properties fail, cause of cache_client_dkms_config_info param[regionId or endpoint or clientKeyFile] is empty")
		}
		if !dkmsConfig.IgnoreSslCerts && !strings.Contains(dkmsConfig.CaCert, "-----BEGIN CERTIFICATE-----") {
			caCert, err := ioutil.ReadFile(dkmsConfig.CaCert)
			if err != nil {
				return errors.New(fmt.Sprintf("dkms config CaCert[%s] is illegal, expect certificate pem or correct file path", dkmsConfig.CaCert))
			}
			dkmsConfig.CaCert = string(caCert)
		}
		password, err := GetPassword(configMap, dkmsConfig.PasswordFromEnvVariable, dkmsConfig.PasswordFromFilePathName)
		if err != nil {
			if credentialsProperties.Password == "" {
				return err
			}
			dkmsConfig.Password = tea.String(credentialsProperties.Password)
		} else {
			dkmsConfig.Password = tea.String(password)
		}
		regionInfo := models.NewRegionInfoWithKmsType(
			tea.StringValue(dkmsConfig.RegionId),
			false,
			tea.StringValue(dkmsConfig.Endpoint),
			DkmsType,
		)
		credentialsProperties.DkmsConfigsMap[regionInfo] = dkmsConfig
		*regionInfoSlice = append(*regionInfoSlice, regionInfo)
	}
	return nil
}

func initCredentials(configMap map[string]string, credentialsProperties *models.CredentialsProperties) error {
	credentialsType := configMap[EnvCredentialsTypeKey]
	if credentialsType != "" {
		var credential auth.Credential
		switch credentialsType {
		case "ak":
			accessKeyId := configMap[EnvCredentialsAccessKeyIdKey]
			err := checkConfigParam(accessKeyId, EnvCredentialsAccessKeyIdKey)
			if err != nil {
				return err
			}
			accessSecret := configMap[EnvCredentialsAccessSecretKey]
			err = checkConfigParam(accessSecret, EnvCredentialsAccessKeyIdKey)
			if err != nil {
				return err
			}
			credential = CredentialsWithAccessKey(accessKeyId, accessSecret)
		case "token":
			tokenId := configMap[EnvCredentialsAccessTokenIdKey]
			err := checkConfigParam(tokenId, EnvCredentialsAccessTokenIdKey)
			if err != nil {
				return err
			}
			token := configMap[EnvCredentialsAccessTokenKey]
			err = checkConfigParam(token, EnvCredentialsAccessTokenKey)
			if err != nil {
				return err
			}
			credential = CredentialsWithToken(tokenId, token)
		case "sts", "ram_role":
			accessKeyId := configMap[EnvCredentialsAccessKeyIdKey]
			err := checkConfigParam(accessKeyId, EnvCredentialsAccessKeyIdKey)
			if err != nil {
				return err
			}
			accessSecret := configMap[EnvCredentialsAccessSecretKey]
			err = checkConfigParam(accessSecret, EnvCredentialsAccessKeyIdKey)
			if err != nil {
				return err
			}
			roleSessionName := configMap[EnvCredentialsRoleSessionNameKey]
			err = checkConfigParam(roleSessionName, EnvCredentialsRoleSessionNameKey)
			if err != nil {
				return err
			}
			roleArn := configMap[EnvCredentialsRoleArnKey]
			err = checkConfigParam(roleArn, EnvCredentialsRoleArnKey)
			if err != nil {
				return err
			}
			policy := configMap[EnvCredentialsPolicyKey]
			credential = CredentialsWithRamRoleArnOrSts(accessKeyId, accessSecret, roleSessionName, roleArn, policy)
		case "ecs_ram_role":
			roleName := configMap[EnvCredentialsRoleNameKey]
			err := checkConfigParam(roleName, EnvCredentialsRoleNameKey)
			if err != nil {
				return err
			}
			credential = CredentialsWithEcsRamRole(roleName)
		case "client_key":
			privateKeyPath := configMap[EnvClientKeyPrivateKeyPathNameKey]
			err := checkConfigParam(privateKeyPath, EnvClientKeyPrivateKeyPathNameKey)
			if err != nil {
				return err
			}
			password, err := GetPassword(configMap, EnvClientKeyPasswordFromEnvVariableNameKey, PropertiesClientKeyPasswordFromFilePathName)
			if err != nil {
				return err
			}
			cred, signer, err := LoadRsaKeyPairCredentialAndClientKeySigner(privateKeyPath, password)
			if err != nil {
				return err
			}
			credential = models.NewClientKeyCredential(signer, cred)
		default:
			return errors.New(fmt.Sprintf("config param[%s] is illegal", EnvCredentialsTypeKey))
		}
		credentialsProperties.Credential = credential
	}
	return nil
}
