# 阿里云凭据管家客户端系统环境变量设置 

通过以下系统环境变量设置方式使用阿里云凭据管家客户端：

* 通过使用AK访问KMS, 你必须要设置如下系统环境变量 (linux):

	- export credentials\_type=ak
	- export credentials\_access\_key\_id=\<your access key id>
	- export credentials\_access\_secret=\<your access key secret>
	- export cache\_client\_region\_id=[{"regionId":"\<your region id>"}]

* 通过使用STS访问KMS, 你必须要设置如下系统环境变量 (linux):

	- export credentials\_type=sts
	- export credentials\_role\_session_name=\<your role name>
	- export credentials\_role\_arn=\<your role arn>
	- export credentials\_access\_key\_id=\<your access key id>
	- export credentials\_access\_secret=\<your access key secret>
	- export cache\_client\_region\_id=[{"regionId":"\<your region id>"}]
	
* 通过使用RAM Role访问KMS, 你必须要设置如下系统环境变量 (linux):

	- export credentials_type=ram\_role
	- export credentials\_role\_session\_name=\<your role name>
	- export credentials\_role\_arn=\<your role arn>
	- export credentials\_access\_key\_id=\<your access key id>
	- export credentials\_access\_secret=\<your access key secret>
	- export cache\_client\_region\_id=[{"regionId":"\<your region id>"}]

* 通过使用ECS RAM Role访问KMS, 你必须要设置如下系统环境变量 (linux):

	- export credentials\_type=ecs\_ram\_role
	- export credentials\_role\_session\_name=\<your role name>
	- export cache\_client\_region\_id=[{"regionId":"\<your region id>"}]

* 通过使用Client Key访问KMS, 你必须要设置如下系统环境变量 (linux):

    - export credentials\_type=client\_key
    - export client\_key\_password\_from\_env\_variable=\<your client key private key password from environment variable>
    - export client\_key\_password\_from\_file\_path=\<your client key private key password from file>
    - export client\_key\_private\_key\_path=\<your client key private key file path>
    - export cache\_client\_region\_id=[{"regionId":"\<your region id>"}]

* 访问DKMS，你必须要设置如下系统环境变量 (linux):

	- export cache_client_dkms_config_info=[{"regionId":"\<your dkms region>","endpoint":"\<your dkms endpoint>","passwordFromEnvVariable":"your_password_env_variable","clientKeyFile":"\<your client key file path>","ignoreSslCerts":false,"caFilePath":"\<your CA certificate file path>"}]
    ```
        cache_client_dkms_config_info配置项说明:
        1. cache_client_dkms_config_info配置项为json数组，支持配置多个region实例
        2. regionId:地域Id
        3. endpoint:专属kms的域名地址
        4. passwordFromFilePath和passwordFromEnvVariable
           passwordFromFilePath:client key密码配置从文件中获取，与passwordFromEnvVariable二选一
           例:当配置passwordFromFilePath:<你的client key密码文件所在的路径>,需在配置的路径下配置写有password的文件
           passwordFromEnvVariable:client key密码配置从环境变量中获取，与passwordFromFilePath二选一
           例:当配置"passwordFromEnvVariable":"your_password_env_variable"时，
             需在环境变量中添加your_password_env_variable=<你的client key对应的密码>
        5. clientKeyFile:client key json文件的路径
        6. ignoreSslCerts:是否忽略ssl证书 (true:忽略ssl证书,false:验证ssl证书)
        7. caFilePath:专属kms的CA证书路径
    ```
 