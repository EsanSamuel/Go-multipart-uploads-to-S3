package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

func S3Config() *s3.S3 {
	err := godotenv.Load(".env")

	if err != nil {
		log.Println("Warning: unable to find .env file")
	}
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	//bucketName := os.Getenv("AWS_BUCKET_NAME")

	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String("us-west-1"),
		Endpoint: aws.String("https://t3.storage.dev"),
		Credentials: credentials.NewStaticCredentials(
			accessKey, secretKey, "",
		),
		S3ForcePathStyle: aws.Bool(true),
	})

	if err != nil {
		fmt.Println(err.Error())
	}

	S3Client := s3.New(sess)

	return S3Client
}

var s3client = S3Config()

func GetUploadId(filename string) *string {
	bucketName := os.Getenv("AWS_BUCKET_NAME")

	resp, err := s3client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: &bucketName,
		Key:    aws.String(filename),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	return resp.UploadId
}

func UploadPartToS3(filename string, uploadId string, pathNumber int, data []byte) (*s3.UploadPartOutput, error) {
	bucketName := os.Getenv("AWS_BUCKET_NAME")
	input, err := s3client.UploadPart(&s3.UploadPartInput{
		Body:       bytes.NewReader(data),
		Bucket:     aws.String(bucketName),
		Key:        aws.String(filename),
		PartNumber: aws.Int64(int64(pathNumber)),
		UploadId:   aws.String(uploadId),
	})

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return input, nil
}

func UploadFinish(filename string, parts []*s3.CompletedPart, uploadId string) (string, error) {
	bucketName := os.Getenv("AWS_BUCKET_NAME")
	sort.Slice(parts, func(i, j int) bool {
		return *parts[i].PartNumber < *parts[j].PartNumber
	})
	_, err := s3client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucketName),
		Key:      aws.String(filename),
		UploadId: aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: parts,
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	url := "https://file-uploads.t3.storage.dev/" + filename

	return url, nil
}
