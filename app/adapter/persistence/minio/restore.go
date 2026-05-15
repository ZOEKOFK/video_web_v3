package minio

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/minio/minio-go/v7"
)

var BucketName = "uploads"

var videoContentTypes = map[string]string{
	"mp4":  "video/mp4",
	"avi":  "video/x-msvideo",
	"mov":  "video/quicktime",
	"wmv":  "video/x-ms-wmv",
	"flv":  "video/x-flv",
	"mkv":  "video/x-matroska",
	"webm": "video/webm",
}

func getVideoContentType(ext string) string {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	if ct, ok := videoContentTypes[ext]; ok {
		return ct
	}
	return "video/octet-stream"
}

func Save(objectName string, data []byte) error {
	contentType := "application/octet-stream"
	_, err := MinioClient.PutObject(context.Background(), BucketName, objectName, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Printf("[MinIO] 上传失败: %v", err)
		return err
	}
	log.Printf("[MinIO] 上传成功: %s (size=%d)", objectName, len(data))
	return nil
}

func SaveVideo(objectName string, data []byte, fileExtension string) (string, error) {
	contentType := getVideoContentType(fileExtension)
	info, err := MinioClient.PutObject(context.Background(), BucketName, objectName, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Printf("视频上传失败: %v", err)
		return "", err
	}
	log.Println(info)
	url := fmt.Sprintf("%s/%s", BucketName, objectName)
	return url, nil
}

func Delete(objectName string) error {
	err := MinioClient.RemoveObject(context.Background(), BucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("[MinIO] 删除失败: %v", err)
		return err
	}
	log.Printf("[MinIO] 删除成功: %s", objectName)
	return nil
}

func EnsureBucketExists() error {
	exists, err := MinioClient.BucketExists(context.Background(), BucketName)
	if err != nil {
		return fmt.Errorf("检查 bucket 失败: %w", err)
	}
	if !exists {
		err = MinioClient.MakeBucket(context.Background(), BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("创建 bucket 失败: %w", err)
		}
		log.Printf("[MinIO] 创建 bucket 成功: %s", BucketName)
	} else {
		log.Printf("[MinIO] bucket 已存在: %s", BucketName)
	}
	return nil
}
