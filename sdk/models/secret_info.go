package models

type SecretInfo struct {
	SecretName            string `json:"secretName"`
	VersionId             string `json:"versionId"`
	SecretValue           string `json:"secretValue"`
	SecretValueByteBuffer []byte `json:"secretValueByteBuffer"`
	SecretDataType        string `json:"secretDataType"`
	CreateTime            string `json:"createTime"`
	SecretType            string `json:"secretType"`
	AutomaticRotation     string `json:"automaticRotation"`
	ExtendedConfig        string `json:"extendedConfig"`
	RotationInterval      string `json:"rotationInterval"`
	NextRotationDate      string `json:"nextRotationDate"`
}

func (si *SecretInfo) Clone() *SecretInfo {
	return &SecretInfo{
		SecretName:            si.SecretName,
		VersionId:             si.VersionId,
		SecretValue:           si.SecretValue,
		SecretValueByteBuffer: append(make([]byte, 0, len(si.SecretValueByteBuffer)), si.SecretValueByteBuffer...),
		SecretDataType:        si.SecretDataType,
		CreateTime:            si.CreateTime,
		SecretType:            si.SecretType,
		AutomaticRotation:     si.AutomaticRotation,
		ExtendedConfig:        si.ExtendedConfig,
		RotationInterval:      si.RotationInterval,
		NextRotationDate:      si.NextRotationDate,
	}
}
