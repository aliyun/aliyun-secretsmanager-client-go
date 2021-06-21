# Aliyun Secrets Manager Client for Go

The Aliyun Secrets Manager Client for Go enables Go developers to easily work with Aliyun KMS Secrets. 

*Read this in other languages: [English](README.md), [简体中文](README.zh-cn.md)*

- [Aliyun Secrets Manager Client Homepage](https://help.aliyun.com/document_detail/190269.html?spm=a2c4g.11186623.6.621.201623668WpoMj)
- [Issues](https://github.com/aliyun/alibabacloud-secretsmanager-client-go/issues)
- [Release](https://github.com/aliyun/alibabacloud-secretsmanager-client-go/releases)

## License

[Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0.html)

## Features
* Provide quick integration capability to gain secret information
* Provide Alibaba secrets cache ( memory cache or encryption file cache )
* Provide tolerated disaster by the secrets with the same secret name and secret data in different regions
* Provide default backoff strategy and user-defined backoff strategy

## Requirements

- You must use Go 1.10.x or later.

## Installation
Use `go get` to install SDK：

```sh
$ go get -u github.com/aliyun/aliyun-secretsmanager-client-go
```


## Sample Code
### Ordinary User Sample Code
* Build Secrets Manager Client by system environment variables ([system environment variables setting for details](README_environment.md))

```go
package main

import "github.com/aliyun/aliyun-secretsmanager-client-go/sdk/service"

func main() {
	client, err := service.NewClient()
	if err != nil {
		// Handle exceptions
		panic(err)
	}
	secretInfo, err := client.GetSecretInfo("#secretName#")
	if err != nil {
		// Handle exceptions
		panic(err)
	}
}
    
```


*  Build Secrets Manager Client by the given parameters(accessKey, accessSecret, regionId, etc)

```go
package main

import (
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/service"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk"
)

func main() {
	client, err := sdk.NewSecretCacheClientBuilder(service.NewDefaultSecretManagerClientBuilder().Standard().WithAccessKey("#accessKeyId#", "#accessKeySecret#").WithRegion("#regionId#").Build()).Build()
	if err != nil {
		// Handle exceptions
		panic(err)
	}
	secretInfo, err := client.GetSecretInfo("#secretName#")
	if err != nil {
		// Handle exceptions
		panic(err)
	}
}
```
### Particular User Sample Code
* Use custom parameters or customized implementation

```go
package main

import (
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/service"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk"
)

func main() {
	client, err := sdk.NewSecretCacheClientBuilder(
		service.NewDefaultSecretManagerClientBuilder().Standard()
	.WithAccessKey(accessKeyId, accessKeySecret).WithRegion("#regionId#")
	.WithBackoffStrategy(&service.FullJitterBackoffStrategy{RetryMaxAttempts: 3, RetryInitialIntervalMills: 2000, Capacity: 10000}).Build())
	.WithCacheSecretStrategy(cache.NewFileCacheSecretStoreStrategy("#cacheSecretPath#", true, "#salt#"))
	.WithRefreshSecretStrategy(service.NewDefaultRefreshSecretStrategy("#jsonTTLPropertyName#"))
	.WithCacheStage("ACSCurrent")
	.WithSecretTTL(secretName, 1*60*1000).Build()
	if err != nil {
		// Handle exceptions
		panic(err)
	}
	secretInfo, err := client.GetSecretInfo("#secretName#")
	if err != nil {
		// Handle exceptions
		panic(err)
	}
}
```

 
