# 阿里云凭据管家Go客户端

阿里云凭据管家Go客户端可以使Go开发者快速使用阿里云凭据。

*其他语言版本: [English](README.md), [简体中文](README.zh-cn.md)*

## 许可证

[Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0.html)


## 优势
* 支持用户快速集成获取凭据信息
* 支持阿里云凭据管家内存和文件两种缓存凭据机制
* 支持凭据名称相同场景下的跨地域容灾
* 支持默认规避策略和用户自定义规避策略

## 软件要求

- 您的系统需要达到 环境要求, 例如，安装了不低于 1.10.x 版本的 Go 环境。

## 安装

使用 go get 下载安装 SDK

```sh
$ go get -u github.com/aliyun/aliyun-secretsmanager-client-go
```


## 示例代码
### 一般用户代码
* 通过系统环境变量或配置文件(secretsmanager.properties)构建客户端([系统环境变量设置详情](README_environment.zh-cn.md)、[配置文件设置详情](README_config.zh-cn.md))
```go
package main

import "github.com/aliyun/aliyun-secretsmanager-client-go/sdk"

func main() {
	client, err := sdk.NewClient()
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
* 通过指定参数(accessKey、accessSecret、regionId等)构建客户端
```go
package main

import (
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/service"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk"
	"os"
)

func main() {
	client, err := sdk.NewSecretCacheClientBuilder(service.NewDefaultSecretManagerClientBuilder().Standard().WithAccessKey(os.Getenv("#accessKeyId#"), os.Getenv("#accessKeySecret#")).WithRegion("#regionId#").Build()).Build()
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

### 定制化用户代码
* 使用自定义参数或用户自己实现
```go
package main

import (
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/service"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk"
    "github.com/aliyun/aliyun-secretsmanager-client-go/sdk/cache"
	"os"
)

func main() {
	client, err := sdk.NewSecretCacheClientBuilder(
    		service.NewDefaultSecretManagerClientBuilder().Standard().WithAccessKey(os.Getenv("#accessKeyId#"), os.Getenv("#accessKeySecret#")).WithRegion("#regionId#").WithBackoffStrategy(&service.FullJitterBackoffStrategy{RetryMaxAttempts: 3, RetryInitialIntervalMills: 2000, Capacity: 10000}).Build()).WithCacheSecretStrategy(cache.NewFileCacheSecretStoreStrategy("#cacheSecretPath#", true, "#salt#")).WithRefreshSecretStrategy(service.NewDefaultRefreshSecretStrategy("#jsonTTLPropertyName#")).WithCacheStage("ACSCurrent").WithSecretTTL("#secretName#", 1*60*1000).Build()
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
