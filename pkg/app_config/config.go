package app_config

type StudentConfig struct {
	UserToken      string `yaml:"user_token"`
	StudentRepo    string
	StudentRef     string
	ConfigFilename string
}

type S3Config struct {
	Region          string
	EndpointUrl     string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
}

type Config struct {
	GithubRepo      string
	LmsCompanyToken string
	LmsBaseUrl      string
	CallbackTaskId  string
	StudentConfig   StudentConfig
	S3Config        S3Config
}
