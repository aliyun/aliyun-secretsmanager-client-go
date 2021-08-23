package models

type ClientKeyInfo struct {
	KeyId          string
	PrivateKeyData string
}

func NewClientKeyInfo(keyId string, privateKeyData string) *ClientKeyInfo {
	return &ClientKeyInfo{KeyId: keyId, PrivateKeyData: privateKeyData}
}
