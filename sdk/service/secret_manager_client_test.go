package service

import (
	"os"
	"reflect"
	"testing"

	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/models"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/utils"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/kms"
	"github.com/stretchr/testify/assert"
)

var (
	accessKeyId     = os.Getenv("AccessKeyId")
	accessKeySecret = os.Getenv("AccessKeySecret")
)

func TestNewBaseSecretManagerClientBuilder(t *testing.T) {
	base := NewBaseSecretManagerClientBuilder()
	assert.Equal(t, true, *base == baseSecretManagerClientBuilder{})
}

func TestNewDefaultSecretManagerClientBuilder(t *testing.T) {
	builder := NewDefaultSecretManagerClientBuilder()
	assert.Equal(t, true, reflect.DeepEqual(*builder, defaultSecretManagerClientBuilder{}))
}

func TestBaseSecretManagerClientBuilder_Standard(t *testing.T) {
	base := NewBaseSecretManagerClientBuilder()
	builder := base.Standard()
	assert.Equal(t, true, reflect.DeepEqual(*builder, defaultSecretManagerClientBuilder{}))
}

func TestDefaultSecretManagerClientBuilder_Build(t *testing.T) {
	regionId1 := "regionId"
	vpc := true
	vpcEndpoint := "kms-vpc.cn-hangzhou.aliyuncs.com"

	builder := NewDefaultSecretManagerClientBuilder().Standard()
	builder.WithRegion(regionId1)
	builder.AddRegionInfo(&models.RegionInfo{Vpc: vpc, Endpoint: vpcEndpoint})
	builder.WithAccessKey(accessKeyId, accessKeySecret)
	builder.WithBackoffStrategy(&FullJitterBackoffStrategy{3, 2000, 10000})

	client := builder.Build()

	defaultClient, ok := client.(*defaultSecretManagerClient)
	assert.Equal(t, true, ok)
	assert.Equal(t, regionId1, defaultClient.regionInfos[0].RegionId)
	assert.Equal(t, vpc, defaultClient.regionInfos[1].Vpc)
	assert.Equal(t, vpcEndpoint, defaultClient.regionInfos[1].Endpoint)

	credential, ok := defaultClient.credential.(*credentials.AccessKeyCredential)
	assert.Equal(t, true, ok)
	assert.Equal(t, accessKeyId, credential.AccessKeyId)
	assert.Equal(t, accessKeySecret, credential.AccessKeySecret)
}

func TestDefaultSecretManagerClient_GetSecretValue(t *testing.T) {
	regionIds := []string{"cn-hangzhou", "cn-shanghai", "cn-beijing", "cn-xxxx"}
	secretName := "sdk"

	builder := NewDefaultSecretManagerClientBuilder().Standard()
	builder.WithAccessKey(accessKeyId, accessKeySecret)
	builder.WithBackoffStrategy(&FullJitterBackoffStrategy{3, 2000, 10000})
	builder.WithRegion(regionIds...)

	client := builder.Build()

	err := client.Init()
	assert.Nil(t, err)
	defaultClient, ok := client.(*defaultSecretManagerClient)
	assert.Equal(t, true, ok)
	assert.Equal(t, "cn-xxxx", defaultClient.regionInfos[3].RegionId)

	for i := 0; i < 10; i++ {
		request := kms.CreateGetSecretValueRequest()
		request.Scheme = "https"
		request.SecretName = secretName
		request.VersionStage = utils.StageAcsCurrent
		request.FetchExtendedConfig = requests.NewBoolean(true)
		value, err := client.GetSecretValue(request)
		assert.Nil(t, err)
		assert.NotNil(t, value)
	}
}
