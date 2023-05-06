package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func uploadPart(svc *s3.S3, resp *s3.CreateMultipartUploadOutput, fileBytes []byte, partNumber int, maxRetries int) (*s3.CompletedPart, error) {
	tryNum := 1

	partInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(fileBytes),
		Bucket:        resp.Bucket,
		Key:           resp.Key,
		PartNumber:    aws.Int64(int64(partNumber)),
		UploadId:      resp.UploadId,
		ContentLength: aws.Int64(int64(len(fileBytes))),
	}

	for tryNum <= maxRetries {
		uploadResult, err := svc.UploadPart(partInput)
		if err != nil {
			if tryNum == maxRetries {
				if aerr, ok := err.(awserr.Error); ok {
					return nil, aerr
				}
				return nil, err
			}
			fmt.Printf("Retrying to upload part #%v\n", partNumber)
			tryNum++
		} else {
			fmt.Printf("Uploaded part #%v\n", partNumber)
			return &s3.CompletedPart{
				ETag:       uploadResult.ETag,
				PartNumber: aws.Int64(int64(partNumber)),
			}, nil
		}
	}

	return nil, nil
}

func abortMultipartUpload(svc *s3.S3, resp *s3.CreateMultipartUploadOutput) error {
	fmt.Println("Aborting multipart upload for UploadId#" + *resp.UploadId)
	abortInput := &s3.AbortMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
	}
	_, err := svc.AbortMultipartUpload(abortInput)
	return err
}

func completeMultipartUpload(svc *s3.S3, resp *s3.CreateMultipartUploadOutput, completedParts []*s3.CompletedPart) (*s3.CompleteMultipartUploadOutput, error) {
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}
	return svc.CompleteMultipartUpload(completeInput)
}

func main() {
	sess := session.Must(session.NewSession())
	svc := s3.New(sess)

	file, err := os.Open("data/2GB.txt")
	if err != nil {
		fmt.Printf("Err opening file: %v\n", err)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	buffer := make([]byte, fileSize)
	fileType := http.DetectContentType(buffer)
	file.Read(buffer)

	input := &s3.CreateMultipartUploadInput{
		Bucket:      aws.String("testpipeuploads"),
		Key:         aws.String("multiparttest"),
		ContentType: aws.String(fileType),
	}

	resp, err := svc.CreateMultipartUpload(input)
	if err != nil {
		fmt.Printf("ERR: %v\n", err.Error())
	}
	fmt.Println("Created multipart upload request")

	maxPartSize := int64(5 * 1024 * 1024)
	maxRetries := 3

	var curr, partLength int64
	var remaining = fileSize
	var completedParts []*s3.CompletedPart
	partNumber := 1
	for curr = 0; remaining != 0; curr += partLength {
		if remaining < maxPartSize {
			partLength = remaining
		} else {
			partLength = int64(maxPartSize)
		}

		completedPart, err := uploadPart(svc, resp, buffer[curr:curr+partLength], partNumber, maxRetries)
		if err != nil {
			fmt.Println(err.Error())
			err := abortMultipartUpload(svc, resp)
			if err != nil {
				fmt.Println(err.Error())
			}
			return
		}
		remaining -= partLength
		partNumber++
		completedParts = append(completedParts, completedPart)
	}

	completeResponse, err := completeMultipartUpload(svc, resp, completedParts)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("Success: %s", completeResponse.String())
}
