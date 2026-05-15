package minio

import (
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var MinioClient *minio.Client

func init() {
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	endPoint := "127.0.0.1:9000"
	useSSL := false // 用不用http通讯
	minioClient, err := minio.New(endPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}
	MinioClient = minioClient
}
