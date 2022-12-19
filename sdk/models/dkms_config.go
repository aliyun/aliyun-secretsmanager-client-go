package models

import dedicatedkmsopenapi "github.com/aliyun/alibabacloud-dkms-gcs-go-sdk/openapi"

type DkmsConfig struct {
	*dedicatedkmsopenapi.Config
	IgnoreSslCerts           bool
	PasswordFromEnvVariable  string
	PasswordFromFilePathName string
	CaCert                   string
	CaFilePath               string
	PasswordFromFilePath     string
}
