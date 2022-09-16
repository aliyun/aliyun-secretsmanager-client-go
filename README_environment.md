# System Environment Variables Setting For Aliyun Secrets Manager Client 

Use Aliyun Secrets Manager client by system environment variables with the below ways:

* Use access key to access aliyun kms, you must set the following system environment variables (for linux):

	- export credentials\_type=ak
	- export credentials\_access\_key\_id=\<your access key id>
	- export credentials\_access\_secret=\<your access key secret>
	- export cache\_client\_region\_id=[{"regionId":"\<your region id>"}]

* Use STS to access aliyun kms, you must set the following system environment variables (for linux):

	- export credentials\_type=sts
	- export credentials\_role\_session_name=\<your role name>
	- export credentials\_role\_arn=\<your role arn>
	- export credentials\_access\_key\_id=\<your access key id>
	- export credentials\_access\_secret=\<your access key secret>
	- export cache\_client\_region\_id=[{"regionId":"\<your region id>"}]

* Use RAM role to access aliyun kms, you must set the following system environment variables (for linux):

	- export credentials_type=ram\_role
	- export credentials\_role\_session\_name=\<your role name>
	- export credentials\_role\_arn=\<your role arn>
	- export credentials\_access\_key\_id=\<your access key id>
	- export credentials\_access\_secret=\<your access key secret>
	- export cache\_client\_region\_id=[{"regionId":"\<your region id>"}]

* Use ECS RAM role to access aliyun kms, you must set the following system environment variables (for linux):

	- export credentials\_type=ecs\_ram\_role
	- export credentials\_role\_session\_name=\<your role name>
	- export cache\_client\_region\_id=[{"regionId":"\<your region id>"}]

* Use client key to access aliyun kms, you must set the following system environment variables (for linux):

	- export credentials\_type=client\_key
	- export client\_key\_password\_from\_env\_variable=\<your client key private key password from environment variable>
	- export client\_key\_password\_from\_file\_path=\<your client key private key password from file>
	- export client\_key\_private\_key\_path=\<your client key private key file path>
	- export cache\_client\_region\_id=[{"regionId":"\<your region id>"}]

* Access aliyun dedicated kms, you must set the following system environment variables (for linux):

	- export cache_client_dkms_config_info=[{"ignoreSslCerts":false,"caCert":"path/to/caCert","passwordFromFilePathName":"client_key_password_from_file_path","clientKeyFile":"\<your client key file absolute path>","regionId":"\<your dkms region>","endpoint":"\<your dkms endpoint>"}]
    ```
        The details of the configuration item named cache_client_dkms_config_info:
        1. The configuration item named cache_client_dkms_config_info must be configured as a json array, you can configure multiple region instances
        2. ignoreSslCerts:If ignore ssl certs (true: Ignores the ssl certificate, false: Validates the ssl certificate)
  		3. caCert:CA certificate file path, or certificate pem content. If ignoreSslCerts is false, caCert is required. if ignoreSslCerts is true, ignore caCert.
        4. passwordFromFilePathName and passwordFromEnvVariable
          passwordFromFilePathName:The client key password configuration is obtained from the file,choose one of the two with passwordFromEnvVariable.
          e.g. while configuring passwordFromFilePathName: "client_key_password_from_file_path",
                       You need to add client_key_password_from_file_path=< your password file absolute path > in env.
                       and correspond to a file with a password written on it.
          passwordFromEnvVariable:The client key password configuration is obtained from the environment variable,choose one of the two with passwordFromFilePathName.
          e.g. while configuring passwordFromEnvVariable: "client_key_password_from_env_variable",
                       You need to add client_key_password_from_env_variable=< your client key private key password from environment variable > in env
                       and the corresponding env variable (xxx_env_variable=<your password>).
        5. clientKeyFile:The absolute path to the client key json file
        6. regionId:Region id
        7. endpoint:Domain address of dkms
    ```