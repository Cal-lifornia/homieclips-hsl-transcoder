package main

import (
	"context"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func setupMinio(ctx context.Context) (*minio.Client, error) {
	minioClient, err := minio.New(os.Getenv("MINIO_URL"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), ""),
		Secure: true,
	})
	if err != nil {
		return nil, err
	}

	exists, errBucketExists := minioClient.BucketExists(ctx, os.Getenv("BUCKET_NAME"))
	if errBucketExists == nil && exists {
		log.Printf("bucket exists and is ours")
		return minioClient, nil
	}

	return nil, errBucketExists
}
