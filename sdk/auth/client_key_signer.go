package auth

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/signers"
)

type ClientKeySigner struct {
	PublicKeyId string
	PrivateKey  string
}

func NewClientKeySigner(publicKeyId, privateKey string) *ClientKeySigner {
	return &ClientKeySigner{
		PublicKeyId: publicKeyId,
		PrivateKey:  privateKey,
	}
}

func (signer *ClientKeySigner) GetName() string {
	return "SHA256withRSA"
}
func (signer *ClientKeySigner) GetType() string {
	return "PRIVATEKEY"
}

func (signer *ClientKeySigner) GetVersion() string {
	return "1.0"
}

func (signer *ClientKeySigner) GetAccessKeyId() (string, error) {
	return signer.PublicKeyId, nil
}

func (signer *ClientKeySigner) GetExtraParam() map[string]string {
	return make(map[string]string)
}

func (signer *ClientKeySigner) Sign(stringToSign, secretSuffix string) string {
	return signers.Sha256WithRsa(stringToSign, signer.PrivateKey)
}
