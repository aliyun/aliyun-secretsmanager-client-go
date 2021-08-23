package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/models"
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
		fileName = DEFAULT_CONFIG_NAME
	}
	configMap, err := LoadProperties(fileName)
	if err != nil {
		return nil, err
	}
	var regionInfoSlice []*models.RegionInfo
	var secretNameSlice []string
	if configMap != nil && len(configMap) > 0 {
		credentialsType := configMap[EnvCredentialsTypeKey]
		accessKeyId := configMap[EnvCredentialsAccessKeyIdKey]
		accessSecret := configMap[EnvCredentialsAccessSecretKey]
		regionInfoJson := configMap[EnvCacheClientRegionIdKey]
		err = checkConfigParam(credentialsType, EnvCredentialsTypeKey)
		if err != nil {
			return nil, err
		}
		err = checkConfigParam(regionInfoJson, EnvCacheClientRegionIdKey)
		if err != nil {
			return nil, err
		}
		var regionInfoList []map[string]interface{}
		err := json.Unmarshal([]byte(regionInfoJson), &regionInfoList)
		if err != nil {
			return nil, err
		}
		for _, regionInfoMap := range regionInfoList {
			regionId, err := ParseString(regionInfoMap[EnvRegionRegionIdNameKey])
			if err != nil {
				return nil, err
			}
			endpoint, err := ParseString(regionInfoMap[EnvRegionEndpointNameKey])
			if err != nil {
				return nil, err
			}
			var vpc bool
			if regionInfoMap[EnvRegionVpcNameKey] == "" {
				vpc = false
			} else {
				vpc, err = ParseBool(regionInfoMap[EnvRegionVpcNameKey])
				if err != nil {
					return nil, err
				}
			}
			regionInfo := models.NewRegionInfoWithVpcEndpoint(regionId, vpc, endpoint)
			regionInfoSlice = append(regionInfoSlice, regionInfo)
		}
		var credential auth.Credential
		switch credentialsType {
		case "ak":
			credential = CredentialsWithAccessKey(accessKeyId, accessSecret)
			break
		case "token":
			tokenId := configMap[EnvCredentialsAccessTokenIdKey]
			err = checkConfigParam(tokenId, EnvCredentialsAccessTokenIdKey)
			if err != nil {
				return nil, err
			}
			token := configMap[EnvCredentialsAccessTokenKey]
			err = checkConfigParam(token, EnvCredentialsAccessTokenKey)
			if err != nil {
				return nil, err
			}
			credential = CredentialsWithToken(tokenId, token)
			break
		case "sts":
		case "ram_role":
			roleSessionName := configMap[EnvCredentialsRoleSessionNameKey]
			err = checkConfigParam(roleSessionName, EnvCredentialsRoleSessionNameKey)
			if err != nil {
				return nil, err
			}
			roleArn := configMap[EnvCredentialsRoleArnKey]
			err = checkConfigParam(roleArn, EnvCredentialsRoleArnKey)
			if err != nil {
				return nil, err
			}
			policy := configMap[EnvCredentialsPolicyKey]
			credential = CredentialsWithRamRoleArnOrSts(accessKeyId, accessSecret, roleSessionName, roleArn, policy)
			break
		case "ecs_ram_role":
			roleName := configMap[EnvCredentialsRoleNameKey]
			err = checkConfigParam(roleName, EnvCredentialsRoleNameKey)
			if err != nil {
				return nil, err
			}
			credential = CredentialsWithEcsRamRole(roleName)
			break
		case "client_key":
			privateKeyPath := configMap[EnvClientKeyPrivateKeyPathNameKey]
			err = checkConfigParam(privateKeyPath, EnvClientKeyPrivateKeyPathNameKey)
			if err != nil {
				return nil, err
			}
			password, err := GetPassword(configMap)
			if err != nil {
				return nil, err
			}
			cred, signer, err := LoadRsaKeyPairCredentialAndClientKeySigner(privateKeyPath, password)
			if err != nil {
				return nil, err
			}
			credential = models.NewClientKeyCredential(signer, cred)
			break
		default:
			return nil, errors.New(fmt.Sprintf("config param[%s] is illegal", EnvCredentialsTypeKey))
		}
		secretNames := configMap[PROPERTIES_SECRET_NAMES_KEY]
		if secretNames != "" {
			secretNameSlice = append(secretNameSlice, strings.Split(secretNames, ",")...)
		}
		credentialsProperties := models.NewCredentialsProperties(credential, secretNameSlice, regionInfoSlice, configMap)
		return credentialsProperties, nil
	}
	return nil, nil
}
