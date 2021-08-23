package models

import "github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth"

type ClientKeyCredential struct {
	Signer     auth.Signer
	Credential auth.Credential
}

func NewClientKeyCredential(signer auth.Signer, credential auth.Credential) *ClientKeyCredential {
	return &ClientKeyCredential{Signer: signer, Credential: credential}
}
