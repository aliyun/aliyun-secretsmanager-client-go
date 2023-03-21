package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/auth"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/models"
	"golang.org/x/crypto/pkcs12"
	"io/ioutil"
	"os"
	"strings"
)

func LoadRsaKeyPairCredentialAndClientKeySigner(clientKeyPath, password string) (*credentials.RsaKeyPairCredential, *auth.ClientKeySigner, error) {
	clientKeyInfo := &models.ClientKeyInfo{}
	err := parseConfig(clientKeyPath, clientKeyInfo)
	if err != nil {
		return nil, nil, err
	}

	pfx, err := base64.StdEncoding.DecodeString(clientKeyInfo.PrivateKeyData)
	if err != nil {
		return nil, nil, err
	}

	privateKey, _, err := pkcs12.Decode(pfx, password)
	if err != nil {
		return nil, nil, err
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, errors.New("is not rsa private key")
	}

	raw, err := x509.MarshalPKCS8PrivateKey(rsaPrivateKey)
	if err != nil {
		return nil, nil, err
	}
	rawBase64 := base64.StdEncoding.EncodeToString(raw)
	keyPairCredentials := credentials.NewRsaKeyPairCredential(rawBase64, clientKeyInfo.KeyId, 0)
	signer := auth.NewClientKeySigner(clientKeyInfo.KeyId, rawBase64)
	return keyPairCredentials, signer, nil

}
func parseConfig(filepath string, v interface{}) error {
	file, err := os.Open(filepath)
	defer file.Close()
	if err != nil {
		return err
	}
	return json.NewDecoder(file).Decode(v)
}

func GetPassword(configMap map[string]string, envVariableName string, filePathName string) (string, error) {
	var passwordFromEnvVariable, password string
	if configMap != nil {
		passwordFromEnvVariable = configMap[envVariableName]
		if passwordFromEnvVariable != "" {
			password = os.Getenv(passwordFromEnvVariable)
		}
		if password == "" {
			passwordFilePath := configMap[filePathName]
			if passwordFilePath != "" {
				return ReadPasswordFile(passwordFilePath)
			}
		}
	} else {
		passwordFromEnvVariable = os.Getenv(envVariableName)
		if passwordFromEnvVariable != "" {
			password = os.Getenv(passwordFromEnvVariable)
		}
		if password == "" {
			passwordFilePath := os.Getenv(filePathName)
			if passwordFilePath != "" {
				return ReadPasswordFile(passwordFilePath)
			}
		}
	}
	if password == "" {
		password = os.Getenv(EnvClientKeyPasswordNameKey)
	}
	if password == "" {
		return "", errors.New("client key password is not provided")
	}
	return password, nil
}

func ReadPasswordFile(passwordFilePath string) (string, error) {
	content, err := ioutil.ReadFile(passwordFilePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}
