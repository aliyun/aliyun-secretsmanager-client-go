package models

import "github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth"

type CredentialsProperties struct {
	Credential       auth.Credential
	SecretNameSlice  []string
	RegionInfoSlice  []*RegionInfo
	SourceProperties map[string]string
}

func NewCredentialsProperties(credential auth.Credential, secretNameSlice []string, regionInfoSlice []*RegionInfo, sourceProperties map[string]string) *CredentialsProperties {
	return &CredentialsProperties{Credential: credential, SecretNameSlice: secretNameSlice, RegionInfoSlice: regionInfoSlice, SourceProperties: sourceProperties}
}
