package s3helpers

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func BucketValidator(bucket string) (bool, error) {
	sess := session.Must(session.NewSession())
	svc := s3.New(sess)
	_, err := svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case "NotFound":
				fmt.Println(awsErr.Error())
				msg := fmt.Sprintf("Bucket %s does not exist", bucket)
				return false, errors.New(msg)
			case "BadRequest":
				fmt.Println(awsErr.Error())
				msg := fmt.Sprintf("Bucket %s is invalid", bucket)
				return false, errors.New(msg)
			case "Forbidden":
				fmt.Println(awsErr.Error())
				msg := fmt.Sprintf("Bucket %s is forbidden", bucket)
				return false, errors.New(msg)
			}
		} else {
			return false, err
		}
	}

	return true, nil
}
