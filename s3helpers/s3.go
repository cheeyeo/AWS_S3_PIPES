package s3helpers

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func BucketValidator(sess *session.Session, bucket string) (bool, error) {
	svc := s3.New(sess)
	_, err := svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case "NotFound":
				msg := fmt.Sprintf("Bucket %s does not exist", bucket)
				return false, errors.New(msg)
			case "BadRequest":
				msg := fmt.Sprintf("Bucket %s is invalid", bucket)
				return false, errors.New(msg)
			case "Forbidden":
				msg := fmt.Sprintf("Bucket %s is forbidden", bucket)
				return false, errors.New(msg)
			}
		} else {
			return false, err
		}
	}

	return true, nil
}

func GetS3FileSize(sess *session.Session, bucket string, prefix string) (filesize int64, error error) {
	svc := s3.New(sess)
	params := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(prefix),
	}

	resp, err := svc.HeadObject(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case "NotFound":
				msg := fmt.Sprintf("Bucket %s does not exist", bucket)
				return 0, errors.New(msg)
			case "BadRequest":
				msg := fmt.Sprintf("Bucket %s is invalid", bucket)
				return 0, errors.New(msg)
			case "Forbidden":
				msg := fmt.Sprintf("Bucket %s is forbidden", bucket)
				return 0, errors.New(msg)
			default:
				return 0, errors.New(awsErr.Message())
			}
		} else {
			return 0, err
		}
	}

	return *resp.ContentLength, nil
}
