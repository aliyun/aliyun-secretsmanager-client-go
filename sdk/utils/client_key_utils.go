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

func GetPassword(configMap map[string]string) (string, error) {
	passwordFromEnvVariable := os.Getenv(EnvClientKeyPasswordFromEnvVariableNameKey)
	password := ""
	if passwordFromEnvVariable != "" {
		password = os.Getenv(passwordFromEnvVariable)
	}
	if password == "" {
		if configMap != nil {
			passwordFilePath := configMap[PropertiesClientKeyPasswordFromFilePathName]
			if passwordFilePath != "" {
				file, err := os.Open(passwordFilePath)
				defer file.Close()
				if err != nil {
					return "", err
				}
				fd, err := ioutil.ReadAll(file)
				if err != nil {
					return "", err
				}
				password = string(fd)
			}
			if password == "" {
				password = configMap[EnvClientKeyPasswordNameKey]
			}
		}
	}
	if password == "" {
		password = os.Getenv(EnvClientKeyPasswordNameKey)
	}
	return password, nil
}
