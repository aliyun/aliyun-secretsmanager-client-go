package utils

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
)

func CredentialsWithAccessKey(accessKeyId, accessKeySecret string) auth.Credential {
	return &credentials.AccessKeyCredential{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
	}
}

func CredentialsWithToken(tokenId, token string) auth.Credential {
	return &credentials.AccessKeyCredential{
		AccessKeyId:     tokenId,
		AccessKeySecret: token,
	}
}

func CredentialsWithRamRoleArnOrSts(accessKeyId, accessKeySecret, roleSessionName, roleArn, policy string) auth.Credential {
	return &credentials.RamRoleArnCredential{
		AccessKeyId:           accessKeyId,
		AccessKeySecret:       accessKeySecret,
		RoleSessionName:       roleSessionName,
		RoleArn:               roleArn,
		Policy:                policy,
		RoleSessionExpiration: 3600,
	}
}

func CredentialsWithEcsRamRole(roleName string) auth.Credential {
	return &credentials.EcsRamRoleCredential{
		RoleName: roleName,
	}
}
