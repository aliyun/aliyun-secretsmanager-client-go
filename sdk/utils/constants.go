package utils

const (
	// 凭据当前stage
	StageAcsCurrent = "ACSCurrent"

	// 随机IV字节长度
	IvLength = 16

	// 随机密钥字节长度
	RandomKeyLength = 32

	// 默认最大重试次数
	DefaultRetryMaxAttempts = 5

	// 默认重试间隔时间
	DefaultRetryInitialIntervalMills = 2000

	// 默认最大等待时间
	DefaultCapacity = 10000

	// 环境变量cache_client_region_id key
	EnvCacheClientRegionIdKey = "cache_client_region_id"

	// 环境变量credentials_type key
	EnvCredentialsTypeKey = "credentials_type"

	// 环境变量credentials_access_key_id key
	EnvCredentialsAccessKeyIdKey = "credentials_access_key_id"

	// 环境变量credentials_access_secret key
	EnvCredentialsAccessSecretKey = "credentials_access_secret"

	// 环境变量credentials_access_token_id key
	EnvCredentialsAccessTokenIdKey = "credentials_access_token_id"

	// 环境变量credentials_access_token key
	EnvCredentialsAccessTokenKey = "credentials_access_token"

	// 环境变量credentials_role_session_name key
	EnvCredentialsRoleSessionNameKey = "credentials_role_session_name"

	// 环境变量credentials_role_arn key
	EnvCredentialsRoleArnKey = "credentials_role_arn"

	// 环境变量credentials_policy key
	EnvCredentialsPolicyKey = "credentials_policy"

	// 环境变量credentials_role_name key
	EnvCredentialsRoleNameKey = "credentials_role_name"

	// 欠费errorCode
	ErrorCodeForbiddenInDebtOverDue = "Forbidden.InDebtOverdue"

	// 欠费errorCode
	ErrorCodeForbiddenInDebt = "Forbidden.InDebt"

	// 模块名称
	ModeName = "CacheClient"

	// 凭据文本数据类型
	TextDataType = "text"

	// 凭据二进制数据类型
	BinaryDataType = "binary"

	// 项目版本
	ProjectVersion = "1.1.1"

	// the user agent of secrets manager golang
	UserAgentOfSecretsManagerGolang = "aliyun-secretsmanager-client-go"

	// 环境变量region中endPoint key
	EnvRegionEndpointNameKey = "endpoint"

	// 环境变量region中regionId key
	EnvRegionRegionIdNameKey = "regionId"

	// 环境变量region中regionId key
	EnvRegionVpcNameKey = "vpc"

	// KMS服务Socket连接超时错误码
	SdkReadTimeout  = "SDK.ReadTimeout"
	SdkTimeoutError = "SDK.TimeoutError"

	// KMS服务无法连接错误码
	SdkServerUnreachable = "SDK.ServerUnreachable"

	//环境变量client_key_password key
	EnvClientKeyPasswordNameKey = "client_key_password"

	// 环境变量credentials_client_key_private_key_path key
	EnvClientKeyPrivateKeyPathNameKey = "client_key_private_key_path"

	// 环境变量 client_key_password_from_env_variable key
	EnvClientKeyPasswordFromEnvVariableNameKey = "client_key_password_from_env_variable"

	// 配置文件 client_key_password_from_file_path key
	PropertiesClientKeyPasswordFromFilePathName = "client_key_password_from_file_path"
	// 默认的 credentials配置文件
	DefaultConfigName = "secretsmanager.properties"

	// 配置文件 secret_names key
	PropertiesSecretNamesKey = "secret_names"

	// 环境变量cache_client_dkms_config_info key
	CacheClientDkmsConfigInfoKey = "cache_client_dkms_config_info"

	// 环境变量 ignoreSSLCerts key
	ENV_IGNORE_SSL_CERTS_KEY = "ignoreSSLCerts"

	// 虚假的AKSK
	PretendAk = "PRETEND_AK"
	PretendSk = "PRETEND_SK"

	// KMS类型
	DkmsType = 1
	KmsType  = 0
)
