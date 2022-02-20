package main

import "os"

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
	minioEndpoint := os.Getenv("S3_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "minio:9000"
	}
	return ApplicationConfig{
		Minio: &MinioConfig{

			Endpoint:  minioEndpoint,
			AccessKey: "root",
			SecretKey: "topsecret",
			Secure:    false,
			Bucket:    "s3-comment",
		},
	}
}
