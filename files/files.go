package files

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Get file size of remote s3 object given a bucket and prefix
func GetS3FileSize(sess *session.Session, bucket string, prefix string) (filesize int64, error error) {
	svc := s3.New(sess)
	params := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(prefix),
	}

	resp, err := svc.HeadObject(params)
	if err != nil {
		return 0, err
	}

	return *resp.ContentLength, nil
}

func ByteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func GetLocalFileSize(fname string) (int64, error) {
	f1, err := os.Stat(fname)
	if err != nil {
		return 0, err
	}

	return f1.Size(), nil
}

// Create temp FIFO file
// func CreateTempFIFO()
