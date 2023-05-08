package pipes

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cheeyeo/AWS_S3_PIPES/files"
	"github.com/cheeyeo/AWS_S3_PIPES/reader"
	"github.com/cheeyeo/AWS_S3_PIPES/s3helpers"
	"github.com/schollz/progressbar/v3"
)

type UploadOutput struct {
	Location  string
	VersionID string
	UploadID  string
	ETag      string
}

type UploadInput struct {
	UploadFile string
}

func (up *UploadOutput) Stream(ctx context.Context, pipe string, bucket string, key string) error {
	sess := session.Must(session.NewSession())

	// Check bucket exists and we can access it
	exists, err := s3helpers.BucketValidator(sess, bucket)
	if !exists {
		return err
	}

	pipeFile, err := os.OpenFile(pipe, os.O_RDONLY, 0640)
	if err != nil {
		return err
	}
	defer pipeFile.Close()

	n, err := reader.PipeUpload(ctx, sess, bucket, key, pipeFile)
	if err != nil {
		return err
	}

	up.Location = n.Location
	up.ETag = *n.ETag
	if n.VersionID != nil {
		up.VersionID = *n.VersionID
	}
	if len(n.UploadID) > 0 {
		up.UploadID = n.UploadID
	}

	return nil
}

func (ui *UploadInput) Stream(ctx context.Context, pipe string, bucket string, key string) error {
	if len(ui.UploadFile) > 0 {
		fmt.Printf("Upload from local file %s\n", ui.UploadFile)

		orig, err := os.Open(ui.UploadFile)
		if err != nil {
			return err
		}
		defer orig.Close()

		dst, err := os.OpenFile(pipe, os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer dst.Close()

		fileSize, err := files.GetLocalFileSize(ui.UploadFile)
		if err != nil {
			return err
		}
		uploadMsg := fmt.Sprintf("Uploading %s", key)
		bar := progressbar.DefaultBytes(
			fileSize,
			uploadMsg,
		)

		_, err = io.Copy(io.MultiWriter(dst, bar), orig)
		if err != nil {
			return err
		}
	} else {
		instruct := `
		No file upload target has been specified.
		Write to the pipe manually like so:
	
		cat myfile.txt > %s
		`
		fmt.Printf(instruct, pipe)
		fmt.Println()
	}

	return nil
}
