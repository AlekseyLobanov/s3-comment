package main

type ApplicationConfig struct {
	Minio *MinioConfig
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    bool
	Bucket    string
}

func ReadConfigFromEnvs() ApplicationConfig {
	return ApplicationConfig{
		Minio: &MinioConfig{
			Endpoint:  "minio:9000",
			AccessKey: "root",
			SecretKey: "topsecret",
			Secure:    false,
			Bucket:    "s3-comment",
		},
	}
}
