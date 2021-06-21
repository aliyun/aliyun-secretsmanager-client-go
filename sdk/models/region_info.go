package models

type RegionInfo struct {
	// region id
	RegionId string
	// 表示程序运行的网络是否为VPC网络
	Vpc bool
	// 终端地址信息
	Endpoint string
}

type RegionInfoExtend struct {
	*RegionInfo
	Escaped   float64
	Reachable bool
}

func NewRegionInfoWithRegionId(regionId string) *RegionInfo {
	return &RegionInfo{
		RegionId: regionId,
	}
}

func NewRegionInfoWithEndpoint(regionId string, endpoint string) *RegionInfo {
	return &RegionInfo{
		RegionId: regionId,
		Endpoint: endpoint,
	}
}

func NewRegionInfoWithVpcEndpoint(regionId string, vpc bool, endpoint string) *RegionInfo {
	return &RegionInfo{
		RegionId: regionId,
		Vpc:      vpc,
		Endpoint: endpoint,
	}
}
