package reader

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3Reader struct {
	r    io.Reader
	read int64
}

func (r *s3Reader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	r.read += int64(n)

	// TODO: Figure out how to create progressbars here?
	if err == nil {
		// if n > 0 {
		// 	fmt.Printf("Total Read: %d\r", r.read)
		// }
	}
	if err == io.EOF {
		// fmt.Println("Finished Read!")
	}

	return n, err
}

func PipeUpload(ctx context.Context, sess *session.Session, bucket string, key string, pipeFile *os.File) (*s3manager.UploadOutput, error) {
	uploader := s3manager.NewUploader(sess, func(d *s3manager.Uploader) {
		d.Concurrency = 10
		d.PartSize = 25 * 1024 * 1024 // 20MB part size
		d.BufferProvider = s3manager.NewBufferedReadSeekerWriteToPool(25 * 1024 * 1024)
	})

	reader := &s3Reader{r: pipeFile}
	n, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return &s3manager.UploadOutput{}, err
	}

	fmt.Println("S3 upload successful.")
	fmt.Printf("Location: %s\n", n.Location)
	if n.VersionID != nil {
		fmt.Printf("VersionID: %s\n", *n.VersionID)
	}
	if len(n.UploadID) > 0 {
		fmt.Printf("UploadID: %s\n", n.UploadID)
	}
	fmt.Printf("ETag: %s\n", *n.ETag)

	return n, nil
}
