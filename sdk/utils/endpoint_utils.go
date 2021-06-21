package utils

func GetVpcEndpoint(regionId string) string {
	return "kms-vpc." + regionId + ".aliyuncs.com"
}

func GetEndpoint(regionId string) string {
	return "kms." + regionId + ".aliyuncs.com"
}
