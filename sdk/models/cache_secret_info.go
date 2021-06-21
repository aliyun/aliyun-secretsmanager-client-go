package models

type CacheSecretInfo struct {
	SecretInfo       *SecretInfo `json:"secretInfo"`
	Stage            string      `json:"stage"`
	RefreshTimestamp int64       `json:"refreshTimestamp"`
}

func (csi *CacheSecretInfo) Clone() *CacheSecretInfo {
	return &CacheSecretInfo{
		SecretInfo:       csi.SecretInfo.Clone(),
		Stage:            csi.Stage,
		RefreshTimestamp: csi.RefreshTimestamp,
	}
}
